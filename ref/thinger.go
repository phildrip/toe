package ref

type Thinger interface {
	Thing() error
	ThingWithParam(arg1 int) error
}
