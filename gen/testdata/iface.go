package testdata

type Interface interface {
	Method1()
	Method2() error
}

type EmptyInterface interface {
}

type InterfaceWithParam interface {
	Method1(arg1 int) error
}
