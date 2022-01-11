package regpass

import (
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/core"
	"elauffenburger/hypescript/typeutils"
)

func (ctx *Context) registerInterface(i *ast.InterfaceDefinition) error {
	members := make(map[string]*core.Member, len(i.Members))
	for _, m := range i.Members {
		switch {
		case m.Field != nil:
			t, err := ctx.typeSpecFromAst(&m.Field.Type)
			if err != nil {
				return err
			}

			members[m.Field.Name] = &core.Member{
				Field: &core.ObjectTypeField{
					Name:     m.Field.Name,
					Optional: m.Field.Optional,
					Type:     t,
				},
			}
		case m.Method != nil:
			t, err := ctx.typeSpecFromAst(m.Method.ReturnType)
			if err != nil {
				return err
			}

			params := make([]*core.FunctionParameter, len(m.Method.Parameters))
			for i, p := range m.Method.Parameters {
				paramType, err := ctx.typeSpecFromAst(&p.Type)
				if err != nil {
					return err
				}

				params[i] = &core.FunctionParameter{
					Name:     p.Name,
					Optional: p.Optional,
					Type:     paramType,
				}
			}

			members[m.Method.Name] = &core.Member{
				Function: &core.Function{
					Name:               typeutils.StrRef(m.Method.Name),
					Parameters:         params,
					ExplicitReturnType: t,
					ImplicitReturnType: t,
				},
			}
		}
	}

	ctx.currentScope().AddType(&core.TypeSpec{
		Interface: &core.Interface{
			Name:    i.Name,
			Members: members,
		},
	})

	return nil
}
