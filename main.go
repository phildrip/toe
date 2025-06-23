package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// ParamData represents a parameter or result in a method signature.
type ParamData struct {
	Name string
	Type types.Type // Store the actual types.Type object
}

// MethodData represents a method of an interface.
type MethodData struct {
	Name    string
	Params  []ParamData
	Results []ResultData
}

// ResultData represents a result in a method signature.
type ResultData struct {
	Name string
	Type types.Type // Store the actual types.Type object
}

// InterfaceData represents a parsed interface, including its methods and type parameters.
type InterfaceData struct {
	PackageName string
	Name        string
	Methods     []MethodData
	TypeParams  []ParamData // For generic interfaces, e.g., [T comparable]
	Imports     map[string]string // map[importPath]packageName
}

// StubOptions is now deprecated as locking is a runtime option.
type StubOptions struct{}

func run(stdout, stderr io.Writer, args []string) int {
	var testPackage bool
	var stubDirFlag string
	var outputFile string // Keep outputFile as a flag

	fs := flag.NewFlagSet("toe", flag.ContinueOnError)
	fs.SetOutput(stderr) // Direct flag errors to stderr

	fs.BoolVar(&testPackage, "test-package", false, "generate stub in a _test package")
	fs.StringVar(&stubDirFlag, "stub-dir", "", "generate stub in a specific subdirectory (e.g., \"stubs\") and use its name as package")
	fs.StringVar(&outputFile, "o", "", "output file name (if not provided, defaults to stub_<interface-lowercased>.go in the default/specified stub-dir)")

	// Parse command-line arguments, excluding the program name
	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	// Get inputDir and interfaceName now that flags are parsed
	inputDir := fs.Arg(0)
	interfaceName := fs.Arg(1)

	// Determine the stub directory and filename
	actualStubDir := stubDirFlag
	if actualStubDir == "" {
		actualStubDir = "stubs"
	}

	outputFilename := fmt.Sprintf("stub_%s.go", strings.ToLower(interfaceName))
	// If it's a test package, the filename also changes to include _test suffix
	if testPackage {
		outputFilename = fmt.Sprintf("stub_%s_test.go", strings.ToLower(interfaceName))
	}

	// If -o is not provided, derive it from stubDir and interface name
	if outputFile == "" {
		outputFile = filepath.Join(actualStubDir, outputFilename)
	}

	// Ensure the output directory exists
	// This is handled by main_test.go for tests. For direct CLI usage, it's needed.
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		fmt.Fprintf(stderr, "Error creating output directory for %s: %v\n", outputFile, err)
		return 1
	}

	if fs.NArg() != 2 {
		fmt.Fprintf(stderr,
			"Usage: %s [-test-package] [-stub-dir <dir>] [-o <output.go>] <input_directory> <interface>\n",
			args[0])
		return 1
	}

	var opts = &StubOptions{}

	interfaceData, err := findInterface(inputDir, interfaceName, testPackage, actualStubDir)
	if err != nil {
		fmt.Fprintf(stderr, "Error finding interface: %v\n", err)
		return 1
	}

	stubCode, err := generateStubCode(interfaceData, opts)
	if err != nil {
		fmt.Fprintf(stderr, "Error generating stub: %v\n", err)
		return 1
	}

	// Format the generated code
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, outputFile, []byte(stubCode), parser.ParseComments)
	if err != nil {
		fmt.Fprintf(stderr, "Error parsing generated code for formatting (%s): %v\n", outputFile, err)
		return 1
	}

	var formattedBuf strings.Builder
	err = format.Node(&formattedBuf, fset, node)
	if err != nil {
		fmt.Fprintf(stderr, "Error formatting generated code (%s): %v\n", outputFile, err)
		return 1
	}
	stubCode = formattedBuf.String()

	
		err = os.WriteFile(outputFile, []byte(stubCode), 0644)
		if err != nil {
			fmt.Fprintf(stderr, "Error writing output file: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Stub generated in %s\n", outputFile)
	
	return 0
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args))
}

// collectImports recursively traverses a types.Type and collects external package imports.
func collectImports(data *InterfaceData, t types.Type) {
	switch typ := t.(type) {
	case *types.Named:
		if typ.Obj().Pkg() != nil && typ.Obj().Pkg().Path() != data.PackageName {
			data.Imports[typ.Obj().Pkg().Path()] = typ.Obj().Pkg().Name()
		}
		// Also check underlying type, e.g., for struct fields of named types
		collectImports(data, typ.Underlying())
	case *types.Pointer:
		collectImports(data, typ.Elem())
	case *types.Slice:
		collectImports(data, typ.Elem())
	case *types.Array:
		collectImports(data, typ.Elem())
	case *types.Map:
		collectImports(data, typ.Key())
		collectImports(data, typ.Elem())
	case *types.Chan:
		collectImports(data, typ.Elem())
	case *types.Signature:
		if typ.Params() != nil {
			for i := 0; i < typ.Params().Len(); i++ {
				collectImports(data, typ.Params().At(i).Type())
			}
		}
		if typ.Results() != nil {
			for i := 0; i < typ.Results().Len(); i++ {
				collectImports(data, typ.Results().At(i).Type())
			}
		}
	case *types.Basic:
		// Basic types do not have associated packages.
	case *types.Interface:
		// Interface types themselves might not directly reference packages unless embedded named types.
		// Handled by recursive calls to collectImports on method signatures.
	case *types.TypeParam:
		// Type parameters themselves don't directly reference packages, but their constraints might.
		collectImports(data, typ.Constraint())
	}
}

func findInterface(inputDir string, interfaceName string, testPackage bool, stubDir string) (*InterfaceData, error) {
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
			return nil, fmt.Errorf("found duplicate interface %s in package %s and %s",
				interfaceName,
				foundInterface.PackageName,
				pkg.Name)
		}

		generatedPackageName := pkg.Name
		if stubDir != "" {
			generatedPackageName = filepath.Base(stubDir)
		}
		if testPackage {
			generatedPackageName += "_test"
		}

		data := &InterfaceData{
			PackageName: generatedPackageName,
			Name:        interfaceName,
			Imports:     make(map[string]string),
		}

		// Handle generic interfaces
		if isNamed && namedType.TypeParams() != nil {
			for i := 0; i < namedType.TypeParams().Len(); i++ {
				tp := namedType.TypeParams().At(i)
				data.TypeParams = append(data.TypeParams, ParamData{
					Name: tp.Obj().Name(),
					Type: tp.Constraint(), // Store types.Type directly
				})
				collectImports(data, tp.Constraint())
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
						Type: param.Type(), // Store types.Type directly
					})
					collectImports(data, param.Type())
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
						Type: result.Type(), // Store types.Type directly
					})
					collectImports(data, result.Type())
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

// typeToExpr converts a types.Type to an ast.Expr, handling package imports.
func typeToExpr(t types.Type, currentPackageName string, imports map[string]string) ast.Expr {
	switch typ := t.(type) {
	case *types.Basic:
		return ast.NewIdent(typ.Name())
	case *types.Named:
		// If the named type belongs to an external package, use a selector expression
		if typ.Obj().Pkg() != nil && typ.Obj().Pkg().Path() != currentPackageName {
			// Check if we collected an alias for this package
			pkgName, ok := imports[typ.Obj().Pkg().Path()]
			if !ok {
				pkgName = typ.Obj().Pkg().Name() // Fallback to actual package name
			}
			return &ast.SelectorExpr{
				X:   ast.NewIdent(pkgName),
				Sel: ast.NewIdent(typ.Obj().Name()),
			}
		}
		// Otherwise, it's a type in the current package or a predeclared type
		return ast.NewIdent(typ.Obj().Name())
	case *types.Pointer:
		return &ast.StarExpr{X: typeToExpr(typ.Elem(), currentPackageName, imports)}
	case *types.Slice:
		return &ast.ArrayType{Elt: typeToExpr(typ.Elem(), currentPackageName, imports)}
	case *types.Array:
		return &ast.ArrayType{Len: &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", typ.Len())}, Elt: typeToExpr(typ.Elem(), currentPackageName, imports)}
	case *types.Map:
		return &ast.MapType{Key: typeToExpr(typ.Key(), currentPackageName, imports), Value: typeToExpr(typ.Elem(), currentPackageName, imports)}
	case *types.Chan:
		return &ast.ChanType{Dir: ast.ChanDir(typ.Dir()), Value: typeToExpr(typ.Elem(), currentPackageName, imports)}
	case *types.Signature:
		// This case is for function types (e.g., func(...) (...))
		params := &ast.FieldList{}
		if typ.Params() != nil {
			for i := 0; i < typ.Params().Len(); i++ {
				param := typ.Params().At(i)
				paramNames := []*ast.Ident{}
				if param.Name() != "" {
					paramNames = append(paramNames, ast.NewIdent(param.Name()))
				}
				params.List = append(params.List, &ast.Field{
					Names: paramNames,
					Type:  typeToExpr(param.Type(), currentPackageName, imports),
				})
			}
		}

		results := &ast.FieldList{}
		if typ.Results() != nil {
			for i := 0; i < typ.Results().Len(); i++ {
				result := typ.Results().At(i)
				resultNames := []*ast.Ident{}
				if result.Name() != "" {
					resultNames = append(resultNames, ast.NewIdent(result.Name()))
				}
				results.List = append(results.List, &ast.Field{
					Names: resultNames,
					Type:  typeToExpr(result.Type(), currentPackageName, imports),
				})
			}
		}
		return &ast.FuncType{Params: params, Results: results}
	case *types.Interface:
		// If it's the empty interface, return an empty interface literal
		if typ.Empty() {
			return &ast.InterfaceType{Methods: &ast.FieldList{}} // represents interface{}
		}
		// For other interfaces, fall back to its string representation
		return ast.NewIdent(typ.String())

	case *types.TypeParam:
		return ast.NewIdent(typ.Obj().Name())
	default:
		// Fallback for types not explicitly handled (e.g., complex structs, unexported types)
		return ast.NewIdent(typ.String())
	}
}

func generateStubCode(ifaceData *InterfaceData, opts *StubOptions) (string, error) {
	// Create a new file set and AST file
	fset := token.NewFileSet()
	file := &ast.File{
		Name: ast.NewIdent(ifaceData.PackageName),
	}

	// Add collected imports
	importSpecs := []ast.Spec{ // Always import sync for the mutex
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"sync"`}},
	}
	for path, name := range ifaceData.Imports {
		var importName *ast.Ident
		// Only add name if it's different from the last part of the path or if it's explicitly needed
		lastSlash := strings.LastIndex(path, "/")
		lastPart := path
		if lastSlash != -1 {
			lastPart = path[lastSlash+1:]
		}
		if name != lastPart {
			importName = ast.NewIdent(name)
		}
		importSpecs = append(importSpecs, &ast.ImportSpec{
			Name: importName,
			Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", path)},
		})
	}
	file.Decls = append(file.Decls, &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: importSpecs,
	})

	// Create the stub struct definition
	stubName := "Stub" + ifaceData.Name
	stubStruct := &ast.TypeSpec{
		Name: ast.NewIdent(stubName),
		Type: &ast.StructType{
			Fields: &ast.FieldList{},
		},
	}

	// Add sync.Mutex and _isLocked fields
	stubStruct.Type.(*ast.StructType).Fields.List = append(
		stubStruct.Type.(*ast.StructType).Fields.List, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent("mu")},
			Type:  &ast.SelectorExpr{X: ast.NewIdent("sync"), Sel: ast.NewIdent("Mutex")},
		},
		&ast.Field{
			Names: []*ast.Ident{ast.NewIdent("_isLocked")},
			Type:  ast.NewIdent("bool"),
		},
	)

	// Add fields for call recording and function stubs to the stub struct
	for _, method := range ifaceData.Methods {
		// Add MethodNameFunc field (for lambda stubbing)
		funcType := &ast.FuncType{}
		params := &ast.FieldList{}
		for _, p := range method.Params {
			params.List = append(params.List, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(p.Name)},
				Type:  typeToExpr(p.Type, ifaceData.PackageName, ifaceData.Imports),
			})
		}
		funcType.Params = params

		results := &ast.FieldList{}
		for _, r := range method.Results {
			results.List = append(results.List, &ast.Field{
				Type: typeToExpr(r.Type, ifaceData.PackageName, ifaceData.Imports),
			})
		}
		funcType.Results = results

		stubStruct.Type.(*ast.StructType).Fields.List = append(
			stubStruct.Type.(*ast.StructType).Fields.List, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(method.Name + "Func")},
				Type:  funcType,
			})

		// Add MethodNameCall struct type and its field
		callStructName := stubName + method.Name + "Call"
		callStruct := &ast.TypeSpec{
			Name: ast.NewIdent(callStructName),
			Type: &ast.StructType{
				Fields: &ast.FieldList{},
			},
		}

		// Add type parameters to call struct if the main struct is generic
		if len(ifaceData.TypeParams) > 0 {
			callStruct.TypeParams = &ast.FieldList{List: copyTypeParams(ifaceData.TypeParams, ifaceData.PackageName, ifaceData.Imports)}
		}

		for _, p := range method.Params {
			callStruct.Type.(*ast.StructType).Fields.List = append(
				callStruct.Type.(*ast.StructType).Fields.List, &ast.Field{
					Names: []*ast.Ident{ast.NewIdent(strings.Title(p.Name))},
					Type:  typeToExpr(p.Type, ifaceData.PackageName, ifaceData.Imports),
				})
		}

		file.Decls = append(file.Decls, &ast.GenDecl{
			Tok:   token.TYPE,
			Specs: []ast.Spec{callStruct},
		})

		// Add MethodNameCalls field (slice of callStructName[T])
		var callListType ast.Expr = ast.NewIdent(callStructName)
		if len(ifaceData.TypeParams) > 0 {
			var typeArgs []ast.Expr
			for _, tp := range ifaceData.TypeParams {
				typeArgs = append(typeArgs, ast.NewIdent(tp.Name))
			}
			if len(typeArgs) == 1 {
				callListType = &ast.IndexExpr{X: ast.NewIdent(callStructName), Index: typeArgs[0]}
			} else {
				callListType = &ast.IndexListExpr{X: ast.NewIdent(callStructName), Indices: typeArgs}
			}
		}

		stubStruct.Type.(*ast.StructType).Fields.List = append(
			stubStruct.Type.(*ast.StructType).Fields.List, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(method.Name + "Calls")},
				Type:  &ast.ArrayType{Elt: callListType},
			})

		// Add fields for fixed return values
		for i, res := range method.Results {
			stubStruct.Type.(*ast.StructType).Fields.List = append(
				stubStruct.Type.(*ast.StructType).Fields.List, &ast.Field{
					Names: []*ast.Ident{ast.NewIdent(fmt.Sprintf("%sReturns%d", method.Name, i))},
					Type:  typeToExpr(res.Type, ifaceData.PackageName, ifaceData.Imports),
				})
		}
	}

	// Add type parameters for generic interfaces
	if len(ifaceData.TypeParams) > 0 {
		fields := make([]*ast.Field, len(ifaceData.TypeParams))
		for i, tp := range ifaceData.TypeParams {
			// Determine the AST representation of the type parameter's constraint
			var constraintType ast.Expr
			if tp.Type.String() == "interface{}" { // Now using .String() on types.Type
				constraintType = &ast.InterfaceType{Methods: &ast.FieldList{}} // Represents 'interface{}'
			} else {
				constraintType = typeToExpr(tp.Type, ifaceData.PackageName, ifaceData.Imports)
			}

			fields[i] = &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(tp.Name)},
				Type:  constraintType,
			}
		}
		stubStruct.TypeParams = &ast.FieldList{List: fields}
	}

	file.Decls = append(file.Decls, &ast.GenDecl{ // Changed from decls = append(decls, ...)
		Tok:   token.TYPE,
		Specs: []ast.Spec{stubStruct},
	})

	// Create constructor
	file.Decls = append(file.Decls, createConstructor(stubName, ifaceData.TypeParams, ifaceData.PackageName, ifaceData.Imports))

	// Create methods for the stub struct
	for _, method := range ifaceData.Methods {
		file.Decls = append(file.Decls, createMethod(stubName, method, ifaceData.TypeParams, ifaceData.PackageName, ifaceData.Imports))
	}

	// Generate the code
	var buf strings.Builder
	if err := format.Node(&buf, fset, file); err != nil {
		return "", fmt.Errorf("error formatting generated code: %v", err)
	}

	return buf.String(), nil
}

// copyTypeParams creates a new slice of ast.Field representing type parameters.
// It's used to safely copy type parameters for nested generic structs.
func copyTypeParams(params []ParamData, currentPackageName string, imports map[string]string) []*ast.Field {
	copied := make([]*ast.Field, len(params))
	for i, p := range params {
		var constraintType ast.Expr
		if p.Type.String() == "interface{}" {
			constraintType = &ast.InterfaceType{Methods: &ast.FieldList{}}
		} else {
			constraintType = typeToExpr(p.Type, currentPackageName, imports) // Use typeToExpr here
		}
		copied[i] = &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(p.Name)},
			Type:  constraintType,
		}
	}
	return copied
}

func createConstructor(stubName string, typeParams []ParamData, currentPackageName string, imports map[string]string) *ast.FuncDecl {
	constructorName := "New" + stubName
	withLockingParam := &ast.Field{Names: []*ast.Ident{ast.NewIdent("withLocking")}, Type: ast.NewIdent("bool")}

	// Build receiver type for the constructor
	var resultType ast.Expr = ast.NewIdent(stubName)

	// Type parameters for the constructor function itself
	var funcTypeParams *ast.FieldList
	if len(typeParams) > 0 {
		funcTypeParams = &ast.FieldList{List: copyTypeParams(typeParams, currentPackageName, imports)}

		var typeArgs []ast.Expr
		for _, tp := range typeParams {
			typeArgs = append(typeArgs, ast.NewIdent(tp.Name))
		}
		if len(typeArgs) == 1 {
			resultType = &ast.IndexExpr{X: ast.NewIdent(stubName), Index: typeArgs[0]}
		} else {
			resultType = &ast.IndexListExpr{X: ast.NewIdent(stubName), Indices: typeArgs}
		}
	}

	return &ast.FuncDecl{
		Name: ast.NewIdent(constructorName),
		Type: &ast.FuncType{
			TypeParams: funcTypeParams, // Add type parameters to the function declaration
			Params:     &ast.FieldList{List: []*ast.Field{withLockingParam}},
			Results:    &ast.FieldList{List: []*ast.Field{{Type: &ast.StarExpr{X: resultType}}}},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.UnaryExpr{
							Op: token.AND,
							X: &ast.CompositeLit{
								Type: resultType,
								Elts: []ast.Expr{
									&ast.KeyValueExpr{Key: ast.NewIdent("_isLocked"), Value: ast.NewIdent("withLocking")},
								},
							},
						},
					},
				},
			},
		},
	}
}

func createMethod(stubName string, method MethodData, typeParams []ParamData, currentPackageName string, imports map[string]string) *ast.FuncDecl {
	// Method receiver
	recv := &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent("s")},
				Type:  &ast.StarExpr{X: ast.NewIdent(stubName)},
			},
		},
	}

	// If the struct is generic, build the full receiver type e.g., *StubName[T]
	if len(typeParams) > 0 {
		var typeArgs []ast.Expr
		for _, tp := range typeParams {
			typeArgs = append(typeArgs, ast.NewIdent(tp.Name))
		}

		if len(typeArgs) == 1 {
			recv.List[0].Type = &ast.StarExpr{
				X: &ast.IndexExpr{
					X:     ast.NewIdent(stubName),
					Index: typeArgs[0],
				},
			}
		} else {
			recv.List[0].Type = &ast.StarExpr{
				X: &ast.IndexListExpr{
					X:       ast.NewIdent(stubName),
					Indices: typeArgs,
				},
			}
		}
	}

	// Method parameters
	params := &ast.FieldList{}
	for _, p := range method.Params {
		params.List = append(params.List, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(p.Name)},
			Type:  typeToExpr(p.Type, currentPackageName, imports),
		})
	}

	// Method results
	results := &ast.FieldList{}
	for _, r := range method.Results {
		results.List = append(results.List, &ast.Field{
			Type: typeToExpr(r.Type, currentPackageName, imports),
		})
	}

	// Method body
	var bodyStmts []ast.Stmt

	// Add conditional locking
	bodyStmts = append(bodyStmts, &ast.IfStmt{
		Cond: ast.NewIdent("s._isLocked"),
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.SelectorExpr{X: ast.NewIdent("s"), Sel: ast.NewIdent("mu")}, Sel: ast.NewIdent("Lock")},
					},
				},
				&ast.DeferStmt{
					Call: &ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.SelectorExpr{X: ast.NewIdent("s"), Sel: ast.NewIdent("mu")}, Sel: ast.NewIdent("Unlock")},
					},
				},
			},
		},
	})

	// Add call recording
	callStructName := stubName + method.Name + "Call"
	var callElts []ast.Expr
	for _, p := range method.Params {
		callElts = append(callElts, &ast.KeyValueExpr{
			Key:   ast.NewIdent(strings.Title(p.Name)), // Capitalize key for public field
			Value: ast.NewIdent(p.Name),
		})
	}

	// The type of the call instance needs to be generic if the interface is generic
	var callInstanceType ast.Expr = ast.NewIdent(callStructName)
	if len(typeParams) > 0 {
		var typeArgs []ast.Expr
		for _, tp := range typeParams {
			typeArgs = append(typeArgs, ast.NewIdent(tp.Name))
		}
		if len(typeArgs) == 1 {
			callInstanceType = &ast.IndexExpr{
				X:     ast.NewIdent(callStructName),
				Index: typeArgs[0],
			}
		} else {
			callInstanceType = &ast.IndexListExpr{
				X:       ast.NewIdent(callStructName),
				Indices: typeArgs,
			}
		}
	}

	callInstance := &ast.CompositeLit{
		Type: callInstanceType,
		Elts: callElts,
	}

	// Assign the result of append back to the slice
	bodyStmts = append(bodyStmts, &ast.AssignStmt{
		Tok: token.ASSIGN,
		Lhs: []ast.Expr{&ast.SelectorExpr{X: ast.NewIdent("s"), Sel: ast.NewIdent(method.Name + "Calls")}},
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun: ast.NewIdent("append"),
			Args: []ast.Expr{
				&ast.SelectorExpr{X: ast.NewIdent("s"), Sel: ast.NewIdent(method.Name + "Calls")},
				callInstance,
			},
		}},
	})


	// Add logic for MethodNameFunc (lambda stubbing) or fixed return values
	var returnArgs []ast.Expr
	for i := range method.Results {
		returnArgs = append(returnArgs, &ast.SelectorExpr{
			X:   ast.NewIdent("s"),
			Sel: ast.NewIdent(fmt.Sprintf("%sReturns%d", method.Name, i)),
		})
	}

	if len(method.Results) > 0 { // Only if the method has return values
		// if s.MethodNameFunc != nil {
		ifStmt := &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  &ast.SelectorExpr{X: ast.NewIdent("s"), Sel: ast.NewIdent(method.Name + "Func")},
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					// return s.MethodNameFunc(params...)
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{X: ast.NewIdent("s"), Sel: ast.NewIdent(method.Name + "Func")},
								Args: func() []ast.Expr {
									var args []ast.Expr
									for _, p := range method.Params {
										args = append(args, ast.NewIdent(p.Name)) // Use p.Name for the argument
									}
									return args
								}(),
							},
						},
					},
				},
			},
			Else: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{Results: returnArgs},
				},
			},
		}
		bodyStmts = append(bodyStmts, ifStmt)
	} else { // No return values, just add a simple return
		bodyStmts = append(bodyStmts, &ast.ReturnStmt{})
	}

	tbody := &ast.BlockStmt{
		List: bodyStmts,
	}

	return &ast.FuncDecl{
		Recv: recv,
		Name: ast.NewIdent(method.Name),
		Type: &ast.FuncType{
			Params:  params,
			Results: results,
		},
		Body: tbody,
	}
}
