package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"strings"

	"github.com/phildrip/toe/options"
)

// typeToExpr converts a types.Type to an ast.Expr, handling package imports.
func typeToExpr(t types.Type, currentPackageName string, imports map[string]string) ast.Expr {
	switch typ := t.(type) {
	case *types.Basic:
		return ast.NewIdent(typ.Name())
	case *types.Named:
		// If the named type belongs to an external package, use a selector expression
		if typ.Obj().Pkg() != nil && typ.Obj().Pkg().Path() != currentPackageName {
			// Check if we collected an alias for this package
			pkgName, ok := imports[typ.Obj().Pkg().Path()]
			if !ok {
				pkgName = typ.Obj().Pkg().Name() // Fallback to actual package name
			}
			return &ast.SelectorExpr{
				X:   ast.NewIdent(pkgName),
				Sel: ast.NewIdent(typ.Obj().Name()),
			}
		}
		// Otherwise, it's a type in the current package or a predeclared type
		return ast.NewIdent(typ.Obj().Name())
	case *types.Pointer:
		return &ast.StarExpr{X: typeToExpr(typ.Elem(), currentPackageName, imports)}
	case *types.Slice:
		return &ast.ArrayType{Elt: typeToExpr(typ.Elem(), currentPackageName, imports)}
	case *types.Array:
		return &ast.ArrayType{Len: &ast.BasicLit{Kind: token.INT,
			Value: fmt.Sprintf("%d", typ.Len())},
			Elt: typeToExpr(typ.Elem(), currentPackageName, imports)}
	case *types.Map:
		return &ast.MapType{Key: typeToExpr(typ.Key(), currentPackageName, imports),
			Value: typeToExpr(typ.Elem(), currentPackageName, imports)}
	case *types.Chan:
		return &ast.ChanType{Dir: ast.ChanDir(typ.Dir()),
			Value: typeToExpr(typ.Elem(), currentPackageName, imports)}
	case *types.Signature:
		// This case is for function types (e.g., func(...) (...))
		params := &ast.FieldList{}
		if typ.Params() != nil {
			for i := 0; i < typ.Params().Len(); i++ {
				param := typ.Params().At(i)
				paramNames := []*ast.Ident{}
				if param.Name() != "" {
					paramNames = append(paramNames, ast.NewIdent(param.Name()))
				}
				params.List = append(params.List, &ast.Field{
					Names: paramNames,
					Type:  typeToExpr(param.Type(), currentPackageName, imports),
				})
			}
		}

		results := &ast.FieldList{}
		if typ.Results() != nil {
			for i := 0; i < typ.Results().Len(); i++ {
				result := typ.Results().At(i)
				resultNames := []*ast.Ident{}
				if result.Name() != "" {
					resultNames = append(resultNames, ast.NewIdent(result.Name()))
				}
				results.List = append(results.List, &ast.Field{
					Names: resultNames,
					Type:  typeToExpr(result.Type(), currentPackageName, imports),
				})
			}
		}
		return &ast.FuncType{Params: params, Results: results}
	case *types.Interface:
		// If it's the empty interface, return an empty interface literal
		if typ.Empty() {
			return &ast.InterfaceType{Methods: &ast.FieldList{}} // represents interface{}
		}
		// For other interfaces, fall back to its string representation
		return ast.NewIdent(typ.String())

	case *types.TypeParam:
		return ast.NewIdent(typ.Obj().Name())
	default:
		// Fallback for types not explicitly handled (e.g., complex structs, unexported types)
		return ast.NewIdent(typ.String())
	}
}

// getBaseTypeName returns a concise string representation of the base type.
func getBaseTypeName(t types.Type) string {
	switch typ := t.(type) {
	case *types.Basic:
		return typ.Name()
	case *types.Named:
		return typ.Obj().Name()
	case *types.Pointer:
		return getBaseTypeName(typ.Elem())
	case *types.Slice:
		return getBaseTypeName(typ.Elem())
	case *types.Array:
		return getBaseTypeName(typ.Elem())
	case *types.Map:
		return "Map" // Keep it simple for maps
	case *types.Chan:
		return getBaseTypeName(typ.Elem())
	case *types.Interface:
		if typ.Empty() {
			return "Interface" // For interface{}
		}
		return "Interface"
	case *types.TypeParam:
		return typ.Obj().Name()
	default:
		return "Any" // Fallback for unhandled types
	}
}

// parseStmt parses a string into an ast.Stmt. It assumes the string represents a single statement.
func parseStmt(stmtStr string) ast.Stmt {
	fset := token.NewFileSet()
	// Use a dummy filename and a package clause to make the parser happy.
	// We only care about the statement body.
	src := fmt.Sprintf("package p\nfunc _() { %s }", stmtStr)
	node, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		panic(fmt.Errorf("failed to parse statement: %w\n%s", err, stmtStr))
	}
	// The statement we want is inside the function body. The function body is a BlockStmt.
	// We expect only one statement inside the function body.
	if len(node.Decls) == 0 {
		panic(fmt.Errorf("no declarations found for statement: %s", stmtStr))
	}
	funcDecl, ok := node.Decls[0].(*ast.FuncDecl)
	if !ok || funcDecl.Body == nil || len(funcDecl.Body.List) == 0 || funcDecl.Body.List[0] == nil {
		panic(fmt.Errorf("no function body or statements found for statement: %s", stmtStr))
	}
	return funcDecl.Body.List[0]
}

func GenerateStubCode(ifaceData *InterfaceData, opts *options.StubOptions) (string, error) {
	// Create a new file set and AST file
	fset := token.NewFileSet()
	file := &ast.File{
		Name: ast.NewIdent(ifaceData.PackageName),
	}

	// Add collected imports
	importSpecs := []ast.Spec{ // Always import sync for the mutex
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"sync"`}},
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"github.com/phildrip/toe/options"`}},
	}
	for path, name := range ifaceData.Imports {
		var importName *ast.Ident
		// Only add name if it's different from the last part of the path or if it's explicitly needed
		lastSlash := strings.LastIndex(path, "/")
		lastPart := path
		if lastSlash != -1 {
			lastPart = path[lastSlash+1:]
		}
		if name != lastPart {
			importName = ast.NewIdent(name)
		}
		importSpecs = append(importSpecs, &ast.ImportSpec{
			Name: importName,
			Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", path)},
		})
	}
	file.Decls = append(file.Decls, &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: importSpecs,
	})

	// Create the stub struct definition
	stubName := "Stub" + ifaceData.Name
	stubStruct := &ast.TypeSpec{
		Name: ast.NewIdent(stubName),
		Type: &ast.StructType{
			Fields: &ast.FieldList{},
		},
	}

	// Add sync.Mutex and _isLocked fields
	stubStruct.Type.(*ast.StructType).Fields.List = append(
		stubStruct.Type.(*ast.StructType).Fields.List, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent("mu")},
			Type:  &ast.SelectorExpr{X: ast.NewIdent("sync"), Sel: ast.NewIdent("Mutex")},
		},
		&ast.Field{
			Names: []*ast.Ident{ast.NewIdent("isLocked")},
			Type:  ast.NewIdent("bool"),
		},
	)

	// Add fields for call recording and function stubs to the stub struct
	for _, method := range ifaceData.Methods {
		// Add MethodNameFunc field (for lambda stubbing)
		funcType := &ast.FuncType{}
		params := &ast.FieldList{}
		for _, p := range method.Params {
			params.List = append(params.List, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(p.Name)},
				Type:  typeToExpr(p.Type, ifaceData.PackageName, ifaceData.Imports),
			})
		}
		funcType.Params = params

		results := &ast.FieldList{}
		for _, r := range method.Results {
			results.List = append(results.List, &ast.Field{
				Type: typeToExpr(r.Type, ifaceData.PackageName, ifaceData.Imports),
			})
		}
		funcType.Results = results

		stubStruct.Type.(*ast.StructType).Fields.List = append(
			stubStruct.Type.(*ast.StructType).Fields.List, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(method.Name + "Func")},
				Type:  funcType,
			})

		// Add MethodNameCall struct type and its field
		callStructName := stubName + method.Name + "Call"
		callStruct := &ast.TypeSpec{
			Name: ast.NewIdent(callStructName),
			Type: &ast.StructType{
				Fields: &ast.FieldList{},
			},
		}

		// Add type parameters to call struct if the main struct is generic
		if len(ifaceData.TypeParams) > 0 {
			callStruct.TypeParams = &ast.FieldList{List: copyTypeParams(ifaceData.TypeParams,
				ifaceData.PackageName,
				ifaceData.Imports)}
		}

		for _, p := range method.Params {
			callStruct.Type.(*ast.StructType).Fields.List = append(
				callStruct.Type.(*ast.StructType).Fields.List, &ast.Field{
					Names: []*ast.Ident{ast.NewIdent(strings.Title(p.Name))},
					Type:  typeToExpr(p.Type, ifaceData.PackageName, ifaceData.Imports),
				})
		}

		file.Decls = append(file.Decls, &ast.GenDecl{
			Tok:   token.TYPE,
			Specs: []ast.Spec{callStruct},
		})

		// Add MethodNameCalls field (slice of callStructName[T])
		var callListType ast.Expr = ast.NewIdent(callStructName)
		if len(ifaceData.TypeParams) > 0 {
			var typeArgs []ast.Expr
			for _, tp := range ifaceData.TypeParams {
				typeArgs = append(typeArgs, ast.NewIdent(tp.Name))
			}

			if len(typeArgs) == 1 {
				callListType = &ast.IndexExpr{X: ast.NewIdent(callStructName), Index: typeArgs[0]}
			} else {
				callListType = &ast.IndexListExpr{X: ast.NewIdent(callStructName),
					Indices: typeArgs}
			}
		}

		stubStruct.Type.(*ast.StructType).Fields.List = append(
			stubStruct.Type.(*ast.StructType).Fields.List, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(method.Name + "Calls")},
				Type:  &ast.ArrayType{Elt: callListType},
			})

		// Add MethodNameReturns struct type and its field
		if len(method.Results) > 0 {
			returnsStructName := stubName + method.Name + "Returns"
			returnsStruct := &ast.TypeSpec{
				Name: ast.NewIdent(returnsStructName),
				Type: &ast.StructType{
					Fields: &ast.FieldList{},
				},
			}

			// Add type parameters to returns struct if the main struct is generic
			if len(ifaceData.TypeParams) > 0 {
				returnsStruct.TypeParams = &ast.FieldList{List: copyTypeParams(ifaceData.TypeParams,
					ifaceData.PackageName,
					ifaceData.Imports)}
			}

			for i, res := range method.Results {
				var fieldName string
				if res.Name != "" {
					fieldName = strings.Title(res.Name)
				} else {
					baseTypeName := getBaseTypeName(res.Type)
					fieldName = strings.Title(baseTypeName) + fmt.Sprintf("%d", i)
				}
				returnsStruct.Type.(*ast.StructType).Fields.List = append(
					returnsStruct.Type.(*ast.StructType).Fields.List, &ast.Field{
						Names: []*ast.Ident{ast.NewIdent(fieldName)},
						Type:  typeToExpr(res.Type, ifaceData.PackageName, ifaceData.Imports),
					})
			}

			file.Decls = append(file.Decls, &ast.GenDecl{
				Tok:   token.TYPE,
				Specs: []ast.Spec{returnsStruct},
			})

			// Add MethodNameReturns field to the stub struct
			var returnsFieldType ast.Expr = ast.NewIdent(returnsStructName)
			if len(ifaceData.TypeParams) > 0 {
				var typeArgs []ast.Expr
				for _, tp := range ifaceData.TypeParams {
					typeArgs = append(typeArgs, ast.NewIdent(tp.Name))
				}

				if len(typeArgs) == 1 {
					returnsFieldType = &ast.IndexExpr{X: ast.NewIdent(returnsStructName), Index: typeArgs[0]}
				} else {
					returnsFieldType = &ast.IndexListExpr{X: ast.NewIdent(returnsStructName),
						Indices: typeArgs}
				}
			}

			stubStruct.Type.(*ast.StructType).Fields.List = append(
				stubStruct.Type.(*ast.StructType).Fields.List, &ast.Field{
					Names: []*ast.Ident{ast.NewIdent(method.Name + "Returns")},
					Type: returnsFieldType,
				})

		}

	}

	// Add type parameters for generic interfaces
	if len(ifaceData.TypeParams) > 0 {
		fields := make([]*ast.Field, len(ifaceData.TypeParams))
		for i, tp := range ifaceData.TypeParams {
			// Determine the AST representation of the type parameter's constraint
			var constraintType ast.Expr
			if tp.Type.String() == "interface{}" { // Now using .String() on types.Type
				constraintType = &ast.InterfaceType{Methods: &ast.FieldList{}} // Represents 'interface{}'
			} else {
				constraintType = typeToExpr(tp.Type, ifaceData.PackageName, ifaceData.Imports)
			}

			fields[i] = &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(tp.Name)},
				Type:  constraintType,
			}
		}
		stubStruct.TypeParams = &ast.FieldList{List: fields}
	}

	file.Decls = append(file.Decls, &ast.GenDecl{ // Changed from decls = append(decls, ...)
		Tok:   token.TYPE,
		Specs: []ast.Spec{stubStruct},
	})

	// Create constructor
	file.Decls = append(file.Decls,
		createConstructor(stubName, ifaceData.TypeParams, ifaceData.PackageName, ifaceData.Imports, opts))

	// Create methods for the stub struct
	for _, method := range ifaceData.Methods {
		file.Decls = append(file.Decls,
			createMethod(stubName,
				method,
				ifaceData.TypeParams,
				ifaceData.PackageName,
				ifaceData.Imports,
				opts))
	}

	// Generate the code
	var buf strings.Builder
	if err := format.Node(&buf, fset, file); err != nil {
		return "", fmt.Errorf("error formatting generated code: %v", err)
	}

	return buf.String(), nil
}

// copyTypeParams creates a new slice of ast.Field representing type parameters.
// It's used to safely copy type parameters for nested generic structs.
func copyTypeParams(params []ParamData,
	currentPackageName string,
	imports map[string]string) []*ast.Field {
	copied := make([]*ast.Field, len(params))
	for i, p := range params {
		var constraintType ast.Expr
		if p.Type.String() == "interface{}" {
			constraintType = &ast.InterfaceType{Methods: &ast.FieldList{}}
		} else {
			constraintType = typeToExpr(p.Type, currentPackageName, imports) // Use typeToExpr here
		}
		copied[i] = &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(p.Name)},
			Type:  constraintType,
		}
	}
	return copied
}

func createConstructor(stubName string,
	typeParams []ParamData,
	currentPackageName string,
	imports map[string]string,
	opts *options.StubOptions) *ast.FuncDecl {
	constructorName := "New" + stubName

	// Build receiver type for the constructor
	var resultType ast.Expr = ast.NewIdent(stubName)

	// Type parameters for the constructor function itself
	var funcTypeParams *ast.FieldList
	if len(typeParams) > 0 {
		funcTypeParams = &ast.FieldList{List: copyTypeParams(typeParams,
			currentPackageName,
			imports)}

		var typeArgs []ast.Expr
		for _, tp := range typeParams {
			typeArgs = append(typeArgs, ast.NewIdent(tp.Name))
		}
		if len(typeArgs) == 1 {
			resultType = &ast.IndexExpr{X: ast.NewIdent(stubName), Index: typeArgs[0]}
		} else {
			resultType = &ast.IndexListExpr{X: ast.NewIdent(stubName), Indices: typeArgs}
		}
	}

	return &ast.FuncDecl{
		Name: ast.NewIdent(constructorName),
		Type: &ast.FuncType{
			TypeParams: funcTypeParams, // Add type parameters to the function declaration
			Params: &ast.FieldList{List: []*ast.Field{{
				Names: []*ast.Ident{ast.NewIdent("opts")},
				Type:  &ast.SelectorExpr{X: ast.NewIdent("options"), Sel: ast.NewIdent("StubOptions")},
			}}},
			Results:    &ast.FieldList{List: []*ast.Field{{Type: &ast.StarExpr{X: resultType}}}},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.UnaryExpr{
							Op: token.AND,
							X: &ast.CompositeLit{
								Type: resultType,
								Elts: []ast.Expr{
									&ast.KeyValueExpr{Key: ast.NewIdent("isLocked"), Value: ast.NewIdent("opts.WithLocking")},
								},
							},
						},
					},
				},
			},
		},
	}
}

func createMethod(stubName string,
	method MethodData,
	typeParams []ParamData,
	currentPackageName string,
	imports map[string]string,
	opts *options.StubOptions) *ast.FuncDecl {
	// Method receiver
	recv := &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent("s")},
				Type:  &ast.StarExpr{X: ast.NewIdent(stubName)},
			},
		},
	}

	// If the struct is generic, build the full receiver type e.g., *StubName[T]
	if len(typeParams) > 0 {
		var typeArgs []ast.Expr
		for _, tp := range typeParams {
			typeArgs = append(typeArgs, ast.NewIdent(tp.Name))
		}

		// Correctly specify the receiver type in the generated method
		if len(typeArgs) == 1 {
			recv.List[0].Type = &ast.StarExpr{
				X: &ast.IndexExpr{
					X:     ast.NewIdent(stubName),
					Index: typeArgs[0],
				},
			}
		} else {
			recv.List[0].Type = &ast.StarExpr{
				X: &ast.IndexListExpr{
					X:       ast.NewIdent(stubName),
					Indices: typeArgs,
				},
			}
		}
	}

	// Method parameters
	params := &ast.FieldList{}
	for _, p := range method.Params {
		params.List = append(params.List, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(p.Name)},
			Type:  typeToExpr(p.Type, currentPackageName, imports),
		})
	}

	// Method results
	results := &ast.FieldList{}
	for _, r := range method.Results {
		results.List = append(results.List, &ast.Field{
			Type: typeToExpr(r.Type, currentPackageName, imports),
		})
	}

	// Method body
	var bodyStmts []ast.Stmt

	// Add conditional locking. The sync import will now always be used.
	bodyStmts = append(bodyStmts, parseStmt(fmt.Sprintf(`
	if s.isLocked {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	`)))

	// Add call recording
	callStructName := stubName + method.Name + "Call"
	var callElts []ast.Expr
	for _, p := range method.Params {
		callElts = append(callElts, &ast.KeyValueExpr{
			Key:   ast.NewIdent(strings.Title(p.Name)), // Capitalize key for public field
			Value: ast.NewIdent(p.Name),
		})
	}

	// The type of the call instance needs to be generic if the interface is generic
	var callInstanceType ast.Expr = ast.NewIdent(callStructName)
	if len(typeParams) > 0 {
		var typeArgs []ast.Expr
		for _, tp := range typeParams {
			typeArgs = append(typeArgs, ast.NewIdent(tp.Name))
		}

		if len(typeArgs) == 1 {
			callInstanceType = &ast.IndexExpr{X: ast.NewIdent(callStructName), Index: typeArgs[0]}
		} else {
			callInstanceType = &ast.IndexListExpr{X: ast.NewIdent(callStructName),
				Indices: typeArgs}
		}
	}

	callInstance := &ast.CompositeLit{
		Type: callInstanceType,
		Elts: callElts,
	}

	// Assign the result of append back to the slice
	bodyStmts = append(bodyStmts, &ast.AssignStmt{
		Tok: token.ASSIGN,
		Lhs: []ast.Expr{&ast.SelectorExpr{X: ast.NewIdent("s"),
			Sel: ast.NewIdent(method.Name + "Calls")}},
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun: ast.NewIdent("append"),
			Args: []ast.Expr{
				&ast.SelectorExpr{X: ast.NewIdent("s"), Sel: ast.NewIdent(method.Name + "Calls")},
				callInstance,
			},
		}},
	})

	// Handle return values
	if len(method.Results) > 0 { // Only if the method has return values
		// If MethodNameFunc is set, use it.
		funcName := method.Name + "Func"
		returnsName := method.Name + "Returns"
		// Generate string for MethodNameFunc call args
		var funcCallArgs []string
		for _, p := range method.Params {
			funcCallArgs = append(funcCallArgs, p.Name)
		}
		funcCallArgsStr := strings.Join(funcCallArgs, ", ")

		// Construct the if-else logic as a string and parse it into an AST statement
		var returnValues []string
		for i, r := range method.Results {
			var fieldName string
			if r.Name != "" {
				fieldName = strings.Title(r.Name)
			} else {
				baseTypeName := getBaseTypeName(r.Type)
				fieldName = strings.Title(baseTypeName) + fmt.Sprintf("%d", i)
			}
			returnValues = append(returnValues, fmt.Sprintf("s.%s.%s", returnsName, fieldName))
		}
		returnValuesStr := strings.Join(returnValues, ", ")

		returnLogicStr := fmt.Sprintf(`
		if s.%s != nil {
			return s.%s(%s)
		} else {
			return %s
		}
		`,
			funcName,
			funcName,
			funcCallArgsStr,
			returnValuesStr)

		bodyStmts = append(bodyStmts, parseStmt(returnLogicStr))

	} else { // No return values in method signature, just add a simple return
		bodyStmts = append(bodyStmts, &ast.ReturnStmt{})
	}

	tbody := &ast.BlockStmt{
		List: bodyStmts,
	}

	return &ast.FuncDecl{
		Recv: recv,
		Name: ast.NewIdent(method.Name),
		Type: &ast.FuncType{
			Params:  params,
			Results: results,
		},
		Body: tbody,
	}
}
