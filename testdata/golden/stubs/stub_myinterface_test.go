package stubs_test

import (
	"github.com/phildrip/toe/options"
	"sync"
)

type StubMyInterfaceCalculateCall struct {
	X int
	Y int
}
type StubMyInterfaceCalculateReturns struct {
	Int0   int
	Error1 error
}
type StubMyInterfaceGetValueCall struct {
}
type StubMyInterfaceGetValueReturns struct {
	String0 string
}
type StubMyInterfaceSetValueCall struct {
	Val string
}
type StubMyInterface struct {
	mu               sync.Mutex
	isLocked         bool
	CalculateFunc    func(x int, y int) (int, error)
	CalculateCalls   []StubMyInterfaceCalculateCall
	CalculateReturns StubMyInterfaceCalculateReturns
	GetValueFunc     func() string
	GetValueCalls    []StubMyInterfaceGetValueCall
	GetValueReturns  StubMyInterfaceGetValueReturns
	SetValueFunc     func(val string)
	SetValueCalls    []StubMyInterfaceSetValueCall
}

func NewStubMyInterface(opts options.StubOptions) *StubMyInterface {
	return &StubMyInterface{isLocked: opts.WithLocking}
}
func (s *StubMyInterface) Calculate(x int, y int) (int, error) {
	if s.isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.CalculateCalls = append(s.CalculateCalls, StubMyInterfaceCalculateCall{X: x, Y: y})
	if s.CalculateFunc != nil {
		return s.CalculateFunc(x, y)
	} else {
		return s.CalculateReturns.Int0, s.CalculateReturns.Error1
	}
}
func (s *StubMyInterface) GetValue() string {
	if s.isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.GetValueCalls = append(s.GetValueCalls, StubMyInterfaceGetValueCall{})
	if s.GetValueFunc != nil {
		return s.GetValueFunc()
	} else {
		return s.GetValueReturns.String0
	}
}
func (s *StubMyInterface) SetValue(val string) {
	if s.isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.SetValueCalls = append(s.SetValueCalls, StubMyInterfaceSetValueCall{Val: val})
	return
}
