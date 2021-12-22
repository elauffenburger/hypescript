package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
)

func validateInterface(ctx *Context, intdef *ast.InterfaceDefinition) error {
	if ctx.CurrentScope.ContainsIdent(intdef.Name) {
		return fmt.Errorf("existing ident %v in scope", intdef.Name)
	}

	for _, member := range intdef.Members {
		if member.Field != nil {
			t, err := fromAstTypeIdentifier(&member.Field.Type)
			if err != nil {
				return err
			}

			if err = ctx.CurrentScope.ValidateHasType(t); err != nil {
				return err
			}
		} else if m := member.Method; m != nil {
			for _, param := range m.Parameters {
				paramType, err := fromAstTypeIdentifier(&param.Type)
				if err != nil {
					return err
				}

				if err = ctx.CurrentScope.ValidateHasType(paramType); err != nil {
					return err
				}
			}

			rtnType, err := fromAstTypeIdentifier(m.ReturnType)
			if err != nil {
				return err
			}

			if err = ctx.CurrentScope.ValidateHasType(rtnType); err != nil {
				return err
			}
		}
	}

	return nil
}
