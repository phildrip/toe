package generic

type StubGenericInterfaceDoCall[T any] struct {
	value T
}
type StubGenericInterfaceGetCall[T any] struct {
}
type StubGenericInterface[T any] struct {
	DoFunc      func(value T) (T, error)
	DoCalls     []StubGenericInterfaceDoCall[T]
	DoReturns0  T
	DoReturns1  error
	GetFunc     func() T
	GetCalls    []StubGenericInterfaceGetCall[T]
	GetReturns0 T
}

func (s *StubGenericInterface[T]) Do(value T) (T, error) {
	s.DoCalls = append(s.DoCalls, StubGenericInterfaceDoCall[T]{value: value})
	if s.DoFunc != nil {
		return s.DoFunc(value)
	} else {
		return s.DoReturns0, s.DoReturns1
	}
}
func (s *StubGenericInterface[T]) Get() T {
	s.GetCalls = append(s.GetCalls, StubGenericInterfaceGetCall[T]{})
	if s.GetFunc != nil {
		return s.GetFunc()
	} else {
		return s.GetReturns0
	}
}
