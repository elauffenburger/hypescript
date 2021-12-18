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

					fieldType, err := getRuntimeTypeName(&field.Type)
					if err != nil {
						return err
					}

					ctx.WriteString(fmt.Sprintf("%s.getField(%s)", mangleTypeNamePtr(fieldType)))

					if link.next != nil {
						err = writeLink(ctx, link.next)
						if err != nil {
							return err
						}
					} else {
						if link.accessee.Ident == nil {
							return fmt.Errorf("expected ident for accessee: %#v", link)
						}

						writeIdent(ctx, *link.accessee.Ident)
					}

					ctx.WriteString(fmt.Sprintf(", ts_string_new(\"%s\"))", field.Name))

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
					err := writeIdent(ctx, *link.accessee.Ident)
					if err != nil {
						return err
					}

					return writeObjectInvocation(ctx, link.operation.Invocation)
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

		// TODO: handle actual cases of types that need metadata.
		metadata := "NULL"

		fieldDescriptor := fmt.Sprintf("ts_object_field_descriptor_new(ts_string_new(\"%s\"), %d, %s)", name, typeId, metadata)

		value, err := ctx.WithinPrintContext(func(printCtx *Context) error {
			return writeExpression(printCtx, &field.Value)
		})

		if err != nil {
			return err
		}

		formattedFields.WriteString(fmt.Sprintf("ts_object_field_new(%s, %s)", fieldDescriptor, value))

		if i != len(objInst.Fields)-1 {
			formattedFields.WriteString(", ")
		}
	}

	ctx.WriteString(fmt.Sprintf("ts_object_new((ts_object_field*[]){%s}, %d)", formattedFields.String(), len(objInst.Fields)))

	return nil
}

func writeObjectInvocation(ctx *Context, invocation *ast.ObjectInvocation) error {
	// TODO: support arguments.
	ctx.WriteString("()")

	return nil
}

func writeChainedObjectOperation(ctx *Context, op *ast.ChainedObjectOperation) error {
	firstLink, lastLink, err := buildOperationChain(ctx, op)
	if err != nil {
		return err
	}

	err = writeLink(ctx, firstLink)
	if err != nil {
		return err
	}

	if op.Assignment != nil {
		ctx.WriteString(fmt.Sprintf(".setField(%s, ", lastLink.operation.Access.AccessedIdent))
		writeExpression(ctx, &op.Assignment.Value)
		ctx.WriteString(")")
	}

	return nil
}
