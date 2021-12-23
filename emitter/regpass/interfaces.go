package regpass

import (
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/core"
)

func (ctx *Context) registerInterface(i *ast.InterfaceDefinition) error {
	members := make([]*core.InterfaceMember, len(i.Members))
	for i, m := range i.Members {
		var member *core.InterfaceMember
		if m.Field != nil {
			t, err := ctx.typeSpecFromAst(&m.Field.Type)
			if err != nil {
				return err
			}

			member = &core.InterfaceMember{
				Field: &core.InterfaceField{
					Name: m.Field.Name,
					Type: t,
				},
			}
		} else if m.Method != nil {
			t, err := ctx.typeSpecFromAst(m.Method.ReturnType)
			if err != nil {
				return err
			}

			member = &core.InterfaceMember{
				Method: &core.InterfaceMethod{
					Name:       m.Field.Name,
					ReturnType: t,
				},
			}
		}

		members[i] = member
	}

	ctx.currentScope().AddType(&core.TypeSpec{
		Interface: &core.Interface{
			Name:    i.Name,
			Members: members,
		},
	})

	return nil
}
