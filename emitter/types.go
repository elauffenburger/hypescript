package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
)

func mangleFunctionName(name string) string {
	return name
}

func mangleIdentName(name string, identType *TypeSpec) string {
	if identType.Function != nil {
		return mangleFunctionName(name)
	}

	return name
}

func inferType(ctx *Context, expr *Expression) (*TypeSpec, error) {
	// TODO -- need to actually impl!

	if expr.String != nil {
		t := string(TsString)

		return &TypeSpec{TypeReference: &t}, nil
	}

	if expr.Number != nil {
		t := string(TsNumber)

		return &TypeSpec{TypeReference: &t}, nil
	}

	if expr.Ident != nil {
		return ctx.CurrentScope.TypeOf(*expr.Ident)
	}

	if fn := expr.FunctionInstantiation; fn != nil {
		return &TypeSpec{Function: fn}, nil
	}

	if objInst := expr.ObjectInstantiation; objInst != nil {
		fields := make([]*ObjectTypeField, len(objInst.Fields))
		for i, f := range objInst.Fields {
			fields[i] = &ObjectTypeField{
				Name: f.Name,
				Type: f.Type,
			}
		}

		return &TypeSpec{Object: &Object{Fields: fields}}, nil
	}

	if chain := expr.ChainedObjectOperation; chain != nil {
		return chain.Last.Accessee.Type, nil
	}

	return nil, fmt.Errorf("unable to infer type")
}

func inferTypeFromAst(ctx *Context, astExpr *ast.Expression) (*TypeSpec, error) {
	// TODO -- need to actually impl!

	if astExpr.String != nil {
		t := string(TsString)

		return &TypeSpec{TypeReference: &t}, nil
	}

	if astExpr.Number != nil {
		t := string(TsNumber)

		return &TypeSpec{TypeReference: &t}, nil
	}

	if astExpr.Ident != nil {
		return ctx.CurrentScope.TypeOf(*astExpr.Ident)
	}

	if fnInst := astExpr.FunctionInstantiation; fnInst != nil {
		fn, err := functionFromAst(ctx, fnInst)
		if err != nil {
			return nil, err
		}

		return &TypeSpec{Function: fn}, nil
	}

	if objInst := astExpr.ObjectInstantiation; objInst != nil {
		obj, err := objectFromInstAst(ctx, objInst.Fields)
		if err != nil {
			return nil, err
		}

		return &TypeSpec{Object: obj}, nil
	}

	if astExpr.ChainedObjectOperation != nil {
		chain, err := chainedObjOperationFromAst(ctx, astExpr.ChainedObjectOperation)
		if err != nil {
			return nil, err
		}

		return chain.Last.Accessee.Type, nil
	}

	return nil, fmt.Errorf("unable to infer type")
}

func objectFromInstAst(ctx *Context, fields []*ast.ObjectFieldInstantiation) (*Object, error) {
	objFields := make([]*ObjectTypeField, len(fields))

	for i, f := range fields {
		t, err := inferTypeFromAst(ctx, &f.Value)
		if err != nil {
			return nil, err
		}

		objFields[i] = &ObjectTypeField{Name: f.Name, Type: t}
	}

	return &Object{Fields: objFields}, nil
}

func objectInstFromAst(ctx *Context, fields []*ast.ObjectFieldInstantiation) (*ObjectInstantiation, error) {
	objFields := make([]*ObjectFieldInstantiation, len(fields))

	for i, f := range fields {
		t, err := inferTypeFromAst(ctx, &f.Value)
		if err != nil {
			return nil, err
		}

		expr, err := expressionFromAst(ctx, &f.Value)
		if err != nil {
			return nil, err
		}

		objFields[i] = &ObjectFieldInstantiation{Name: f.Name, Type: t, Value: expr}
	}

	return &ObjectInstantiation{Fields: objFields}, nil
}

func objectFromAst(ctx *Context, fields []*ast.ObjectTypeField) (*Object, error) {
	objFields := make([]*ObjectTypeField, len(fields))

	for i, field := range fields {
		fieldType, err := fromAstTypeIdentifier(ctx, &field.Type)
		if err != nil {
			return nil, err
		}

		objFields[i] = &ObjectTypeField{
			Name: field.Name,
			Type: fieldType,
		}
	}

	return &Object{Fields: objFields}, nil
}

func functionFromAst(ctx *Context, fn *ast.FunctionInstantiation) (*Function, error) {
	res, err := ctx.WithinTempScope(func() (interface{}, error) {
		return ctx.registerFunctionDeclaration(fn)
	})

	if err != nil {
		return nil, err
	}

	return res.(*StatementOrExpression).Statement.FunctionInstantiation, nil
}

func fromAstTypeIdentifier(ctx *Context, t *ast.TypeIdentifier) (*TypeSpec, error) {
	if t == nil {
		return nil, nil
	}

	if t := t.NonUnionType; t != nil {
		if t := t.LiteralType; t != nil {
			if t := t.FunctionType; t != nil {
				fn, err := functionFromAst(ctx, &ast.FunctionInstantiation{
					Parameters: t.Parameters,
					ReturnType: t.ReturnType,
				})

				if err != nil {
					return nil, err
				}

				return &TypeSpec{Function: fn}, nil
			}

			if t.ObjectType != nil {
				obj, err := objectFromAst(ctx, t.ObjectType.Fields)
				if err != nil {
					return nil, err
				}

				return &TypeSpec{Object: obj}, nil
			}
		}

		if t := t.TypeReference; t != nil {
			return &TypeSpec{TypeReference: t}, nil
		}
	}

	return nil, fmt.Errorf("unknown type identifier %#v", t)
}

func createUnionType(left, right *TypeSpec) (*TypeSpec, error) {
	return &TypeSpec{Union: &Union{Head: left, Tail: []*TypeSpec{right}}}, nil
}

func strRef(str string) *string {
	return &str
}
