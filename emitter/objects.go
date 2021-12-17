package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type chainedObjectOperationLink struct {
	accessee     *ast.Accessable
	accesseeType *ast.Type
	operation    ast.ObjectOperation
	next         *chainedObjectOperationLink
	prev         *chainedObjectOperationLink
}

func buildOperationChain(ctx *Context, chainedOp *ast.ChainedObjectOperation) (lastLink *chainedObjectOperationLink, err error) {
	accessee := &chainedOp.Accessee
	var currentLink *chainedObjectOperationLink

	for _, op := range chainedOp.Operations {
		accesseeType, err := inferAccessableType(ctx, *accessee)
		if err != nil {
			return nil, errors.Wrap(err, "failed to infer accessable type during operation chain")
		}

		// Create a new link.
		link := chainedObjectOperationLink{
			accessee:     accessee,
			accesseeType: accesseeType,
			operation:    op,
			prev:         currentLink,
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
							return nil, fmt.Errorf("failed to find field %s in %#v", fieldName, accesseeType)
						}

						accessee, err = typeToAccessee(&field.Type)
						if err != nil {
							return nil, errors.Wrap(err, "failed to convert type to accessee for obj access")
						}

						continue
					}
				}
			}

			return nil, fmt.Errorf("unknown type in object access: %#v", accesseeType)
		}

		// Add "invocation" chain operation.
		if objInvoc := op.Invocation; objInvoc != nil {
			if t := accesseeType.NonUnionType; t != nil {
				if t := t.LiteralType; t != nil {
					accessee, err = typeToAccessee(accesseeType)
					if err != nil {
						return nil, errors.Wrap(err, "failed to convert type to accessee for obj invocation")
					}

					continue
				}
			}
		}

		return nil, fmt.Errorf("unknown operation in chained object operation: %#v", op)
	}

	return currentLink, nil
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
						return errors.Wrap(err, "failed to find field type for object literal type")
					}

					ctx.WriteString(fmt.Sprintf("(%s)ts_object_get_field(", mangleTypeNamePtr(fieldType)))

					if link.prev != nil {
						err = writeLink(ctx, link.prev)
						if err != nil {
							return errors.Wrap(err, "failed to write chained object link")
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
						return errors.Wrap(err, "failed to write ident")
					}

					err = writeObjectInvocation(ctx, link.operation.Invocation)

					return errors.Wrap(err, "failed to write object invocation")
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
			return errors.Wrap(err, "failed to infer type for field during obj instantiation")
		}

		typeId, err := getTypeIdFor(ctx, fieldType)
		if err != nil {
			return errors.Wrap(err, "failed to find type id for field type during obj instantiation")
		}

		// TODO: handle actual cases of types that need metadata.
		metadata := "NULL"

		fieldDescriptor := fmt.Sprintf("ts_object_field_descriptor_new(ts_string_new(\"%s\"), %d, %s)", name, typeId, metadata)

		value, err := ctx.WithinPrintContext(func(printCtx *Context) error {
			return writeExpression(printCtx, &field.Value)
		})

		if err != nil {
			return errors.Wrap(err, "failed to write value for field during obj instantiation")
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
	lastLink, err := buildOperationChain(ctx, op)
	if err != nil {
		return errors.Wrap(err, "failed to build operation chain")
	}

	if op.Assignment != nil {
		ctx.WriteString("ts_object_set_field(")
	}

	var startLink *chainedObjectOperationLink
	if op.Assignment == nil {
		startLink = lastLink
	} else {
		startLink = lastLink.prev
	}

	err = writeLink(ctx, startLink)
	if err != nil {
		return errors.Wrap(err, "failed to write operation chain")
	}

	if op.Assignment != nil {
		ctx.WriteString(fmt.Sprintf(", ts_string_new(\"%s\"), ", lastLink.operation.Access.AccessedIdent))
		writeExpression(ctx, &op.Assignment.Value)

		ctx.WriteString(")")
	}

	return nil
}
