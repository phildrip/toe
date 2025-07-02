package lib

type Calculator interface {
	Add(a, b int) int
	Subtract(a, b int) (int, error)
}
