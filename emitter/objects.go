package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
	"strings"
)

type chainedObjectOperationLink struct {
	accessee     *ast.Accessable
	accesseeType *ast.Type
	operation    ast.ObjectOperation
	next         *chainedObjectOperationLink
	prev         *chainedObjectOperationLink
}

func buildOperationChain(ctx *Context, chainedOp *ast.ChainedObjectOperation) (first, last *chainedObjectOperationLink, err error) {
	var firstLink *chainedObjectOperationLink
	var currentLink *chainedObjectOperationLink

	accessee := &chainedOp.Accessee
	for _, op := range chainedOp.Operations {
		accesseeType, err := inferAccessableType(ctx, *accessee)
		if err != nil {
			return nil, nil, err
		}

		// Create a new link.
		link := chainedObjectOperationLink{
			accessee:     accessee,
			accesseeType: accesseeType,
			operation:    op,
			prev:         currentLink,
		}

		if firstLink == nil {
			firstLink = &link
		}

		// Link the current link (if any) to the new link.
		if currentLink != nil {
			currentLink.next = &link
		}

		// Make the current link the new link.
		currentLink = &link

		// Add "access" chain operation.
		if access := op.Access; access != nil {
			accessee, err = buildObjectAccessOperation(ctx, access, accessee, accesseeType)
			if err != nil {
				return nil, nil, err
			}

			continue
		}

		// Add "invocation" chain operation.
		if invoc := op.Invocation; invoc != nil {
			accessee, err = buildObjectInvocationOperation(ctx, invoc, accessee, accesseeType)
			if err != nil {
				return nil, nil, err
			}

			continue
		}

		return nil, nil, fmt.Errorf("unknown operation in chained object operation: %#v", op)
	}

	return firstLink, currentLink, nil
}

func buildObjectAccessOperation(ctx *Context, access *ast.ObjectAccess, accessee *ast.Accessable, accesseeType *ast.Type) (nextAccessee *ast.Accessable, err error) {
	if t := accesseeType.NonUnionType; t != nil {
		if t := t.LiteralType; t != nil {
			if t := t.ObjectType; t != nil {
				fieldName := access.AccessedIdent
				var field *ast.ObjectTypeField

				for _, f := range t.Fields {
					if f.Name == fieldName {
						field = &f

						break
					}
				}

				if field == nil {
					return nil, fmt.Errorf("failed to find field %s in %#v", fieldName, accesseeType)
				}

				return typeToAccessee(&field.Type)
			}
		}

		if t := t.TypeReference; t != nil {
			referencedType, err := ctx.TypeOf(*t)
			if err != nil {
				return nil, err
			}

			return buildObjectAccessOperation(ctx, access, accessee, referencedType)
		}
	}

	return nil, fmt.Errorf("unknown type in object access: %#v", accesseeType)
}

func buildObjectInvocationOperation(ctx *Context, invoc *ast.ObjectInvocation, accessee *ast.Accessable, accesseeType *ast.Type) (nextAccessee *ast.Accessable, err error) {
	if t := accesseeType.NonUnionType; t != nil {
		if t := t.LiteralType; t != nil {
			if t := t.FunctionType; t != nil {
				return typeToAccessee(accesseeType)
			}
		}

		if t := t.TypeReference; t != nil {
			referencedType, err := ctx.TypeOf(*t)
			if err != nil {
				return nil, err
			}

			return buildObjectInvocationOperation(ctx, invoc, accessee, referencedType)
		}
	}

	return nil, fmt.Errorf("unknown type in object invocation: %#v", accesseeType)
}

func writeLink(ctx *Context, link, endLink *chainedObjectOperationLink) error {
	if link == endLink {
		return nil
	}

	if access := link.operation.Access; access != nil {
		// TODO: this won't always work!
		ctx.WriteString(fmt.Sprintf("->getFieldValue(\"%s\")", access.AccessedIdent))

		if link.next != nil {
			err := writeLink(ctx, link.next, endLink)
			if err != nil {
				return err
			}
		}

		return nil
	}

	if link.operation.Invocation != nil {
		return writeObjectInvocation(ctx, link.accesseeType, link.operation.Invocation)
	}

	return fmt.Errorf("unknown operation in chained object operation: %#v", link)
}

func writeObjectInstantiation(ctx *Context, objInst *ast.ObjectInstantiation) error {
	formattedFields := strings.Builder{}
	for i, field := range objInst.Fields {
		name := field.Name

		fieldType, err := inferType(ctx, &field.Value)
		if err != nil {
			return err
		}

		typeId, err := getTypeIdFor(ctx, fieldType)
		if err != nil {
			return err
		}

		fieldDescriptor := fmt.Sprintf("TsObjectFieldDescriptor(TsString(\"%s\"), %d)", name, typeId)

		value, err := ctx.WithinPrintContext(func(printCtx *Context) error {
			return writeExpression(printCtx, &field.Value)
		})

		if err != nil {
			return err
		}

		formattedFields.WriteString(
			fmt.Sprintf(
				"new TsObjectField(%s, %s)", fieldDescriptor,
				value,
			),
		)

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

func writeObjectInvocation(ctx *Context, accesseeType *ast.Type, invocation *ast.ObjectInvocation) error {
	// TODO: this isn't always true!
	fn := accesseeType.NonUnionType.LiteralType.FunctionType

	// Write the args.
	args := strings.Builder{}
	numParams := len(fn.Parameters)
	for i, param := range fn.Parameters {
		arg := invocation.Arguments[i]

		expr, err := ctx.WithinPrintContext(func(printCtx *Context) error {
			return writeExpression(printCtx, &arg)
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

func writeChainedObjectOperation(ctx *Context, op *ast.ChainedObjectOperation) error {
	firstLink, lastLink, err := buildOperationChain(ctx, op)
	if err != nil {
		return err
	}

	// TODO: this isn't always true!
	writeIdent(ctx, *firstLink.accessee.Ident)

	var endLink *chainedObjectOperationLink = nil
	if op.Assignment != nil {
		endLink = lastLink
	}

	err = writeLink(ctx, firstLink, endLink)
	if err != nil {
		return err
	}

	if op.Assignment != nil {
		ctx.WriteString(fmt.Sprintf("->setFieldValue(\"%s\", ", lastLink.operation.Access.AccessedIdent))
		writeExpression(ctx, &op.Assignment.Value)
		ctx.WriteString(")")
	}

	return nil
}