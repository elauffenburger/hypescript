package regpass

import (
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/core"
	"fmt"
)

func (ctx *Context) accessableFromAst(accessable ast.Accessable) (*core.Accessable, error) {
	if accessable.Ident != nil {
		t, err := ctx.currentScope().IdentType(*accessable.Ident)
		if err != nil {
			return nil, err
		}

		return &core.Accessable{Ident: accessable.Ident, Type: t}, nil
	}

	if lit := accessable.LiteralType; lit != nil {
		if f := lit.FunctionType; f != nil {
			fn, err := ctx.functionFromAst(&ast.FunctionInstantiation{
				Parameters: f.Parameters,
				ReturnType: f.ReturnType,
			})
			if err != nil {
				return nil, err
			}

			return &core.Accessable{Type: &core.TypeSpec{Function: fn}}, nil
		}

		if lit.ObjectType != nil {
			obj, err := ctx.objectFromAst(lit.ObjectType.Fields)
			if err != nil {
				return nil, err
			}

			return &core.Accessable{Type: &core.TypeSpec{Object: obj}}, nil
		}
	}

	return nil, fmt.Errorf("unknown accessable type: %v", accessable)
}
