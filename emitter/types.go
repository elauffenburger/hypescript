package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
)

func mangleFunctionName(name string) string {
	return fmt.Sprintf("%s", name)
}

func mangleIdentName(name string, identType *TypeSpec) string {
	if identType.FunctionType != nil {
		return mangleFunctionName(name)
	}

	return name
}

func inferType(ctx *Context, expr *ast.Expression) (*TypeSpec, error) {
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
		return ctx.TypeOf(*expr.Ident)
	}

	if expr.FunctionInstantiation != nil {
		return &TypeSpec{
			FunctionType: &ast.FunctionType{
				Parameters: expr.FunctionInstantiation.Parameters,
				ReturnType: expr.FunctionInstantiation.ReturnType,
			},
		}, nil
	}

	if expr.ObjectInstantiation != nil {
		fields := make([]ast.ObjectTypeField, len(expr.ObjectInstantiation.Fields))
		for i, field := range expr.ObjectInstantiation.Fields {
			fieldType, err := inferType(ctx, &field.Value)
			if err != nil {
				return nil, err
			}

			fieldTypeIdent, err := fieldType.toAstTypeIdentifier()
			if err != nil {
				return nil, err
			}

			fields[i] = ast.ObjectTypeField{
				Name: field.Name,
				Type: *fieldTypeIdent,
			}
		}

		return &TypeSpec{
			ObjectType: &ast.ObjectType{
				Fields: fields,
			},
		}, nil
	}

	if expr.ChainedObjectOperation != nil {
		_, tail, err := buildOperationChain(ctx, expr.ChainedObjectOperation)
		if err != nil {
			return nil, err
		}

		return tail.accesseeType, nil
	}

	return nil, fmt.Errorf("could not infer type of %#v", *expr)
}

func (t *TypeSpec) toAstTypeIdentifier() (*ast.TypeIdentifier, error) {
	// TODO: handle errors during mapping.
	return &ast.TypeIdentifier{
		NonUnionType: &ast.NonUnionType{
			LiteralType: &ast.LiteralType{
				FunctionType: t.FunctionType,
				ObjectType:   t.ObjectType,
			},
			TypeReference: t.TypeReference,
		},
		UnionType: t.UnionType,
	}, nil
}

func fromAstTypeIdentifier(t *ast.TypeIdentifier) (*TypeSpec, error) {
	if t == nil {
		return nil, nil
	}

	if t := t.NonUnionType; t != nil {
		if t := t.LiteralType; t != nil {
			if t.FunctionType != nil {
				return &TypeSpec{FunctionType: t.FunctionType}, nil
			}

			if t.ObjectType != nil {
				return &TypeSpec{ObjectType: t.ObjectType}, nil
			}
		}

		if t := t.TypeReference; t != nil {
			return &TypeSpec{TypeReference: t}, nil
		}
	}

	return nil, fmt.Errorf("unknown type identifier %#v", t)
}

func createUnionType(left, right *TypeSpec) (*TypeSpec, error) {
	leftT, err := left.toAstTypeIdentifier()
	if err != nil {
		return nil, err
	}

	rightT, err := right.toAstTypeIdentifier()
	if err != nil {
		return nil, err
	}

	return &TypeSpec{UnionType: ast.CreateUnionType(leftT, rightT)}, nil
}

func strRef(str string) *string {
	return &str
}
