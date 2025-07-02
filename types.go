package main

import (
	"go/types"
)

// ParamData represents a parameter or result in a method signature.
type ParamData struct {
	Name string
	Type types.Type // Store the actual types.Type object
}

// MethodData represents a method of an interface.
type MethodData struct {
	Name    string
	Params  []ParamData
	Results []ResultData
}

// ResultData represents a result in a method signature.
type ResultData struct {
	Name string
	Type types.Type // Store the actual types.Type object
}

// InterfaceData represents a parsed interface, including its methods and type parameters.
type InterfaceData struct {
	PackageName string
	Name        string
	Methods     []MethodData
	TypeParams  []ParamData       // For generic interfaces, e.g., [T comparable]
	Imports     map[string]string // map[importPath]packageName
}
