package simple

type MyInterface interface {
	GetValue() string
	SetValue(val string)
	Calculate(x, y int) (int, error)
}
