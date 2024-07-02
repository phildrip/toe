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
	"text/template"

	"golang.org/x/tools/go/packages"
)

func main() {
	var outputFile string
	flag.StringVar(&outputFile, "o", "", "output file name")
	flag.Parse()

	fmt.Println("NARG:", flag.NArg())
	fmt.Println("ARGS:", flag.Args())
	fmt.Println("outputFile:", outputFile)
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s  -o <output.go> <input_directory> <InterfaceName>\n", os.Args[0])
		os.Exit(1)
	}

	inputDir := flag.Arg(0)
	interfaceName := flag.Arg(1)

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
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	var interfaceMethods []*ast.Field
	var packageName string

	for _, pkg := range pkgs {
		packageName = pkg.Name
		fmt.Println("Package:", pkg.Name)
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == interfaceName {
					if ift, ok := ts.Type.(*ast.InterfaceType); ok {
						interfaceMethods = ift.Methods.List
					}
				}
				return true
			})
		}
	}

	if len(interfaceMethods) == 0 {
		fmt.Fprintf(os.Stderr, "Interface %s not found\n", interfaceName)
		os.Exit(1)
	}

	stubCode := generateStubCode(interfaceName, interfaceMethods, packageName)

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

func generateStubCode(interfaceName string, methods []*ast.Field, packageName string) string {
	stubName := "Stub" + interfaceName

	tmpl := template.Must(template.New("stub").Parse(`
package {{.PackageName}}

import (
    "sync"
)

type {{.StubName}}Call struct {
    {{range .Methods}}
    {{.Name}}Calls []struct {
        {{.Params}}
    }
    {{end}}
}

type {{.StubName}} struct {
    mu sync.Mutex
    calls {{.StubName}}Call
    {{range .Methods}}
    {{.Name}}Func func({{.Params}}) {{.Results}}
    {{end}}
}

func (s *{{.StubName}}) On() *{{.StubName}} {
    return s
}

{{range .Methods}}
func (s *{{$.StubName}}) {{.Name}}({{.Params}}) {{.Results}} {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.calls.{{.Name}}Calls = append(s.calls.{{.Name}}Calls, struct{ {{.Params}} }{ {{.ParamNames}} })
    if s.{{.Name}}Func != nil {
        return s.{{.Name}}Func({{.ParamNames}})
    }
    {{if .Results}}
    var zero {{.Results}}
    return zero
    {{end}}
}

func (s *{{$.StubName}}) {{.Name}}ThenReturn({{.Results}}) *{{$.StubName}} {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.{{.Name}}Func = func({{.Params}}) {{.Results}} {
        return {{.ResultNames}}
    }
    return s
}

func (s *{{$.StubName}}) {{.Name}}Calls() []struct{ {{.Params}} } {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.calls.{{.Name}}Calls
}
{{end}}
`))

	var methodsData []struct {
		Name        string
		Params      string
		ParamNames  string
		Results     string
		ResultNames string
	}

	for _, method := range methods {
		if len(method.Names) == 0 {
			continue
		}
		methodName := method.Names[0].Name
		funcType := method.Type.(*ast.FuncType)

		params := getFieldList(funcType.Params)
		paramNames := getFieldNames(funcType.Params)
		results := getFieldList(funcType.Results)
		resultNames := getFieldNames(funcType.Results)

		methodsData = append(methodsData, struct {
			Name        string
			Params      string
			ParamNames  string
			Results     string
			ResultNames string
		}{
			Name:        methodName,
			Params:      params,
			ParamNames:  paramNames,
			Results:     results,
			ResultNames: resultNames,
		})
	}

	var buf strings.Builder
	err := tmpl.Execute(&buf, struct {
		PackageName   string
		InterfaceName string
		StubName      string
		Methods       []struct {
			Name        string
			Params      string
			ParamNames  string
			Results     string
			ResultNames string
		}
	}{
		PackageName:   packageName,
		InterfaceName: interfaceName,
		StubName:      stubName,
		Methods:       methodsData,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating stub: %v\n", err)
		os.Exit(1)
	}

	// Format the generated code
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", buf.String(), parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing generated code: %v\n", err)
		fmt.Println(buf.String())
		os.Exit(1)
	}

	var formattedBuf strings.Builder
	err = format.Node(&formattedBuf, fset, node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting generated code: %v\n", err)
		os.Exit(1)
	}

	return formattedBuf.String()
}

func getFieldList(fields *ast.FieldList) string {
	if fields == nil {
		return ""
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
	return strings.Join(params, ", ")
}

func getFieldNames(fields *ast.FieldList) string {
	if fields == nil {
		return ""
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
	return strings.Join(names, ", ")
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
		return "func(" + getFieldList(t.Params) + ") " + getFieldList(t.Results)
	default:
		return fmt.Sprintf("%T", expr)
	}
}
