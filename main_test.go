package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"strings"
)

// TestCases defines the structure for our golden file test cases.
type TestCase struct {
	Name          string
	InputFile     string
	InterfaceName string
	GoldenFile    string
	Flags         []string // Additional flags for toe command
}

func TestGenerateStub(t *testing.T) {
	// Define test cases
	testCases := []TestCase{
		{
			Name:          "simple_default_output",
			InputFile:     filepath.Join("testdata", "input", "simple"),
			InterfaceName: "MyInterface",
			GoldenFile:    filepath.Join("testdata", "golden", "stubs", "stub_myinterface.go"), // Default output path
			Flags:         []string{},                                                             // No flags for default
		},
		{
			Name:          "generic_default_output",
			InputFile:     filepath.Join("testdata", "input", "generic"),
			InterfaceName: "GenericInterface",
			GoldenFile:    filepath.Join("testdata", "golden", "stubs", "stub_genericinterface.go"), // Default output path
			Flags:         []string{},                                                              // No flags for default
		},
		{
			Name:          "simple_test_package",
			InputFile:     filepath.Join("testdata", "input", "simple"),
			InterfaceName: "MyInterface",
			GoldenFile:    filepath.Join("testdata", "golden", "stubs", "stub_myinterface_test.go"), // _test package output path
			Flags:         []string{"--test-package"},
		},
		{
			Name:          "simple_custom_stub_dir",
			InputFile:     filepath.Join("testdata", "input", "simple"),
			InterfaceName: "MyInterface",
			GoldenFile:    filepath.Join("testdata", "golden", "customstubs", "stub_myinterface.go"), // Custom stub dir output path
			Flags:         []string{"--stub-dir", "customstubs"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Determine the expected output path that `run` will use
			// This will be relative to the test's TempDir
			generatedStubDir := "stubs" // Default for most cases
			
			// If a custom stub-dir is specified in flags, use that for the generated output path.
			// This logic needs to mirror how `run` determines the actual output directory.
			for i, flag := range tc.Flags {
				if flag == "--stub-dir" && i+1 < len(tc.Flags) {
					generatedStubDir = tc.Flags[i+1]
					break
				}
			}

			outputFilename := fmt.Sprintf("stub_%s.go", strings.ToLower(tc.InterfaceName))
			// If it's a test package, the filename also changes to include _test suffix
			if strings.Contains(tc.Name, "test_package") {
				outputFilename = fmt.Sprintf("stub_%s_test.go", strings.ToLower(tc.InterfaceName))
			}

			// Create a temporary output file path. 
			testTempDir := t.TempDir()
			finalOutputDir := filepath.Join(testTempDir, generatedStubDir) // Use generatedStubDir here
			if err := os.MkdirAll(finalOutputDir, 0755); err != nil {
				t.Fatalf("Failed to create temp output directory: %v", err)
			}
			outputFilePath := filepath.Join(finalOutputDir, outputFilename)

			// Prepare arguments for the run function
			args := []string{"toe"}
			args = append(args, tc.Flags...)
			// The run function now automatically determines the output path and filename.
			// However, we pass -o to force it to write to our temporary file path for comparison.
			args = append(args, "-o", outputFilePath, tc.InputFile, tc.InterfaceName)

			// Capture stdout/stderr
			var outBuffer, errBuffer bytes.Buffer

			// Run the main logic through the run() function
			exitCode := run(&outBuffer, &errBuffer, args)
			if exitCode != 0 {
				t.Fatalf("toe exited with non-zero status: %d\nStderr: %s", exitCode, errBuffer.String())
			}

			if errBuffer.Len() > 0 {
				t.Fatalf("toe produced errors: %s", errBuffer.String())
			}

			// Read generated and golden files
			generated, err := os.ReadFile(outputFilePath)
			if err != nil {
				t.Fatalf("Failed to read generated file %s: %v", outputFilePath, err)
			}
			golden, err := os.ReadFile(tc.GoldenFile)
			if err != nil {
				t.Fatalf("Failed to read golden file %s: %v", tc.GoldenFile, err)
			}

			// Compare generated to golden
			if !bytes.Equal(generated, golden) {
				diff := generateDiff(generated, golden)
				t.Errorf("Generated output for %s does not match golden file.\nDiff:\n%s", tc.Name, diff)
			}

			// Verify generated code compiles
			cmd := exec.Command("go", "build", outputFilePath)
			var buildStderr bytes.Buffer
			cmd.Stderr = &buildStderr // Capture stderr from go build
			cmd.Dir = t.TempDir()      // Run build in temp dir to avoid polluting module
			if err := cmd.Run(); err != nil {
				t.Fatalf("Generated code for %s failed to compile: %v\nStderr: %s", tc.Name, err, buildStderr.String())
			}
		})
	}
}

// generateDiff is a helper to produce a diff string (simplified for demonstration)
func generateDiff(a, b []byte) string {
	// In a real scenario, use a proper diffing library like github.com/sergi/go-diff
	// For simplicity, we'll just show both versions.
	return fmt.Sprintf("--- Generated\n+++ Golden\n%s\n%s", string(a), string(b))
}
