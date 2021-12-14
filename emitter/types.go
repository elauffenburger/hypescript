package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
)

func mangleTypeName(name string) string {
	if name == string(TsCoreVoid) || name == string(TsCoreObj) {
		return name
	}

	return fmt.Sprintf("ts_%s*", name)
}

func mangleFunctionName(name string) string {
	return fmt.Sprintf("ts_%s", name)
}

func mangleIdentName(name string, identType *ast.Type) string {
	if identType.NonUnionType != nil {
		if identType.NonUnionType.LiteralType != nil {
			if identType.NonUnionType.LiteralType.FunctionType != nil {
				return mangleFunctionName(name)
			}
		}
	}

	return name
}

func inferType(ctx *Context, expr *ast.Expression) (*ast.Type, error) {
	// DO NOT SUBMIT -- need to actually impl!

	if expr.String != nil {
		t := string(TsString)
		return &ast.Type{NonUnionType: &ast.NonUnionType{TypeReference: &t}}, nil
	}

	if expr.Number != nil {
		t := string(TsNum)
		return &ast.Type{NonUnionType: &ast.NonUnionType{TypeReference: &t}}, nil
	}

	if expr.Ident != nil {
		return ctx.TypeOf(*expr.Ident)
	}

	if expr.ObjectInstantiation != nil {
		fields := make([]ast.ObjectTypeField, len(expr.ObjectInstantiation.Fields))
		for i, field := range expr.ObjectInstantiation.Fields {
			fieldType, err := inferType(ctx, &field.Value)
			if err != nil {
				return nil, err
			}

			fields[i] = ast.ObjectTypeField{
				Name: field.Name,
				Type: *fieldType,
			}
		}

		return &ast.Type{
			NonUnionType: &ast.NonUnionType{
				LiteralType: &ast.LiteralType{
					ObjectType: &ast.ObjectType{
						Fields: fields,
					},
				},
			},
		}, nil
	}

	return nil, fmt.Errorf("could not infer type of %#v", *expr)
}
