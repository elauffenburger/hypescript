package writepass

import (
	"elauffenburger/hypescript/emitter/core"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

func (ctx *Context) writeObjectOperation(op *core.ObjectOperation) error {
	if access := op.Access; access != nil {
		// TODO: this won't always work!
		ctx.WriteString(fmt.Sprintf("->getFieldValue(\"%s\")", access.AccessedIdent))

		if op.Next != nil {
			return ctx.writeObjectOperation(op.Next)
		}

		return nil
	}

	if op.Invocation != nil {
		err := ctx.writeObjectInvocation(op.Accessee.Type, op.Invocation)
		if err != nil {
			return err
		}

		if op.Next != nil {
			return ctx.writeObjectOperation(op.Next)
		}

		return nil
	}

	if op.Assignment != nil {
		ctx.WriteString(fmt.Sprintf("->setFieldValue(\"%s\", ", *op.Accessee.Ident))
		ctx.writeExpression(op.Assignment.Value)
		ctx.WriteString(")")

		if op.Next != nil {
			return fmt.Errorf("unexpected next operation after assignment on op %#v", op)
		}

		return nil
	}

	return fmt.Errorf("unknown operation in chained object operation: %#v", op)
}

func (ctx *Context) writeObjectInstantiation(objInst *core.ObjectInstantiation) error {
	formattedFields := strings.Builder{}
	for i, f := range objInst.Fields {
		name := f.Name

		typeId, err := ctx.currentScope().GetTypeIdFor(f.Type)
		if err != nil {
			return err
		}

		descriptor := fmt.Sprintf("TsObjectFieldDescriptor(TsString(\"%s\"), %d)", name, typeId)

		value, err := ctx.WithinPrintContext(func(printCtx *Context) error {
			return printCtx.writeExpression(f.Value)
		})
		if err != nil {
			return err
		}

		formattedFields.WriteString(fmt.Sprintf("new TsObjectField(%s, %s)", descriptor, value))

		if i != len(objInst.Fields)-1 {
			formattedFields.WriteString(", ")
		}
	}

	ctx.WriteString(
		fmt.Sprintf(
			"new TsObject(%d, TsCoreHelpers::toVector<TsObjectField*>({%s}))",
			core.TypeIdTsObject,
			formattedFields.String(),
		),
	)

	return nil
}

func (ctx *Context) writeObjectInvocation(accesseeType *core.TypeSpec, invocation *core.ObjectInvocation) error {
	// TODO: this isn't always true!
	fn := accesseeType.Function

	// Write the args.
	args := strings.Builder{}
	numParams := len(fn.Parameters)
	for i, param := range fn.Parameters {
		arg := invocation.Arguments[i]

		expr, err := ctx.WithinPrintContext(func(printCtx *Context) error {
			return printCtx.writeExpression(arg)
		})
		if err != nil {
			return err
		}

		args.WriteString(fmt.Sprintf("TsFunctionArg(\"%s\", %s)", param.Name, expr))

		if i != numParams-1 {
			args.WriteString(", ")
		}
	}

	ctx.WriteString(fmt.Sprintf("->invoke(TsCoreHelpers::toVector<TsFunctionArg>({%s}))", args.String()))

	return nil
}

func (ctx *Context) writeChainedObjectOperation(op *core.ChainedObjectOperation) error {
	// TODO: this isn't always true!
	ctx.writeIdent(*op.First.Accessee.Ident)

	if err := ctx.rectifyChainedOperation(op); err != nil {
		return err
	}

	return ctx.writeObjectOperation(op.First)
}

func (ctx *Context) rectifyChainedOperation(op *core.ChainedObjectOperation) error {
	// TODO: this won't always be true!
	// Figure out the type of the base object.
	t, err := ctx.currentScope().ResolveIdentType(*op.First.Accessee.Ident)
	if err != nil {
		return err
	}

	// Attach the type.
	op.First.Accessee.Type = t

	currentOp := op.First
	for currentOp != nil {
		if currentOp.Access != nil {
			if currentOp.Next == nil {
				break
			}

			// Figure out the type of the field.
			t, err = ctx.getFieldType(currentOp.Accessee.Type, currentOp.Access.AccessedIdent)
			if err != nil {
				return err
			}

			// Append the type for ourselves on the next chained op.
			currentOp.Next.Accessee.Type = t
		}

		if currentOp.Invocation != nil {
			if prev := currentOp.Prev; prev != nil {
				// If the previous op was an access, move the type over.
				if prev.Access != nil {
					t, err = ctx.getFieldType(currentOp.Prev.Accessee.Type, *currentOp.Invocation.Accessee.Ident)
					if err != nil {
						return err
					}

					currentOp.Accessee.Type = t
				}

				// If the previous op was an invocation, grab the return type of the invoked
				// object.
				if prev.Invocation != nil {
					// TODO: this won't always be true!
					t := currentOp.Prev.Invocation.Accessee.Type.Function.ImplicitReturnType

					currentOp.Invocation.Accessee.Type = t
				}
			}
		}

		currentOp = currentOp.Next
	}

	return nil
}

func (ctx *Context) getFieldType(t *core.TypeSpec, field string) (*core.TypeSpec, error) {
	if t.Object != nil || t.Interface != nil {
		var hasMembers core.HasMembers
		if t.Object != nil {
			hasMembers = t.Object
		} else if t.Interface != nil {
			hasMembers = t.Interface
		}

		return hasMembers.AllMembers()[field].Type(), nil
	}

	return nil, errors.WithStack(fmt.Errorf("failed to resolve field type of %s on %s", field, t))
}
