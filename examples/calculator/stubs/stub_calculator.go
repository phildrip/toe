package stubs

import (
	"github.com/phildrip/toe/options"
	"sync"
)

type StubCalculatorAddCall struct {
	A int
	B int
}
type StubCalculatorAddReturns struct {
	R0 int
}
type StubCalculatorSubtractCall struct {
	A int
	B int
}
type StubCalculatorSubtractReturns struct {
	R0 int
	R1 error
}
type StubCalculator struct {
	mu              sync.Mutex
	isLocked        bool
	AddFunc         func(a int, b int) int
	AddCalls        []StubCalculatorAddCall
	AddReturns      StubCalculatorAddReturns
	SubtractFunc    func(a int, b int) (int, error)
	SubtractCalls   []StubCalculatorSubtractCall
	SubtractReturns StubCalculatorSubtractReturns
}

func NewStubCalculator(opts options.StubOptions) *StubCalculator {
	return &StubCalculator{isLocked: opts.WithLocking}
}
func (s *StubCalculator) Add(a int, b int) int {
	if s.isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.AddCalls = append(s.AddCalls, StubCalculatorAddCall{A: a, B: b})
	if s.AddFunc != nil {
		return s.AddFunc(a, b)
	} else {
		return s.AddReturns.R0
	}
}
func (s *StubCalculator) Subtract(a int, b int) (int, error) {
	if s.isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.SubtractCalls = append(s.SubtractCalls, StubCalculatorSubtractCall{A: a, B: b})
	if s.SubtractFunc != nil {
		return s.SubtractFunc(a, b)
	} else {
		return s.SubtractReturns.R0, s.SubtractReturns.R1
	}
}
