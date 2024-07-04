package ref_stubs

import "sync"

// type Thinger interface {
//	Thing() error
//	ThingWithParam(arg1 int) error
// }

type ThingWithParamsRet struct {
	R0 string
	R1 error
}

type ThingWithParamRet struct {
	R0 error
}

type ThingRet struct {
	R0 error
}

type ThingParams struct{}

type ThingWithParamsParams struct {
	Arg1 int
	Arg2 string
}

type ThingWithParamParams struct {
	Arg1 int
}

func NewStubThinger() *StubThinger {
	stub := &StubThinger{}
	stub.StubThingThen = &StubThingThen{
		stub: stub,
	}
	stub.StubThingWithParamThen = &StubThingWithParamThen{
		stub: stub,
	}
	stub.StubThingWithParamsThen = &StubThingWithParamsThen{
		stub: stub,
	}
	return stub
}

type StubThinger struct {
	ThingRet           ThingRet
	ThingWithParamRet  ThingWithParamRet
	ThingWithParamsRet ThingWithParamsRet

	ThingCalls           []ThingParams
	ThingWithParamCalls  []ThingWithParamParams
	ThingWithParamsCalls []ThingWithParamsParams

	StubThingThen           *StubThingThen
	StubThingWithParamThen  *StubThingWithParamThen
	StubThingWithParamsThen *StubThingWithParamsThen

	mut sync.Mutex
}

func (s *StubThinger) Thing() error {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.ThingCalls = append(s.ThingCalls, ThingParams{})
	return s.ThingRet.R0
}

func (s *StubThinger) ThingWithParam(arg1 int) error {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.ThingWithParamCalls = append(s.ThingWithParamCalls, ThingWithParamParams{
		Arg1: arg1,
	})
	return s.ThingWithParamRet.R0
}

func (s *StubThinger) ThingWithParams(arg1 int, arg2 string) (string, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.ThingWithParamsCalls = append(s.ThingWithParamsCalls, ThingWithParamsParams{
		Arg1: arg1,
		Arg2: arg2,
	})
	return s.ThingWithParamsRet.R0, s.ThingWithParamsRet.R1
}

type StubThingThen struct {
	stub *StubThinger
}

func (s *StubThingThen) Return(R0 error) {
	s.stub.ThingRet = ThingRet{
		R0: R0,
	}
}

type StubThingWithParamThen struct {
	stub *StubThinger
}

func (s *StubThingWithParamThen) Return(R0 error) {
	s.stub.ThingWithParamRet = ThingWithParamRet{
		R0: R0,
	}
}

type StubThingWithParamsThen struct {
	stub *StubThinger
}

func (s *StubThingWithParamsThen) Return(R0 string, R1 error) {
	s.stub.ThingWithParamsRet = ThingWithParamsRet{
		R0: R0,
		R1: R1,
	}
}

func (s *StubThinger) OnThing() *StubThingThen {
	return s.StubThingThen
}

func (s *StubThinger) OnThingWithParam() *StubThingWithParamThen {
	return s.StubThingWithParamThen
}

func (s *StubThinger) OnThingWithParams() *StubThingWithParamsThen {
	return s.StubThingWithParamsThen
}
