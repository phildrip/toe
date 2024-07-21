package gen

import (
	"go/ast"
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
					t.Errorf("FindInterface() gotPackageName = %v, want %v", gotPackageName, tt.wantPackageName)
				}

				if len(gotMethods) != len(tt.wantMethods) {
					t.Errorf("FindInterface() gotMethods length = %v, want %v", len(gotMethods), len(tt.wantMethods))
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
				"func (s *StubInterface) Method2() error",
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
				"func (s *StubInterfaceWithParam) Method1(arg1 int) error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				stubCode, err := GenerateStubCode(tt.interfaceName, tt.methods, tt.packageName, tt.disableFormatting)
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
