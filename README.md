# toe

#### trip-free test stubs

`toe` is a Go tool that automatically generates stub implementations for Go interfaces. It's useful for creating test doubles in unit tests. It is inspired by `pegomock` but focuses exclusively on stubbing.

## Features

- Generates stub implementations for any Go interface, **including those with generic type parameters**.
- **Configurable Concurrency Safety**: Stubs can be configured to use a `sync.Mutex`, which is useful for testing concurrent code and helping to detect race conditions. This behavior is managed at runtime for each stub instance via an `options.StubOptions` struct.
- **Call Recording**: All method calls are recorded, allowing you to assert how many times a method was called and with which parameters.
- **Flexible Return Values**: You can set up stubbed methods to return specific fixed values or to execute a custom lambda function for more complex logic.

## Installation

To install `toe`, use the following command:

```bash
go install github.com/phildrip/toe@latest
```

## Usage

```bash
toe [flags] <input_directory> <interface>
```

-   `<input_directory>`: The directory containing the Go file with the interface definition.
-   `<interface>`: The name of the interface you want to generate a stub for.
-   `-o <output.go>`: (Optional) The output file name. If not provided, the stub code is printed to stdout.
-   `-stub-dir <dir>`: (Optional) Generate the stub in a specific subdirectory (e.g., `stubs`) and use its base name as the package name (e.g., `package stubs`).

### Example

```bash
# Generate a standard stub file for the Calculator interface
./toe -o ./examples/calculator/stubs/stub_calculator.go ./examples/calculator/lib Calculator

# Generate a stub in a 'stubs' subdirectory, with 'stubs' as the package name
# (You might need to create the 'stubs' directory first)
./toe -stub-dir stubs -o ./examples/calculator/stubs/stub_calculator.go ./examples/calculator/lib Calculator
```

## Generated Stub Structure

`toe` generates a struct (e.g., `StubCalculator`) that implements your interface, along with a constructor function (e.g., `NewStubCalculator`).

-   **Constructor**: A `NewStub<InterfaceName>` function is generated which allows you to instantiate the stub with configurable options. For example: `NewStubCalculator(opts *options.StubOptions) *StubCalculator`.
-   **Internal Fields**: The generated stub struct includes:
    -   `mu sync.Mutex`: (Always present, but only used if `opts.WithLocking` is true in the constructor).
    -   `isLocked bool`: A flag indicating if the mutex should be used for this instance.
    -   For each method in the interface, the stub contains three additional fields:
        -   `MethodNameFunc`: A field to assign a lambda function (`func(...) (...)`) that will be executed when the method is called. This takes precedence over fixed return values.
        -   `MethodNameCalls`: A slice of structs that records each call to the method and its parameters.
        -   `MethodNameReturns`: A struct that holds fixed return values for the method. Unnamed return values will be prefixed by their type (e.g., `Int0`, `Error1`).

## Example Usage

Given an interface `Calculator`:

```go
// examples/calculator/lib/calculator.go
package lib

type Calculator interface {
	Add(a, b int) int
	Subtract(a, b int) (int, error)
}
```

A generated stub (e.g., `examples/calculator/stubs/stub_calculator.go`) can be used as follows:

```go
// examples/calculator/main.go
package main

import (
	"fmt"
	"log"

	"github.com/phildrip/toe/examples/calculator/stubs"
)

func main() {
	fmt.Println("Demonstrating Calculator Stub:")

	// Instantiate the stub, enabling locking for this instance
	stub := stubs.NewStubCalculator(&stubs.StubOptions{WithLocking: true})

	// --- Using fixed return values ---
	fmt.Println("\n--- Testing Subtract with fixed return values ---")
	
	// Set a single return value for Subtract.
		stub.SubtractReturns = stubs.StubCalculatorSubtractReturns{Int0: 100, Error1: nil}
	result, err := stub.Subtract(20, 10)
	if err != nil {
		log.Fatalf("Error from Subtract: %v", err)
	}
	fmt.Printf("Subtract(20, 10) returned: %d, %v\n", result, err)

	// If you want to demonstrate an error return, set it again
	stub.SubtractReturns = stubs.StubCalculatorSubtractReturns{Int0: 0, Error1: fmt.Errorf("simulated error")}
	result, err = stub.Subtract(5, 3)
	fmt.Printf("Subtract(5, 3) returned: %d, %v\n", result, err)

	// --- Using a lambda function ---
	fmt.Println("\n--- Testing Add with a lambda function ---")
	stub.AddFunc = func(a, b int) int {
		fmt.Printf("  (AddFunc called with a=%d, b=%d)\n", a, b)
		return a*10 + b*10 // Custom logic
	}

	addResult := stub.Add(5, 5)
	fmt.Printf("Add(5, 5) returned: %d\n", addResult)

	// Demonstrating call recording
	fmt.Println("\n--- Call Recording ---")
	fmt.Printf("Subtract calls: %+v\n", stub.SubtractCalls)
	fmt.Printf("Add calls: %+v\n", stub.AddCalls)
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
