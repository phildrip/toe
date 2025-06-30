package calculator_test

import (
	"testing"

	calculator "github.com/phildrip/toe/examples/calculator"
)

func TestCalculatorStub(t *testing.T) {
	// Instantiate the stub, enabling locking for this instance
	stub := calculator.NewStubCalculator(true)

	// --- Test 1: Using fixed return values ---
	stub.SubtractReturns.R0 = 10
	stub.SubtractReturns.R1 = nil

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
	// Assert on call parameters - now public fields A and B
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
	// Assert on call parameters - now public fields A and B
	if stub.AddCalls[0].A != 5 || stub.AddCalls[0].B != 5 {
		t.Errorf("Incorrect parameters for Add: got %+v", stub.AddCalls[0])
	}
}
