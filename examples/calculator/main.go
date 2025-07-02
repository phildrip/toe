package main

import (
	"examples/calculator/stubs"
	"fmt"
	"log"

	"github.com/phildrip/toe/options"
)

func main() {
	fmt.Println("Demonstrating Calculator Stub:")

	// Instantiate the stub, enabling locking for this instance
	stub := stubs.NewStubCalculator(options.StubOptions{WithLocking: true})

	// --- Using fixed return values ---
	fmt.Println("\n--- Testing Subtract with fixed return values ---")

	// Set a single return value for Subtract.
	// Note: SubtractReturns is now a single struct, not a slice.
	stub.SubtractReturns = stubs.StubCalculatorSubtractReturns{R0: 100, R1: nil}
	result, err := stub.Subtract(20, 10)
	if err != nil {
		log.Fatalf("Error from Subtract: %v", err)
	}
	fmt.Printf("Subtract(20, 10) returned: %d, %v\n", result, err)

	// If you want to demonstrate an error return, set it again
	stub.SubtractReturns = stubs.StubCalculatorSubtractReturns{R0: 0,
		R1: fmt.Errorf("simulated error")}
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

	// Demonstrating call recording (optional, can be removed if not desired in example)
	fmt.Println("\n--- Call Recording ---")
	fmt.Printf("Subtract calls: %+v\n", stub.SubtractCalls)
	fmt.Printf("Add calls: %+v\n", stub.AddCalls)
}
