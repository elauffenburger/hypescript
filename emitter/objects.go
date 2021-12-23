package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

func chainedObjOperationFromAst(ctx *Context, chainedOp *ast.ChainedObjectOperation) (*ChainedObjectOperation, error) {
	var firstLink *ObjectOperation
	var currentLink *ObjectOperation

	accessee, err := fromAstAccessible(ctx, chainedOp.Accessee)
	if err != nil {
		return nil, err
	}

	for _, astOp := range chainedOp.Operations {
		// Create a new link.
		link := &ObjectOperation{
			Accessee: accessee,
			Prev:     currentLink,
		}

		if firstLink == nil {
			firstLink = link
		}

		// Link the current link (if any) to the new link.
		if currentLink != nil {
			currentLink.Next = link
		}

		// Make the current link the new link.
		currentLink = link

		// Add "access" chain operation.
		if access := astOp.Access; access != nil {
			link.Access = &ObjectAccess{
				AccessedIdent: access.AccessedIdent,
			}

			accessee = &Accessable{Ident: &access.AccessedIdent}

			continue
		}

		// Add "invocation" chain operation.
		if astInvoc := astOp.Invocation; astInvoc != nil {
			invoc, err := invocationFromAst(ctx, accessee, astInvoc)
			if err != nil {
				return nil, err
			}

			link.Invocation = invoc

			if accessee.Ident != nil {
				accessee = &Accessable{Ident: accessee.Ident}
			} else if t := accessee.Type; t != nil {
				if fn := accessee.Type.Function; fn != nil {
					accessee = &Accessable{Type: fn.ImplicitReturnType}
				} else {
					accessee = &Accessable{Type: t}
				}
			}

			continue
		}

		return nil, fmt.Errorf("unknown operation in chained object operation: %#v", astOp)
	}

	// Check if there's an assignment.
	if assign := chainedOp.Assignment; assign != nil {
		expr, err := expressionFromAst(ctx, &assign.Value)
		if err != nil {
			return nil, err
		}

		link := &ObjectOperation{
			Accessee: accessee,
			Assignment: &Assignment{
				Value: expr,
			},
		}

		// This is a little weird, but we want to remove the
		// previous link in the case of an assignment because
		// the last link would have _looked_ like an access, but
		// it's actually an assignment.
		//
		// i.e. `[access foo]->[access bar on foo]->[assign bar on bar "baz"]` becomes:
		//      `[access foo]->[assign bar on foo "baz"]`
		if currentLink != nil {
			link.Prev = currentLink.Prev
			currentLink.Prev.Next = link
		}

		currentLink = link
	}

	return &ChainedObjectOperation{First: firstLink, Last: currentLink}, nil
}

func writeObjectOperation(ctx *Context, op *ObjectOperation) error {
	if access := op.Access; access != nil {
		// TODO: this won't always work!
		ctx.WriteString(fmt.Sprintf("->getFieldValue(\"%s\")", access.AccessedIdent))

		if op.Next != nil {
			return writeObjectOperation(ctx, op.Next)
		}

		return nil
	}

	if op.Invocation != nil {
		err := writeObjectInvocation(ctx, op.Accessee.Type, op.Invocation)
		if err != nil {
			return err
		}

		if op.Next != nil {
			return writeObjectOperation(ctx, op.Next)
		}

		return nil
	}

	if op.Assignment != nil {
		ctx.WriteString(fmt.Sprintf("->setFieldValue(\"%s\", ", *op.Accessee.Ident))
		writeExpression(ctx, op.Assignment.Value)
		ctx.WriteString(")")

		if op.Next != nil {
			return fmt.Errorf("unexpected next operation after assignment on op %#v", op)
		}

		return nil
	}

	return fmt.Errorf("unknown operation in chained object operation: %#v", op)
}

func writeObjectInstantiation(ctx *Context, objInst *ObjectInstantiation) error {
	formattedFields := strings.Builder{}
	for i, f := range objInst.Fields {
		name := f.Name

		typeId, err := getTypeIdFor(ctx, f.Type)
		if err != nil {
			return err
		}

		descriptor := fmt.Sprintf("TsObjectFieldDescriptor(TsString(\"%s\"), %d)", name, typeId)

		value, err := ctx.WithinPrintContext(func(printCtx *Context) error {
			return writeExpression(printCtx, f.Value)
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
			TypeIdTsObject,
			formattedFields.String(),
		),
	)

	return nil
}

func writeObjectInvocation(ctx *Context, accesseeType *TypeSpec, invocation *ObjectInvocation) error {
	// TODO: this isn't always true!
	fn := accesseeType.Function

	// Write the args.
	args := strings.Builder{}
	numParams := len(fn.Parameters)
	for i, param := range fn.Parameters {
		arg := invocation.Arguments[i]

		expr, err := ctx.WithinPrintContext(func(printCtx *Context) error {
			return writeExpression(printCtx, arg)
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

func writeChainedObjectOperation(ctx *Context, op *ChainedObjectOperation) error {
	// TODO: this isn't always true!
	writeIdent(ctx, *op.First.Accessee.Ident)

	if err := rectifyChainedOperation(ctx, op); err != nil {
		return err
	}

	return writeObjectOperation(ctx, op.First)
}

func rectifyChainedOperation(ctx *Context, op *ChainedObjectOperation) error {
	// TODO: this won't always be true!
	// Figure out the type of the base object.
	t, err := ctx.CurrentScope.GetResolvedType(*op.First.Accessee.Ident)
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
			t, err = getFieldType(ctx, currentOp.Accessee.Type, currentOp.Access.AccessedIdent)
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
					t, err = getFieldType(ctx, currentOp.Prev.Accessee.Type, *currentOp.Invocation.Accessee.Ident)
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

func getFieldType(ctx *Context, t *TypeSpec, field string) (*TypeSpec, error) {
	if t.Object != nil {
		for _, f := range t.Object.Fields {
			if f.Name == field {
				return ctx.CurrentScope.ResolveType(f.Type)
			}
		}
	}

	if t := t.Interface; t != nil {
		for _, m := range t.Members {
			if m.Field != nil && m.Field.Name == field {
				return m.Field.Type, nil
			}

			if m.Method != nil && m.Method.Name == field {
				return &TypeSpec{
					Function: &Function{
						Parameters: m.Method.Parameters,
						// TODO: attach methods.
					},
				}, nil
			}
		}
	}

	return nil, errors.WithStack(fmt.Errorf("failed to resolve field type of %s on %#v", field, t))
}

func invocationFromAst(ctx *Context, accessee *Accessable, invoc *ast.ObjectInvocation) (*ObjectInvocation, error) {
	args := make([]*Expression, len(invoc.Arguments))
	for i, a := range invoc.Arguments {
		expr, err := expressionFromAst(ctx, &a)
		if err != nil {
			return nil, err
		}

		args[i] = expr
	}

	return &ObjectInvocation{
		Accessee:  accessee,
		Arguments: args,
	}, nil
}
