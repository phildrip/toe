package ref_test

import (
	"errors"
	refstubs "stubz/ref/stubs"
	"testing"
)

func TestRef(t *testing.T) {
	stub := refstubs.NewStubThinger()

	err1 := errors.New("error1")
	err2 := errors.New("error2")
	err3 := errors.New("error3")

	stub.OnThing().Return(err1)
	stub.OnThingWithParam().Return(err2)
	expectedString := "a string"
	stub.OnThingWithParams().Return(expectedString, err3)

	err := stub.Thing()
	if err != err1 {
		t.Errorf("expected %v, got %v", err1, err)
	}

	if len(stub.ThingCalls) != 1 {
		t.Errorf("expected %v, got %v", 1, len(stub.ThingCalls))
	}

	err = stub.ThingWithParam(1)
	if err != err2 {
		t.Errorf("expected %v, got %v", err2, err)
	}

	if len(stub.ThingWithParamCalls) != 1 {
		t.Errorf("expected %v, got %v", 1, len(stub.ThingWithParamCalls))
	}

	outStr, err := stub.ThingWithParams(1, expectedString)
	if stub.ThingWithParamsCalls[0].Arg1 != 1 {
		t.Errorf("expected %v, got %v", 1, stub.ThingWithParamsCalls[0].Arg1)
	}

	if stub.ThingWithParamsCalls[0].Arg2 != expectedString {
		t.Errorf("expected %v, got %v", expectedString, stub.ThingWithParamsCalls[0].Arg2)
	}

	if len(stub.ThingWithParamsCalls) != 1 {
		t.Errorf("expected %v, got %v", 1, len(stub.ThingWithParamsCalls))
	}

	if outStr != expectedString {
		t.Errorf("expected %v, got %v", expectedString, outStr)
	}

}
