package main

type Thinger interface {
	Thing() error
	ThingWithParam(arg1 int) error
}
