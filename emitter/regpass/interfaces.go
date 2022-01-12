package regpass

import (
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/core"
	"elauffenburger/hypescript/typeutils"
)

func (ctx *Context) registerInterface(iface *ast.InterfaceDefinition) error {
	members := make([]*core.Member, len(iface.Members))
	for i, m := range iface.Members {
		switch {
		case m.Field != nil:
			t, err := ctx.typeSpecFromAst(&m.Field.Type)
			if err != nil {
				return err
			}

			if m.Field.Optional {
				union, err := createUnionType(t, &core.TypeSpec{TypeReference: typeutils.StrRef("undefined")})
				if err != nil {
					return err
				}

				t = union
			}

			members[i] = &core.Member{
				Field: &core.ObjectTypeField{
					Name: m.Field.Name,
					Type: t,
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

			name := m.Method.Name

			members[i] = &core.Member{
				Field: &core.ObjectTypeField{
					Name: name,
					Type: &core.TypeSpec{
						Function: &core.Function{
							Name:               &name,
							Parameters:         params,
							ExplicitReturnType: t,
							ImplicitReturnType: t,
						},
					},
				},
			}
		}
	}

	ctx.currentScope().AddType(&core.TypeSpec{
		Interface: core.NewInterface(
			iface.Name,
			members,
		),
	})

	return nil
}
