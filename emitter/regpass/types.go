package regpass

import (
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/core"
	"fmt"
)

func (ctx *Context) objectFromAst(fields []*ast.ObjectTypeField) (*core.Object, error) {
	objFields := make(map[string]*core.ObjectTypeField, len(fields))

	for _, field := range fields {
		fieldType, err := ctx.typeSpecFromAst(&field.Type)
		if err != nil {
			return nil, err
		}

		objFields[field.Name] = &core.ObjectTypeField{
			Name: field.Name,
			Type: fieldType,
		}
	}

	return &core.Object{Fields: objFields}, nil
}

func (ctx *Context) functionFromAst(fn *ast.FunctionInstantiation) (*core.Function, error) {
	res, err := ctx.WithinTempScope(func() (interface{}, error) {
		return ctx.registerFunctionDeclaration(fn)
	})
	if err != nil {
		return nil, err
	}

	return res.(*core.StatementOrExpression).Statement.FunctionInstantiation, nil
}

func (ctx *Context) typeSpecFromAst(ident *ast.TypeIdentifier) (*core.TypeSpec, error) {
	if ident == nil {
		return nil, nil
	}

	if len(ident.Rest) > 0 {
		types := make(map[*core.TypeSpec]bool, len(ident.Rest)+1)

		for _, ut := range append([]*ast.TypeIdentifierPart{{Type: ident.Head}}, ident.Rest...) {
			spec, err := ctx.typeSpecFromAst(&ast.TypeIdentifier{Head: ut.Type})
			if err != nil {
				return nil, err
			}

			types[spec] = true
		}

		return &core.TypeSpec{Union: &core.Union{Types: types}}, nil
	}

	t := ident.Head
	if t := t.LiteralType; t != nil {
		if t := t.FunctionType; t != nil {
			fn, err := ctx.functionFromAst(&ast.FunctionInstantiation{
				Parameters: t.Parameters,
				ReturnType: t.ReturnType,
			})
			if err != nil {
				return nil, err
			}

			return &core.TypeSpec{Function: fn}, nil
		}

		if t.ObjectType != nil {
			obj, err := ctx.objectFromAst(t.ObjectType.Fields)
			if err != nil {
				return nil, err
			}

			return &core.TypeSpec{Object: obj}, nil
		}
	}

	if ref := t.TypeReference; ref != nil {
		t := ctx.currentScope().RegisteredType(*ref)

		return t, nil
	}

	return nil, fmt.Errorf("unknown type identifier %#v", ident)
}

func createUnionType(left, right *core.TypeSpec) (*core.TypeSpec, error) {
	return &core.TypeSpec{
		Union: &core.Union{
			Types: map[*core.TypeSpec]bool{
				left:  true,
				right: true,
			},
		},
	}, nil
}
