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
	return fmt.Sprintf("ts_fn_%s", name)
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
	// TODO -- need to actually impl!

	if expr.String != nil {
		t := string(TsString)
		return &ast.Type{NonUnionType: &ast.NonUnionType{TypeReference: &t}}, nil
	}

	if expr.Number != nil {
		t := string(TsNumber)
		return &ast.Type{NonUnionType: &ast.NonUnionType{TypeReference: &t}}, nil
	}

	if expr.Ident != nil {
		return ctx.TypeOf(*expr.Ident)
	}

	if expr.FunctionInstantiation != nil {
		return &ast.Type{
			NonUnionType: &ast.NonUnionType{
				LiteralType: &ast.LiteralType{
					FunctionType: &ast.FunctionType{
						Parameters: expr.FunctionInstantiation.Parameters,
						ReturnType: expr.FunctionInstantiation.ReturnType,
					},
				},
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

	if expr.ChainedObjectOperation != nil {
		_, tail, err := buildOperationChain(ctx, expr.ChainedObjectOperation)
		if err != nil {
			return nil, err
		}

		return tail.accesseeType, nil
	}

	return nil, fmt.Errorf("could not infer type of %#v", *expr)
}
