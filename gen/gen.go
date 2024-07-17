package gen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/packages"
	"strings"
	"text/template"
)

//go:embed stub.go.tmpl
var stubTemplate string

// FindInterface finds the interface methods and package name for the given
// interface name in the input directory.
// It loads the Go packages in the input directory, searches for the interface
// definition, and returns the list of interface methods
// and the package name containing the interface.
// If the interface is not found or there are errors loading the packages, it
// returns an error.
func FindInterface(inputDir string, interfaceName string) ([]*ast.Field, string, error) {
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
		return nil, "", fmt.Errorf("load: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, "", fmt.Errorf("packages contain errors")
	}

	var interfaceMethods []*ast.Field
	var packageName string

	for _, pkg := range pkgs {
		packageName = pkg.Name
		for _, file := range pkg.Syntax {
			ast.Inspect(
				file, func(n ast.Node) bool {
					if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == interfaceName {
						if ift, ok := ts.Type.(*ast.InterfaceType); ok {
							interfaceMethods = ift.Methods.List
						}
					}
					return true
				})
		}
	}
	return interfaceMethods, packageName, nil
}

type methodData struct {
	Name        string
	Params      []string
	ParamNames  []string
	Results     []string
	ResultNames []string
}

// GenerateStubCode generates a stub implementation of an interface based on the
// provided interface name, methods, and package name.
// The generated stub code is returned as a string, and can be optionally formatted
// if disableFormatting is false.
// The stub implementation is named "Stub" followed by the interface name.
// For each method in the interface, the stub implementation includes parameter and
// return value declarations matching the interface method.
func GenerateStubCode(
	interfaceName string,
	methods []*ast.Field,
	packageName string,
	disableFormatting bool) (string, error) {

	stubName := "Stub" + interfaceName

	funcMap := template.FuncMap{
		"join":  strings.Join,
		"zip":   zip,
		"joinl": joinl,
	}

	tmpl := template.Must(
		template.New("stub").
			Funcs(funcMap).
			Parse(stubTemplate))

	var methodsData []methodData

	for _, method := range methods {
		if len(method.Names) == 0 {
			continue
		}
		methodName := method.Names[0].Name
		funcType := method.Type.(*ast.FuncType)

		params := getFieldList(funcType.Params)
		paramNames := getFieldNames(funcType.Params)
		results := getFieldList(funcType.Results)
		resultNames := getResultNames(funcType.Results)

		methodsData = append(
			methodsData, methodData{
				Name:        methodName,
				Params:      params,
				ParamNames:  paramNames,
				Results:     results,
				ResultNames: resultNames,
			})
	}

	var buf strings.Builder

	//fmt.Println(prettyPrint(methodsData))

	err := tmpl.Execute(
		&buf, struct {
			PackageName   string
			InterfaceName string
			StubName      string
			Methods       []methodData
		}{
			PackageName:   packageName,
			InterfaceName: interfaceName,
			StubName:      stubName,
			Methods:       methodsData,
		})

	if err != nil {
		return "", fmt.Errorf("error generating stub: %v", err)
	}

	if !disableFormatting {
		// Format the generated code
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, "", buf.String(), parser.ParseComments)
		if err != nil {
			return "", fmt.Errorf("error parsing generated code: %v", err)
		}

		var formattedBuf strings.Builder
		err = format.Node(&formattedBuf, fset, node)
		if err != nil {
			return "", fmt.Errorf("error formatting generated code: %v", err)
		}

		return formattedBuf.String(), nil
	} else {
		return buf.String(), nil
	}
}

func getFieldList(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var params []string
	for _, field := range fields.List {
		paramType := getTypeString(field.Type)
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				params = append(params, fmt.Sprintf("%s %s", name.Name, paramType))
			}
		} else {
			params = append(params, paramType)
		}
	}
	return params
}

func getFieldNames(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var names []string
	for _, field := range fields.List {
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
		} else {
			names = append(names, "_")
		}
	}
	return names
}

func getResultNames(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var names []string
	for i, field := range fields.List {
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
		} else {
			names = append(names, fmt.Sprintf("R%d", i))
		}
	}
	return names
}

func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", getTypeString(t.X), t.Sel.Name)
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + getTypeString(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", getTypeString(t.Key), getTypeString(t.Value))
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func(" + strings.Join(
			getFieldList(t.Params), ", "+
				"") + ") " + strings.Join(getFieldList(t.Results), ", ")
	default:
		return fmt.Sprintf("%T", expr)
	}
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
