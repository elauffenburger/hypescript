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
							return nil, nil, fmt.Errorf("failed to find field %s in %#v", fieldName, accesseeType)
						}

						accessee, err = typeToAccessee(&field.Type)
						if err != nil {
							return nil, nil, err
						}

						continue
					}
				}
			}

			return nil, nil, fmt.Errorf("unknown type in object access: %#v", accesseeType)
		}

		// Add "invocation" chain operation.
		if objInvoc := op.Invocation; objInvoc != nil {
			if t := accesseeType.NonUnionType; t != nil {
				if t := t.LiteralType; t != nil {
					accessee, err = typeToAccessee(accesseeType)
					if err != nil {
						return nil, nil, err
					}

					continue
				}
			}
		}

		return nil, nil, fmt.Errorf("unknown operation in chained object operation: %#v", op)
	}

	return firstLink, currentLink, nil
}

func writeLink(ctx *Context, link *chainedObjectOperationLink) error {
	if link.operation.Access != nil {
		if t := link.accesseeType.NonUnionType; t != nil {
			if t := t.LiteralType; t != nil {
				if t := t.ObjectType; t != nil {
					// TODO: this is not always correct!
					fieldName := link.operation.Access.AccessedIdent
					var field *ast.ObjectTypeField

					for _, f := range t.Fields {
						if f.Name == fieldName {
							field = &f

							break
						}
					}

					if field == nil {
						return fmt.Errorf("failed to find field %s in %#v", fieldName, link.accesseeType)
					}

					ctx.WriteString(fmt.Sprintf("->getFieldValue(\"%s\")", field.Name))

					if link.next != nil {
						err := writeLink(ctx, link.next)
						if err != nil {
							return err
						}
					}

					return nil
				}
			}
		}

		return fmt.Errorf("unknown type in object access: %#v", link.accesseeType)
	}

	if link.operation.Invocation != nil {
		if t := link.accesseeType.NonUnionType; t != nil {
			if t := t.LiteralType; t != nil {
				if t.FunctionType != nil {
					return writeObjectInvocation(ctx, link.accesseeType, link.operation.Invocation)
				}
			}
		}
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
				"std::make_shared<TsObjectField>(TsObjectField(%s, %s))", fieldDescriptor,
				value,
			),
		)

		if i != len(objInst.Fields)-1 {
			formattedFields.WriteString(", ")
		}
	}

	ctx.WriteString(
		fmt.Sprintf(
			"std::make_shared<TsObject>(TsObject(%d, TsCoreHelpers::toVector<std::shared_ptr<TsObjectField>>(%s)))",
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

		args.WriteString(fmt.Sprintf("TsFunctionArg(\"%s\"", param.Name))

		err := writeExpression(ctx, &arg)
		if err != nil {
			return err
		}

		args.WriteString(")")

		if i != numParams-1 {
			args.WriteString(", ")
		}
	}

	ctx.WriteString(fmt.Sprintf("->invoke(TsCoreHelpers::toVector<TsFunctionArg>(%s))", args.String()))

	return nil
}

func writeChainedObjectOperation(ctx *Context, op *ast.ChainedObjectOperation) error {
	firstLink, lastLink, err := buildOperationChain(ctx, op)
	if err != nil {
		return err
	}

	// TODO: this isn't always true!
	writeIdent(ctx, *firstLink.accessee.Ident)

	err = writeLink(ctx, firstLink)
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
