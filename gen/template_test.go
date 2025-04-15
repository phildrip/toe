package gen

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
)

func TestStubTemplate(t *testing.T) {
	// Create a template with the same functions as in GenerateStubCode
	funcMap := template.FuncMap{
		"join":  strings.Join,
		"zip":   zip,
		"joinl": joinl,
	}

	tmpl := template.Must(
		template.New("stub").
			Funcs(funcMap).
			Parse(stubTemplate))

	tests := []struct {
		name            string
		packageName     string
		interfaceName   string
		stubName        string
		methods         []methodData
		expectedPhrases []string
	}{
		{
			name:          "Simple interface",
			packageName:   "testpkg",
			interfaceName: "SimpleInterface",
			stubName:      "StubSimpleInterface",
			methods: []methodData{
				{
					Name:        "Method1",
					Params:      []string{},
					ParamNames:  []string{},
					Results:     []string{},
					ResultNames: []string{},
				},
				{
					Name:        "Method2",
					Params:      []string{},
					ParamNames:  []string{},
					Results:     []string{"error"},
					ResultNames: []string{"R0"},
				},
			},
			expectedPhrases: []string{
				"package testpkg",
				"type StubSimpleInterface struct",
				"func (s *StubSimpleInterface) Method1()",
				"func (s *StubSimpleInterface) Method2() error",
			},
		},
		{
			name:          "Complex interface",
			packageName:   "complex",
			interfaceName: "ComplexInterface",
			stubName:      "StubComplexInterface",
			methods: []methodData{
				{
					Name:        "Process",
					Params:      []string{"ctx context.Context", "data string"},
					ParamNames:  []string{"ctx", "data"},
					Results:     []string{"[]string", "error"},
					ResultNames: []string{"R0", "R1"},
				},
			},
			expectedPhrases: []string{
				"package complex",
				"type StubComplexInterface struct",
				"func (s *StubComplexInterface) Process(ctx context.Context, data string) ([]string, error)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				var buf bytes.Buffer
				err := tmpl.Execute(
					&buf, struct {
						PackageName   string
						InterfaceName string
						StubName      string
						Methods       []methodData
					}{
						PackageName:   tt.packageName,
						InterfaceName: tt.interfaceName,
						StubName:      tt.stubName,
						Methods:       tt.methods,
					})

				if err != nil {
					t.Fatalf("Template execution failed: %v", err)
				}

				result := buf.String()
				for _, phrase := range tt.expectedPhrases {
					if !strings.Contains(result, phrase) {
						t.Errorf("Template output missing expected phrase: %s", phrase)
					}
				}
			})
	}
}
