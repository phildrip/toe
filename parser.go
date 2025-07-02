package main

import (
	"fmt"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// collectImports recursively traverses a go.types.Type and collects external package imports.
func collectImports(data *InterfaceData, t types.Type) {
	switch typ := t.(type) {
	case *types.Named:
		if typ.Obj().Pkg() != nil && typ.Obj().Pkg().Path() != data.PackageName {
			data.Imports[typ.Obj().Pkg().Path()] = typ.Obj().Pkg().Name()
		}
		// Also check underlying type, e.g., for struct fields of named types
		collectImports(data, typ.Underlying())
	case *types.Pointer:
		collectImports(data, typ.Elem())
	case *types.Slice:
		collectImports(data, typ.Elem())
	case *types.Array:
		collectImports(data, typ.Elem())
	case *types.Map:
		collectImports(data, typ.Key())
		collectImports(data, typ.Elem())
	case *types.Chan:
		collectImports(data, typ.Elem())
	case *types.Signature:
		if typ.Params() != nil {
			for i := 0; i < typ.Params().Len(); i++ {
				collectImports(data, typ.Params().At(i).Type())
			}
		}
		if typ.Results() != nil {
			for i := 0; i < typ.Results().Len(); i++ {
				collectImports(data, typ.Results().At(i).Type())
			}
		}
	case *types.Basic:
		// Basic types do not have associated packages.
	case *types.Interface:
		// Interface types themselves might not directly reference packages unless embedded named types.
		// Handled by recursive calls to collectImports on method signatures.
	case *types.TypeParam:
		// Type parameters themselves don't directly reference packages, but their constraints might.
		collectImports(data, typ.Constraint())
	}
}

func FindInterface(inputDir string,
	interfaceName string,
	testPackage bool,
	stubDir string) (*InterfaceData, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
		Dir: inputDir,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("load: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("packages contain errors")
	}

	var foundInterface *InterfaceData

	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()
		obj := scope.Lookup(interfaceName)
		if obj == nil {
			continue
		}

		ifaceType, ok := obj.Type().Underlying().(*types.Interface)
		if !ok {
			continue // Not an interface
		}

		namedType, isNamed := obj.Type().(*types.Named)

		if foundInterface != nil {
			return nil, fmt.Errorf("found duplicate interface %s in package %s and %s",
				interfaceName,
				foundInterface.PackageName,
				pkg.Name)
		}

		generatedPackageName := pkg.Name
		if stubDir != "" {
			generatedPackageName = filepath.Base(stubDir)
		}
		if testPackage {
			generatedPackageName += "_test"
		}

		data := &InterfaceData{
			PackageName: generatedPackageName,
			Name:        interfaceName,
			Imports:     make(map[string]string),
		}

		// Handle generic interfaces
		if isNamed && namedType.TypeParams() != nil {
			for i := 0; i < namedType.TypeParams().Len(); i++ {
				tp := namedType.TypeParams().At(i)
				data.TypeParams = append(data.TypeParams, ParamData{
					Name: tp.Obj().Name(),
					Type: tp.Constraint(), // Store types.Type directly
				})
				collectImports(data, tp.Constraint())
			}
		}

		for i := 0; i < ifaceType.NumMethods(); i++ {
			method := ifaceType.Method(i)
			sig := method.Type().(*types.Signature)

			methodData := MethodData{
				Name: method.Name(),
			}

			// Parameters
			if sig.Params() != nil {
				for j := 0; j < sig.Params().Len(); j++ {
					param := sig.Params().At(j)
					methodData.Params = append(methodData.Params, ParamData{
						Name: param.Name(),
						Type: param.Type(), // Store types.Type directly
					})
					collectImports(data, param.Type())
				}
			}

			// Results
			if sig.Results() != nil {
				for j := 0; j < sig.Results().Len(); j++ {
					result := sig.Results().At(j)
					name := result.Name()
					
					methodData.Results = append(methodData.Results, ResultData{
						Name: name,
						Type: result.Type(), // Store types.Type directly
					})
					collectImports(data, result.Type())
				}
			}
			data.Methods = append(data.Methods, methodData)
		}
		foundInterface = data
	}

	if foundInterface == nil {
		return nil, fmt.Errorf("interface %s not found", interfaceName)
	}

	return foundInterface, nil
}
