package gen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindInterface(t *testing.T) {
	for _, tt := range []struct {
		name            string
		inputDir        string
		interfaceName   string
		wantMethods     []*ast.Field
		wantPackageName string
		wantErr         bool
	}{
		{
			name:            "interface not found",
			inputDir:        "testdata/no_interface",
			interfaceName:   "Interface",
			wantMethods:     nil,
			wantPackageName: "",
			wantErr:         true,
		},
		{
			name:          "interface found",
			inputDir:      "testdata",
			interfaceName: "Interface",
			wantMethods: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "Method1"}},
					Type:  &ast.FuncType{},
				},
				{
					Names: []*ast.Ident{{Name: "Method2"}},
					Type:  &ast.FuncType{Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "error"}}}}},
				},
			},
			wantPackageName: "testdata",
			wantErr:         false,
		},
		{
			name:            "empty interface",
			inputDir:        "testdata",
			interfaceName:   "EmptyInterface",
			wantMethods:     []*ast.Field{},
			wantPackageName: "testdata",
			wantErr:         false,
		},
		{
			name:          "interface with param",
			inputDir:      "testdata",
			interfaceName: "InterfaceWithParam",
			wantMethods: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "Method1"}},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{{Name: "arg1"}},
									Type:  &ast.Ident{Name: "int"},
								},
							},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{{Type: &ast.Ident{Name: "error"}}},
						},
					},
				},
			},
			wantPackageName: "testdata",
			wantErr:         false,
		},
	} {
		t.Run(
			tt.name, func(t *testing.T) {
				gotMethods, gotPackageName, err := FindInterface(tt.inputDir, tt.interfaceName)

				if (err != nil) != tt.wantErr {
					t.Errorf("FindInterface() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if gotPackageName != tt.wantPackageName {
					t.Errorf("FindInterface() gotPackageName = %v, want %v",
						gotPackageName,
						tt.wantPackageName)
				}

				if len(gotMethods) != len(tt.wantMethods) {
					t.Errorf("FindInterface() gotMethods length = %v, want %v",
						len(gotMethods),
						len(tt.wantMethods))
					return
				}

				for i, gotMethod := range gotMethods {
					wantMethod := tt.wantMethods[i]
					if gotMethod.Names[0].Name != wantMethod.Names[0].Name {
						t.Errorf(
							"FindInterface() gotMethod name = %v, want %v",
							gotMethod.Names[0].Name,
							wantMethod.Names[0].Name)
					}
				}
			})
	}
}

func TestGenerateStubCode(t *testing.T) {
	tests := []struct {
		name               string
		interfaceName      string
		methods            []*ast.Field
		packageName        string
		disableFormatting  bool
		expectedSubstrings []string
	}{
		{
			name:          "Simple interface",
			interfaceName: "Interface",
			methods: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "Method1"}},
					Type:  &ast.FuncType{},
				},
				{
					Names: []*ast.Ident{{Name: "Method2"}},
					Type:  &ast.FuncType{Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "error"}}}}},
				},
			},
			packageName:       "testdata",
			disableFormatting: false,
			expectedSubstrings: []string{
				"package testdata",
				"type StubInterface struct",
				"func (s *StubInterface) Method1()",
				"func (s *StubInterface) Method2() (error)",
			},
		},
		{
			name:          "Interface with parameters",
			interfaceName: "InterfaceWithParam",
			methods: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "Method1"}},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{{Name: "arg1"}},
									Type:  &ast.Ident{Name: "int"},
								},
							},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{{Type: &ast.Ident{Name: "error"}}},
						},
					},
				},
			},
			packageName:       "testdata",
			disableFormatting: false,
			expectedSubstrings: []string{
				"package testdata",
				"type StubInterfaceWithParam struct",
				"func (s *StubInterfaceWithParam) Method1(arg1 int) (error)",
			},
		},
		{
			name:          "Complex interface with multiple parameters and return values",
			interfaceName: "ComplexInterface",
			methods: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "ComplexMethod"}},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{{Name: "arg1"}},
									Type:  &ast.Ident{Name: "string"},
								},
								{
									Names: []*ast.Ident{{Name: "arg2"}},
									Type:  &ast.Ident{Name: "int"},
								},
							},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "string"}},
								{Type: &ast.Ident{Name: "error"}},
							},
						},
					},
				},
			},
			packageName:       "testdata",
			disableFormatting: false,
			expectedSubstrings: []string{
				"package testdata",
				"type StubComplexInterface struct",
				"func (s *StubComplexInterface) ComplexMethod(arg1 string, arg2 int) (string, error)",
			},
		},
		{
			name:          "Test with disabled formatting",
			interfaceName: "Interface",
			methods: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "Method1"}},
					Type:  &ast.FuncType{},
				},
			},
			packageName:       "testdata",
			disableFormatting: true,
			expectedSubstrings: []string{
				"package testdata",
				"type StubInterface struct",
				"func (s *StubInterface) Method1()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				stubCode, err := GenerateStubCode(tt.interfaceName, tt.methods, tt.packageName)
				if err != nil {
					t.Fatalf("GenerateStubCode() error = %v, stub code: \n%v", err, stubCode)
				}

				for _, expectedSubstring := range tt.expectedSubstrings {
					if !strings.Contains(stubCode, expectedSubstring) {
						t.Errorf(
							"GenerateStubCode() generated code does not contain expected substring: %s",
							expectedSubstring)
					}
				}
			})
	}
}

func TestGetTypeString(t *testing.T) {
	tests := []struct {
		name     string
		typeExpr string
		want     string
	}{
		{
			name:     "Simple type",
			typeExpr: "int",
			want:     "int",
		},
		{
			name:     "Pointer type",
			typeExpr: "*string",
			want:     "*string",
		},
		{
			name:     "Slice type",
			typeExpr: "[]byte",
			want:     "[]byte",
		},
		{
			name:     "Map type",
			typeExpr: "map[string]int",
			want:     "map[string]int",
		},
		{
			name:     "Interface type",
			typeExpr: "interface{}",
			want:     "interface{}",
		},
		{
			name:     "Qualified type",
			typeExpr: "time.Time",
			want:     "time.Time",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Parse the type expression
				expr, err := parser.ParseExpr(tt.typeExpr)
				if err != nil {
					t.Fatalf("Failed to parse type expression: %v", err)
				}

				got := getTypeString(expr)
				if got != tt.want {
					t.Errorf("getTypeString() = %v, want %v", got, tt.want)
				}
			})
	}
}

func TestTemplateFunctions(t *testing.T) {
	t.Run(
		"zip function", func(t *testing.T) {
			a := []string{"a", "b", "c"}
			b := []string{"1", "2", "3"}
			format := "%s:%s"

			result := zip(a, b, format)

			expected := []string{"a:1", "b:2", "c:3"}
			if len(result) != len(expected) {
				t.Fatalf("zip() returned slice of length %d, want %d", len(result), len(expected))
			}

			for i, v := range result {
				if v != expected[i] {
					t.Errorf("zip()[%d] = %s, want %s", i, v, expected[i])
				}
			}
		})

	t.Run(
		"joinl function", func(t *testing.T) {
			a := []string{"a", "b", "c"}
			sep := ", "

			result := joinl(sep, a)

			expected := "a, b, c"
			if result != expected {
				t.Errorf("joinl() = %s, want %s", result, expected)
			}
		})
}

// TestEndToEnd tests the full process of finding an interface and generating a stub
func TestEndToEnd(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "toe-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file with an interface
	testFile := filepath.Join(tempDir, "test.go")
	testCode := `package testpkg

type TestInterface interface {
	DoSomething(input string) (string, error)
	DoSomethingElse(count int) error
}
`
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Find the interface
	methods, pkgName, err := FindInterface(tempDir, "TestInterface")
	if err != nil {
		t.Fatalf("FindInterface() error = %v", err)
	}

	if pkgName != "testpkg" {
		t.Errorf("FindInterface() package name = %v, want testpkg", pkgName)
	}

	if len(methods) != 2 {
		t.Fatalf("FindInterface() found %d methods, want 2", len(methods))
	}

	// Generate stub code
	stubCode, err := GenerateStubCode("TestInterface", methods, pkgName, false)
	if err != nil {
		t.Fatalf("GenerateStubCode() error = %v", err)
	}

	// Verify the generated code
	expectedSubstrings := []string{
		"package testpkg",
		"type StubTestInterface struct",
		"func (s *StubTestInterface) DoSomething(input string) (string, error)",
		"func (s *StubTestInterface) DoSomethingElse(count int) error",
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(stubCode, substr) {
			t.Errorf("Generated code does not contain expected substring: %s", substr)
		}
	}

	// Verify the generated code is valid Go code
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "", stubCode, parser.ParseComments)
	if err != nil {
		t.Errorf("Generated code is not valid Go: %v\nCode:\n%s", err, stubCode)
	}
}

// TestGetTypeStringWithAst tests getTypeString with manually constructed AST nodes
func TestGetTypeStringWithAst(t *testing.T) {
	// Create a file set
	fset := token.NewFileSet()

	// Test with a complex type that might be hard to construct with parser.ParseExpr
	src := `
package test

type T struct{}

func (T) Method(a int, b string) (c float64, d error) {
	var x map[string][]int
	var y interface{}
	var z func(a, b int) (c string, d error)
	return 0, nil
}
`

	// Parse the file
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Find the variable declarations and test getTypeString on their types
	var foundTypes []ast.Expr
	ast.Inspect(
		f, func(n ast.Node) bool {
			if decl, ok := n.(*ast.ValueSpec); ok {
				if decl.Type != nil {
					foundTypes = append(foundTypes, decl.Type)
				}
			}
			return true
		})

	expectedTypes := []string{
		"map[string][]int",
		"interface{}",
		"func(a int, b int) (c string, d error)",
	}

	if len(foundTypes) != len(expectedTypes) {
		t.Fatalf("Expected to find %d types, found %d", len(expectedTypes), len(foundTypes))
	}

	for i, expr := range foundTypes {
		got := getTypeString(expr)
		if got != expectedTypes[i] {
			t.Errorf("getTypeString() = %v, want %v", got, expectedTypes[i])
		}
	}
}
