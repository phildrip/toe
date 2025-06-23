package generic

import "sync"

type StubGenericInterfaceDoCall[T any] struct {
	Value T
}
type StubGenericInterfaceGetCall[T any] struct {
}
type StubGenericInterface[T any] struct {
	mu          sync.Mutex
	_isLocked   bool
	DoFunc      func(value T) (T, error)
	DoCalls     []StubGenericInterfaceDoCall[T]
	DoReturns0  T
	DoReturns1  error
	GetFunc     func() T
	GetCalls    []StubGenericInterfaceGetCall[T]
	GetReturns0 T
}

func NewStubGenericInterface[T any](withLocking bool) *StubGenericInterface[T] {
	return &StubGenericInterface[T]{_isLocked: withLocking}
}
func (s *StubGenericInterface[T]) Do(value T) (T, error) {
	if s._isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.DoCalls = append(s.DoCalls, StubGenericInterfaceDoCall[T]{Value: value})
	if s.DoFunc != nil {
		return s.DoFunc(value)
	} else {
		return s.DoReturns0, s.DoReturns1
	}
}
func (s *StubGenericInterface[T]) Get() T {
	if s._isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	s.GetCalls = append(s.GetCalls, StubGenericInterfaceGetCall[T]{})
	if s.GetFunc != nil {
		return s.GetFunc()
	} else {
		return s.GetReturns0
	}
}
