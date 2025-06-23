# toe

#### trip-free test stubs

`toe` is a Go tool that automatically generates stub implementations for Go interfaces. It's useful for creating test doubles in unit tests. It is inspired by `pegomock` but focuses exclusively on stubbing.

## Features

- Generates stub implementations for any Go interface, **including those with generic type parameters**.
- **Configurable Concurrency Safety**: Stubs can be instantiated with or without a `sync.Mutex` at runtime to help detect race conditions in your code.
- **Call Recording**: All method calls are recorded, allowing you to assert how many times a method was called and with which parameters.
- **Flexible Return Values**: You can set up stubbed methods to return specific fixed values or to execute a custom lambda function for more complex logic.

## Installation

To install `toe`, use the following command:

```bash
go get github.com/phildrip/toe
```

## Usage

```bash
toe [flags] <input_directory> <interface>
```

-   `<input_directory>`: The directory containing the Go file with the interface definition.
-   `<interface>`: The name of the interface you want to generate a stub for.
-   `-o <output.go>`: (Optional) The output file name. If not provided, the stub code is printed to stdout.
-   `-test-package`: (Optional) If provided, the generated stub will be in a `_test` package (e.g., `package mypackage_test`).
-   `-stub-dir <dir>`: (Optional) Generate the stub in a specific subdirectory (e.g., `stubs`) and use its base name as the package name (e.g., `package stubs`). This flag is incompatible with `-test-package` if the generated package name would conflict. 

### Example

```bash
# Generate a standard stub file for the Calculator interface
./toe -o ./examples/stub_calculator.go ./examples/calculator Calculator

# Generate a stub in a _test package
./toe -test-package -o ./examples/stub_calculator_test.go ./examples/calculator Calculator

# Generate a stub in a 'stubs' subdirectory, with 'stubs' as the package name
# (You might need to create the 'stubs' directory first)
./toe -stub-dir stubs -o ./examples/stubs/stub_calculator.go ./examples/calculator Calculator
```
```

## Generated Stub Structure

`toe` generates a struct (e.g., `StubCalculator`) that implements your interface, along with a constructor function (e.g., `NewStubCalculator`).

-   **Constructor**: A `NewStub<InterfaceName>` function is generated which allows you to instantiate the stub with configurable locking. For example: `NewStubCalculator(withLocking bool) *StubCalculator`.
-   **Internal Fields**: The generated stub struct includes:
    -   `mu sync.Mutex`: (Always present, but only used if `withLocking` is true in the constructor).
    -   `_isLocked bool`: A flag indicating if the mutex should be used for this instance.
    -   For each method in the interface, the stub contains three additional fields:
        -   `MethodNameFunc`: A field to assign a lambda function (`func(...) (...)`) that will be executed when the method is called. This takes precedence over fixed return values.
        -   `MethodNameCalls`: A slice of structs that records each call to the method and its parameters.
        -   `MethodNameReturnsX`: Fields that hold fixed return values for the method (e.g., `DoSomethingReturns0`, `DoSomethingReturns1`).

## Example Usage in Tests

Given an interface `Calculator`:

```go
// examples/calculator/calculator.go
package calculator

type Calculator interface {
	Add(a, b int) int
	Subtract(a, b int) (int, error)
}
```

A generated stub (`stub_calculator.go`) can be used in tests as follows:

```go
package calculator_test

import (
	"errors"
	"testing"

	"github.com/phildrip/toe/examples"
)

func TestCalculatorStub(t *testing.T) {
	// Instantiate the stub, enabling locking for this instance
	stub := examples.NewStubCalculator(true)

	// --- Test 1: Using fixed return values ---
	stub.SubtractReturns0 = 10
	stub.SubtractReturns1 = nil

	result, err := stub.Subtract(20, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != 10 {
		t.Errorf("Expected result 10, got %d", result)
	}

	// Assert that Subtract was called once
	if len(stub.SubtractCalls) != 1 {
		t.Errorf("Expected 1 call to Subtract, got %d", len(stub.SubtractCalls))
	}
	// Assert on call parameters
	if stub.SubtractCalls[0].A != 20 || stub.SubtractCalls[0].B != 10 {
		t.Errorf("Incorrect parameters for Subtract: got %+v", stub.SubtractCalls[0])
	}

	// --- Test 2: Using a lambda function ---
	stub.AddFunc = func(a, b int) int {
		// Custom logic
		return a*2 + b*2
	}

	addResult := stub.Add(5, 5)
	if addResult != 20 { // 5*2 + 5*2 = 20
		t.Errorf("Expected AddFunc result 20, got %d", addResult)
	}

	// Assert that Add was called once
	if len(stub.AddCalls) != 1 {
		t.Errorf("Expected 1 call to Add, got %d", len(stub.AddCalls))
	}
}
```

## Building from Source

To build `toe` from source:

1.  Clone the repository:
    ```bash
    git clone https://github.com/phildrip/toe.git
    ```
2.  Navigate to the project directory:
    ```bash
    cd toe
    ```
3.  Build the project:
    ```bash
    go build
    ```

## License

MIT License

## Author

Phil Richards (https://github.com/phildrip)
