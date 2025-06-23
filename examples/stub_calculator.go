package calculator

import "sync"

type StubCalculatorAddCall struct {
	A int
	B int
}
type StubCalculatorSubtractCall struct {
	A int
	B int
}
type StubCalculator struct {
	mu               sync.Mutex
	_isLocked        bool
	AddFunc          func(a int, b int) int
	AddCalls         []StubCalculatorAddCall
	AddReturns0      int
	SubtractFunc     func(a int, b int) (int, error)
	SubtractCalls    []StubCalculatorSubtractCall
	SubtractReturns0 int
	SubtractReturns1 error
}

func NewStubCalculator(withLocking bool) *StubCalculator {
	return &StubCalculator{_isLocked: withLocking}
}
func (s *StubCalculator) Add(a int, b int) int {
	if s._isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.AddCalls = append(s.AddCalls, StubCalculatorAddCall{A: a, B: b})
	if s.AddFunc != nil {
		return s.AddFunc(a, b)
	} else {
		return s.AddReturns0
	}
}
func (s *StubCalculator) Subtract(a int, b int) (int, error) {
	if s._isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.SubtractCalls = append(s.SubtractCalls, StubCalculatorSubtractCall{A: a, B: b})
	if s.SubtractFunc != nil {
		return s.SubtractFunc(a, b)
	} else {
		return s.SubtractReturns0, s.SubtractReturns1
	}
}
