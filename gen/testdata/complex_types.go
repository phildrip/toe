package testdata

import "context"

type ComplexInterface interface {
	// Method with context and multiple parameters
	ProcessWithContext(ctx context.Context, id string, options map[string]interface{}) ([]string, error)

	// Method with pointer parameters
	HandlePointers(data *string, count *int) *bool

	// Method with slices and maps
	TransformData(items []int, mapping map[string]int) map[int][]string

	// Method with variadic parameters
	ProcessVariadic(prefix string, values ...interface{}) error

	// Method with function parameter
	WithCallback(callback func(string) error) error
}
