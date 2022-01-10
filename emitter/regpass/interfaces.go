package regpass

import (
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/core"
	"elauffenburger/hypescript/typeutils"
)

func (ctx *Context) registerInterface(i *ast.InterfaceDefinition) error {
	members := make(map[string]*core.Member, len(i.Members))
	for _, m := range i.Members {
		var member *core.Member
		if m.Field != nil {
			t, err := ctx.typeSpecFromAst(&m.Field.Type)
			if err != nil {
				return err
			}

			member = &core.Member{
				Name: m.Field.Name,
				Field: &core.ObjectTypeField{
					Name: m.Field.Name,
					Type: t,
				},
			}
		} else if m.Method != nil {
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

			member = &core.Member{
				Name: m.Method.Name,
				Function: &core.Function{
					Name:               typeutils.StrRef(m.Method.Name),
					Parameters:         params,
					ExplicitReturnType: t,
					ImplicitReturnType: t,
				},
			}
		}

		members[member.Name] = member
	}

	ctx.currentScope().AddType(&core.TypeSpec{
		Interface: &core.Interface{
			Name:    i.Name,
			Members: members,
		},
	})

	return nil
}
