package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

func writeStatementOrExpression(ctx *Context, stmtOrExpr *ast.StatementOrExpression) error {
	if stmtOrExpr.Statement != nil {
		err := writeStatement(ctx, stmtOrExpr.Statement)

		return errors.Wrap(err, "failed to write statement")
	}

	if stmtOrExpr.Expression != nil {
		err := writeExpression(ctx, stmtOrExpr.Expression)

		return errors.Wrap(err, "failed to write expression")
	}

	return fmt.Errorf("unknown StatementOrExpression: %#v", stmtOrExpr)
}

func writeStatement(ctx *Context, stmt *ast.Statement) error {
	if exprStmt := stmt.ExpressionStmt; exprStmt != nil {
		err := writeExpression(ctx, exprStmt)
		if err != nil {
			return errors.Wrap(err, "failed to write expression statement")
		}

		ctx.WriteString(";")

		return nil
	}

	if letDecl := stmt.LetDecl; letDecl != nil {
		letDeclType, err := inferType(ctx, &letDecl.Value)
		if err != nil {
			return errors.Wrap(err, "failed to write let decl")
		}

		var typeName string
		if nonUnionType := letDeclType.NonUnionType; nonUnionType != nil {
			if nonUnionType.TypeReference != nil {
				typeName = mangleTypeName(*letDeclType.NonUnionType.TypeReference)
			} else if literalType := nonUnionType.LiteralType; literalType != nil {

				if literalType.ObjectType != nil {
					typeName = "ts_object*"
				}
			}
		}

		ctx.WriteString(fmt.Sprintf("%s %s = ", typeName, letDecl.Name))

		err = writeExpression(ctx, &letDecl.Value)
		if err != nil {
			return errors.Wrap(err, "failed to write let decl")
		}

		ctx.WriteString(";")

		ctx.CurrentScope.AddIdentifer(letDecl.Name, *letDeclType)

		return nil
	}

	if returnStmt := stmt.ReturnStmt; returnStmt != nil {
		ctx.WriteString("return ")

		err := writeExpression(ctx, returnStmt)
		if err != nil {
			return errors.Wrap(err, "failed to write return statement")
		}

		ctx.WriteString(";")

		return nil
	}

	return fmt.Errorf("unknown statement type: %#v", stmt)
}

func writeAssignment(ctx *Context, stmt *ast.Assignment) error {
	ctx.WriteString(" = ")

	err := writeExpression(ctx, &stmt.Value)
	return errors.Wrap(err, "failed to write assignment")
}

func writeExpression(ctx *Context, expr *ast.Expression) error {
	if expr.Number != nil {
		if expr.Number.Integer != nil {
			ctx.WriteString(fmt.Sprintf("ts_num_new(%d)", *expr.Number.Integer))
			return nil
		}

		return fmt.Errorf("unknown number expression: %#v", *expr)
	}

	if expr.String != nil {
		ctx.WriteString(fmt.Sprintf("ts_string_new(%s)", *expr.String))
		return nil
	}

	if objInst := expr.ObjectInstantiation; objInst != nil {
		err := writeObjectInstantiation(ctx, objInst)

		return errors.Wrap(err, "failed to write object instantiation")
	}

	if chainedObjOperation := expr.ChainedObjectOperation; chainedObjOperation != nil {
		err := writeChainedObjectOperation(ctx, chainedObjOperation)

		return errors.Wrap(err, "failed to write chained object operation")
	}

	if expr.Ident != nil {
		err := writeIdent(ctx, *expr.Ident)

		return errors.Wrap(err, "failed to write ident")
	}

	if expr.IdentAssignment != nil {
		err := writeIdentAssignment(ctx, expr.IdentAssignment)

		return errors.Wrap(err, "failed to write ident assignment")
	}

	return fmt.Errorf("unknown expression type: %#v", expr)
}

func writeIdentAssignment(ctx *Context, asign *ast.IdentAssignment) error {
	err := writeIdent(ctx, asign.Ident)
	if err != nil {
		return errors.Wrap(err, "failed to write ident in ident assignment")
	}

	err = writeAssignment(ctx, &asign.Assignment)
	return errors.Wrap(err, "failed to write assignment in ident assignment")
}

type chainedObjectOperationLink struct {
	accessee     *ast.Accessable
	accesseeType *ast.Type
	operation    ast.ObjectOperation
	next         *chainedObjectOperationLink
	prev         *chainedObjectOperationLink
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

					ctx.WriteString(fmt.Sprintf("(%s*)ts_object_get_field(", mangleTypeName(fieldType)))

					if link.prev != nil {
						err = writeLink(ctx, link.prev)
						if err != nil {
							return errors.Wrap(err, "failed to write chained object link")
						}
					} else {
						if link.accessee.Ident == nil {
							return fmt.Errorf("expected ident for accessee: %#v", link)
						}

						ctx.WriteString(*link.accessee.Ident)
					}

					ctx.WriteString(fmt.Sprintf(", \"%s\")", field.Name))

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
					ctx.WriteString(*link.accessee.Ident)

					err := writeObjectInvocation(ctx, link.operation.Invocation)

					return errors.Wrap(err, "failed to write object invocation")
				}
			}
		}
	}

	return fmt.Errorf("unknown operation in chained object operation: %#v", link)
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

func writeChainedObjectOperation(ctx *Context, op *ast.ChainedObjectOperation) error {
	lastLink, err := buildOperationChain(ctx, op)
	if err != nil {
		return errors.Wrap(err, "failed to build operation chain")
	}

	err = writeLink(ctx, lastLink)
	if err != nil {
		return errors.Wrap(err, "failed to write operation chain")
	}

	if asign := op.Assignment; asign != nil {
		err = writeAssignment(ctx, asign)

		return errors.Wrap(err, "failed to write assignment")
	}

	return nil
}

func typeToAccessee(t *ast.Type) (*ast.Accessable, error) {
	if t := t.NonUnionType; t != nil {
		if t := t.LiteralType; t != nil {
			return &ast.Accessable{LiteralType: t}, nil
		}

		if t := t.TypeReference; t != nil {
			return &ast.Accessable{Ident: t}, nil
		}
	}

	return nil, fmt.Errorf("could not convert type to accessee: %#v", t)
}

func getRuntimeTypeName(t *ast.Type) (string, error) {
	if t.NonUnionType != nil {
		if t.NonUnionType.TypeReference != nil {
			return *t.NonUnionType.TypeReference, nil
		}

		if t := t.NonUnionType.LiteralType; t != nil {
			if t.ObjectType != nil {
				return string(TsObject), nil
			}

			if t.FunctionType != nil {
				return string(TsFunction), nil
			}
		}
	}

	return "", fmt.Errorf("unknown type: %#v", t)
}

func writeObjectInvocation(ctx *Context, invocation *ast.ObjectInvocation) error {
	// TODO: support arguments.
	ctx.WriteString("()")

	return nil
}

func inferAccessableType(ctx *Context, accessable ast.Accessable) (*ast.Type, error) {
	if accessable.Ident != nil {
		return ctx.TypeOf(*accessable.Ident)
	}

	if accessable.LiteralType != nil {
		return &ast.Type{NonUnionType: &ast.NonUnionType{LiteralType: accessable.LiteralType}}, nil
	}

	return nil, fmt.Errorf("unknown accessable type: %#v", accessable)
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

	ctx.WriteString(fmt.Sprintf("ts_object_new((ts_object_field*[]){%s})", formattedFields.String()))

	return nil
}

func writeIdent(ctx *Context, ident string) error {
	identType, err := ctx.TypeOf(ident)

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

func getTypeIdFor(ctx *Context, t *ast.Type) (int, error) {
	// DO NOT SUBMIT; need to actually make this work!

	return 0, nil
}
