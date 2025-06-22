package generic

import "sync"

type StubGenericInterfaceDoCall struct {
	value T
}
type StubGenericInterfaceGetCall struct {
}
type StubGenericInterface[T any] struct {
	mu          sync.Mutex
	DoFunc      func(value T) (T, error)
	DoCalls     []StubGenericInterfaceDoCall
	DoReturns0  T
	DoReturns1  error
	GetFunc     func() T
	GetCalls    []StubGenericInterfaceGetCall
	GetReturns0 T
}

func (s *StubGenericInterface[T]) Do(value T) (T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DoCalls = append(s.DoCalls, StubGenericInterfaceDoCall{value: value})
	if s.DoFunc != nil {
		return s.DoFunc(value)
	} else {
		return s.DoReturns0, s.DoReturns1
	}
}
func (s *StubGenericInterface[T]) Get() T {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GetCalls = append(s.GetCalls, StubGenericInterfaceGetCall{})
	if s.GetFunc != nil {
		return s.GetFunc()
	} else {
		return s.GetReturns0
	}
}
