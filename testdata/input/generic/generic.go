package generic

type GenericInterface[T any] interface {
	Do(value T) (T, error)
	Get() T
}
