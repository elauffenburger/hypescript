package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
)

func isCoreType(s string) bool {
	for _, t := range coreTypes {
		if s == string(t) {
			return true
		}
	}

	return false
}

func mangleTypeNamePtr(name string) string {
	if name == string(TsVoid) {
		return name
	}

	if isCoreType(name) {
		return fmt.Sprintf("%s*", name)
	}

	return fmt.Sprintf("ts_%s*", name)
}

func mangleFunctionName(name string) string {
	return fmt.Sprintf("%s", name)
}

func mangleIdentName(name string, identType *TypeDefinition) string {
	if identType.FunctionType != nil {
		return mangleFunctionName(name)
	}

	return name
}

func inferType(ctx *Context, expr *ast.Expression) (*TypeDefinition, error) {
	// TODO -- need to actually impl!

	if expr.String != nil {
		t := string(TsString)

		return &TypeDefinition{TypeReference: &t}, nil
	}

	if expr.Number != nil {
		t := string(TsNumber)

		return &TypeDefinition{TypeReference: &t}, nil
	}

	if expr.Ident != nil {
		return ctx.TypeOf(*expr.Ident)
	}

	if expr.FunctionInstantiation != nil {
		return &TypeDefinition{
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

		return &TypeDefinition{
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

func (t *TypeDefinition) toAstTypeIdentifier() (*ast.TypeIdentifier, error) {
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

func fromAstTypeIdentifier(t *ast.TypeIdentifier) (*TypeDefinition, error) {
	if t == nil {
		return nil, nil
	}

	if t := t.NonUnionType; t != nil {
		if t := t.LiteralType; t != nil {
			if t.FunctionType != nil {
				return &TypeDefinition{FunctionType: t.FunctionType}, nil
			}

			if t.ObjectType != nil {
				return &TypeDefinition{ObjectType: t.ObjectType}, nil
			}
		}

		if t := t.TypeReference; t != nil {
			return &TypeDefinition{TypeReference: t}, nil
		}
	}

	return nil, fmt.Errorf("unknown type identifier %v", t)
}

func createUnionType(left, right *TypeDefinition) (*TypeDefinition, error) {
	leftT, err := left.toAstTypeIdentifier()
	if err != nil {
		return nil, err
	}

	rightT, err := right.toAstTypeIdentifier()
	if err != nil {
		return nil, err
	}

	return &TypeDefinition{UnionType: ast.CreateUnionType(leftT, rightT)}, nil
}
