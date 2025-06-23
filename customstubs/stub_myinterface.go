package customstubs

import "sync"

type StubMyInterfaceCalculateCall struct {
	X int
	Y int
}
type StubMyInterfaceGetValueCall struct {
}
type StubMyInterfaceSetValueCall struct {
	Val string
}
type StubMyInterface struct {
	mu                sync.Mutex
	_isLocked         bool
	CalculateFunc     func(x int, y int) (int, error)
	CalculateCalls    []StubMyInterfaceCalculateCall
	CalculateReturns0 int
	CalculateReturns1 error
	GetValueFunc      func() string
	GetValueCalls     []StubMyInterfaceGetValueCall
	GetValueReturns0  string
	SetValueFunc      func(val string)
	SetValueCalls     []StubMyInterfaceSetValueCall
}

func NewStubMyInterface(withLocking bool) *StubMyInterface {
	return &StubMyInterface{_isLocked: withLocking}
}
func (s *StubMyInterface) Calculate(x int, y int) (int, error) {
	if s._isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.CalculateCalls = append(s.CalculateCalls, StubMyInterfaceCalculateCall{X: x, Y: y})
	if s.CalculateFunc != nil {
		return s.CalculateFunc(x, y)
	} else {
		return s.CalculateReturns0, s.CalculateReturns1
	}
}
func (s *StubMyInterface) GetValue() string {
	if s._isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.GetValueCalls = append(s.GetValueCalls, StubMyInterfaceGetValueCall{})
	if s.GetValueFunc != nil {
		return s.GetValueFunc()
	} else {
		return s.GetValueReturns0
	}
}
func (s *StubMyInterface) SetValue(val string) {
	if s._isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.SetValueCalls = append(s.SetValueCalls, StubMyInterfaceSetValueCall{Val: val})
	return
}
