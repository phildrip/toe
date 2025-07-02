package stubs

import (
	"github.com/phildrip/toe/options"
	"sync"
)

type StubGenericInterfaceDoCall[T any] struct {
	Value T
}
type StubGenericInterfaceDoReturns[T any] struct {
	T0     T
	Error1 error
}
type StubGenericInterfaceGetCall[T any] struct {
}
type StubGenericInterfaceGetReturns[T any] struct {
	T0 T
}
type StubGenericInterface[T any] struct {
	mu         sync.Mutex
	isLocked   bool
	DoFunc     func(value T) (T, error)
	DoCalls    []StubGenericInterfaceDoCall[T]
	DoReturns  StubGenericInterfaceDoReturns[T]
	GetFunc    func() T
	GetCalls   []StubGenericInterfaceGetCall[T]
	GetReturns StubGenericInterfaceGetReturns[T]
}

func NewStubGenericInterface[T any](opts options.StubOptions) *StubGenericInterface[T] {
	return &StubGenericInterface[T]{isLocked: opts.WithLocking}
}
func (s *StubGenericInterface[T]) Do(value T) (T, error) {
	if s.isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.DoCalls = append(s.DoCalls, StubGenericInterfaceDoCall[T]{Value: value})
	if s.DoFunc != nil {
		return s.DoFunc(value)
	} else {
		return s.DoReturns.T0, s.DoReturns.Error1
	}
}
func (s *StubGenericInterface[T]) Get() T {
	if s.isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.GetCalls = append(s.GetCalls, StubGenericInterfaceGetCall[T]{})
	if s.GetFunc != nil {
		return s.GetFunc()
	} else {
		return s.GetReturns.T0
	}
}
