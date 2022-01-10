package writepass

import "elauffenburger/hypescript/emitter/core"

func (ctx *Context) writeIdentAssignment(assign *core.IdentAssignment) error {
	err := ctx.writeIdent(assign.Ident)
	if err != nil {
		return err
	}

	ctx.WriteString(" = ")

	return ctx.writeExpression(assign.Assignment.Value)
}

func (ctx *Context) writeIdent(ident string) error {
	identType, err := ctx.currentScope().IdentType(ident)

	// If we couldn't find the type of the ident, don't try mangling it --
	// it's either not defined yet or it's a bug we'll catch later.
	if err != nil {
		ctx.WriteString(ident)
	} else {
		// Otherwise, mangle away!
		ctx.WriteString(mangleIdentName(ident, identType))
	}

	return nil
}
