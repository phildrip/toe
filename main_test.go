package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCases defines the structure for our golden file test cases.
type TestCase struct {
	Name          string
	InputFile     string
	InterfaceName string
	Options       *StubOptions
	GoldenFile    string
}

func TestGenerateStub(t *testing.T) {
	// Define test cases
	testCases := []TestCase{
		{
			Name:          "simple_unlocked",
			InputFile:     filepath.Join("testdata", "input", "simple"),
			InterfaceName: "MyInterface",
			Options:       &StubOptions{WithLocking: false},
			GoldenFile:    filepath.Join("testdata", "golden", "simple_unlocked.go"),
		},
		{
			Name:          "simple_locked",
			InputFile:     filepath.Join("testdata", "input", "simple"),
			InterfaceName: "MyInterface",
			Options:       &StubOptions{WithLocking: true},
			GoldenFile:    filepath.Join("testdata", "golden", "simple_locked.go"),
		},
		{
			Name:          "generic_unlocked",
			InputFile:     filepath.Join("testdata", "input", "generic"),
			InterfaceName: "GenericInterface",
			Options:       &StubOptions{WithLocking: false},
			GoldenFile:    filepath.Join("testdata", "golden", "generic_unlocked.go"),
		},
		{
			Name:          "generic_locked",
			InputFile:     filepath.Join("testdata", "input", "generic"),
			InterfaceName: "GenericInterface",
			Options:       &StubOptions{WithLocking: true},
			GoldenFile:    filepath.Join("testdata", "golden", "generic_locked.go"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Create a temporary output file
			outputFilePath := filepath.Join(t.TempDir(), "generated_stub.go")

			// Prepare arguments for the run function
			args := []string{"toe", "-o", outputFilePath}
			if tc.Options.WithLocking {
				args = append(args, "-with-locking")
			}
			args = append(args, tc.InputFile, tc.InterfaceName)

			// Capture stdout/stderr
			var outBuffer, errBuffer bytes.Buffer

			// Run the main logic through the run() function
			// The `run` function in main.go is now exported as `Run`.
			// So, we need to call `Run` here.
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
