package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"go/types"

	"golang.org/x/tools/go/packages"
)

	var outputFile string
	var withLocking bool

	flag.BoolVar(&withLocking, "with-locking", false, "include sync.Mutex for concurrency safety")
	flag.StringVar(&outputFile, "o", "", "output file name")
	flag.Parse()

	var opts = &StubOptions{
		WithLocking: withLocking,
	}

	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr,
			"Usage: %s [-with-locking] -o <output.go> <input_directory> <interface>\n",
			os.Args[0])

		os.Exit(1)
	}

	inputDir := flag.Arg(0)
	interfaceName := flag.Arg(1)

	interfaceData, err := findInterface(inputDir, interfaceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding interface: %v\n", err)
		os.Exit(1)
	}

	stubCode, err := generateStubCode(interfaceData, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating stub: %v\n", err)
		os.Exit(1)
	}

	if outputFile == "" {
		fmt.Println(stubCode)
	} else {
		err := os.WriteFile(outputFile, []byte(stubCode), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Stub generated in %s\n", outputFile)
	}
}

func findInterface(inputDir string, interfaceName string) (*InterfaceData, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
		Dir: inputDir,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("load: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("packages contain errors")
	}

	var foundInterface *InterfaceData

	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()
		obj := scope.Lookup(interfaceName)
		if obj == nil {
			continue
		}

		ifaceType, ok := obj.Type().Underlying().(*types.Interface)
		if !ok {
			continue // Not an interface
		}

		namedType, isNamed := obj.Type().(*types.Named)

		if foundInterface != nil {
			return nil, fmt.Errorf("found duplicate interface %s in package %s and %s", interfaceName, foundInterface.PackageName, pkg.Name)
		}

		data := &InterfaceData{
			PackageName: pkg.Name,
			Name:        interfaceName,
		}

		// Handle generic interfaces
		if isNamed && namedType.TypeParams() != nil {
			for i := 0; i < namedType.TypeParams().Len(); i++ {
				tp := namedType.TypeParams().At(i)
				data.TypeParams = append(data.TypeParams, ParamData{
					Name: tp.Obj().Name(),
					Type: tp.String(), // e.g., "T comparable"
				})
			}
		}

		for i := 0; i < ifaceType.NumMethods(); i++ {
			method := ifaceType.Method(i)
			sig := method.Type().(*types.Signature)

			methodData := MethodData{
				Name: method.Name(),
			}

			// Parameters
			if sig.Params() != nil {
				for j := 0; j < sig.Params().Len(); j++ {
					param := sig.Params().At(j)
					methodData.Params = append(methodData.Params, ParamData{
						Name: param.Name(),
						Type: param.Type().String(),
					})
				}
			}

			// Results
			if sig.Results() != nil {
				for j := 0; j < sig.Results().Len(); j++ {
					result := sig.Results().At(j)
					name := result.Name()
					if name == "" {
						name = fmt.Sprintf("R%d", j) // Default name for unnamed results
					}
					methodData.Results = append(methodData.Results, ResultData{
						Name: name,
						Type: result.Type().String(),
					})
				}
			}
			data.Methods = append(data.Methods, methodData)
		}
		foundInterface = data
	}

	if foundInterface == nil {
		return nil, fmt.Errorf("interface %s not found", interfaceName)
	}

	return foundInterface, nil
}

// InterfaceData and MethodData define the intermediate representation
// of the parsed interface, including type information.
type ParamData struct {
	Name string
	Type string // String representation of the type, using type.String()
}

type ResultData struct {
	Name string
	Type string // String representation of the type, using type.String()
}

type MethodData struct {
	Name    string
	Params  []ParamData
	Results []ResultData
}

type InterfaceData struct {
	PackageName string
	Name        string
	Methods     []MethodData
	TypeParams  []ParamData // For generic interfaces, e.g., [T comparable]
}

// StubOptions allows configuring the generated stub behavior.
type StubOptions struct {
	WithLocking bool // If true, the stub will include a sync.Mutex for concurrency safety.
}


func generateStubCode(ifaceData *InterfaceData, opts *StubOptions) (string, error) {
	// Create a new file set and AST file
	fset := token.NewFileSet()
	file := &ast.File{
		Name: ast.NewIdent(ifaceData.PackageName),
	}

	// Add sync import if locking is enabled
	if opts.WithLocking {
		file.Decls = append(file.Decls, &ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"sync"`}},
			},
		})
	}

	// Create the stub struct definition
	stubName := "Stub" + ifaceData.Name
	stubStruct := &ast.TypeSpec{
		Name: ast.NewIdent(stubName),
		Type: &ast.StructType{
			Fields: &ast.FieldList{},
		},
	}

	// Add sync.Mutex field if locking is enabled
	if opts.WithLocking {
		stubStruct.Type.(*ast.StructType).Fields.List = append(
			stubStruct.Type.(*ast.StructType).Fields.List, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent("mu")},
				Type:  &ast.SelectorExpr{X: ast.NewIdent("sync"), Sel: ast.NewIdent("Mutex")},
			})
	}

	// Add type parameters for generic interfaces
	if len(ifaceData.TypeParams) > 0 {
		fields := make([]*ast.Field, len(ifaceData.TypeParams))
		for i, tp := range ifaceData.TypeParams {
			fields[i] = &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(tp.Name)},
				Type:  ast.NewIdent(strings.Split(tp.Type, " ")[1]), // e.g., "comparable"
			}
		}
		stubStruct.TypeParams = &ast.FieldList{List: fields}
	}

	decls := []ast.Decl{
		&ast.GenDecl{
			Tok:   token.TYPE,
			Specs: []ast.Spec{stubStruct},
		},
	}

	// Create methods for the stub struct
	for _, method := range ifaceData.Methods {
		decls = append(decls, createMethod(stubName, method, ifaceData.TypeParams, opts))
	}

	file.Decls = decls

	// Generate the code
	var buf strings.Builder
	if err := format.Node(&buf, fset, file); err != nil {
		return "", fmt.Errorf("error formatting generated code: %v", err)
	}

	return buf.String(), nil
}

func createMethod(stubName string, method MethodData, typeParams []ParamData, opts *StubOptions) *ast.FuncDecl {
	// Method receiver
	recv := &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent("s")},
				Type:  &ast.StarExpr{X: ast.NewIdent(stubName)},
			},
		},
	}

	// If the struct is generic, add type params to receiver
	if len(typeParams) > 0 {
		typeArgs := make([]ast.Expr, len(typeParams))
		for i, tp := range typeParams {
			typeArgs[i] = ast.NewIdent(tp.Name)
		}
		recv.List[0].Type = &ast.IndexListExpr{
			X:       &ast.StarExpr{X: ast.NewIdent(stubName)},
			Indices: typeArgs,
		}
	}

	// Method parameters
	params := &ast.FieldList{}
	for _, p := range method.Params {
		params.List = append(params.List, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(p.Name)},
			Type:  ast.NewIdent(p.Type),
		})
	}

	// Method results
	results := &ast.FieldList{}
	for _, r := range method.Results {
		results.List = append(results.List, &ast.Field{
			Type: ast.NewIdent(r.Type),
		})
	}

	// Method body
	var bodyStmts []ast.Stmt

	// Add locking if enabled
	if opts.WithLocking {
		bodyStmts = append(bodyStmts, &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{X: ast.NewIdent("s"), Sel: ast.NewIdent("mu").Sel},
				Args: nil,
			},
		})
		bodyStmts = append(bodyStmts, &ast.DeferStmt{
			Call: &ast.CallExpr{
				Fun: &ast.SelectorExpr{X: ast.NewIdent("s"), Sel: ast.NewIdent("mu").Sel},
				Args: nil,
			},
		})
	}

	// Add return statement (for now, just zero values)
	bodyStmts = append(bodyStmts, &ast.ReturnStmt{})

	body := &ast.BlockStmt{
		List: bodyStmts,
	}

	return &ast.FuncDecl{
		Recv: recv,
		Name: ast.NewIdent(method.Name),
		Type: &ast.FuncType{
			Params:  params,
			Results: results,
		},
		Body: body,
	}
}



