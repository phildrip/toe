package simple

type StubMyInterfaceCalculateCall struct {
	x int
	y int
}
type StubMyInterfaceGetValueCall struct {
}
type StubMyInterfaceSetValueCall struct {
	val string
}
type StubMyInterface struct {
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

func (s *StubMyInterface) Calculate(x int, y int) (int, error) {
	s.CalculateCalls = append(s.CalculateCalls, StubMyInterfaceCalculateCall{x: x, y: y})
	if s.CalculateFunc != nil {
		return s.CalculateFunc(x, y)
	} else {
		return s.CalculateReturns0, s.CalculateReturns1
	}
}
func (s *StubMyInterface) GetValue() string {
	s.GetValueCalls = append(s.GetValueCalls, StubMyInterfaceGetValueCall{})
	if s.GetValueFunc != nil {
		return s.GetValueFunc()
	} else {
		return s.GetValueReturns0
	}
}
func (s *StubMyInterface) SetValue(val string) {
	s.SetValueCalls = append(s.SetValueCalls, StubMyInterfaceSetValueCall{val: val})
	return
}
