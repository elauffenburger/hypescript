package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
	"strings"
)

func writeStatementOrExpression(ctx *Context, stmtOrExpr *ast.StatementOrExpression) error {
	if stmtOrExpr.Statement != nil {
		return writeStatement(ctx, stmtOrExpr.Statement)
	}

	if stmtOrExpr.Expression != nil {
		return writeExpression(ctx, stmtOrExpr.Expression)
	}

	return fmt.Errorf("unknown StatementOrExpression: %#v", stmtOrExpr)
}

func writeStatement(ctx *Context, stmt *ast.Statement) error {
	if exprStmt := stmt.ExpressionStmt; exprStmt != nil {
		err := writeExpression(ctx, exprStmt)
		if err != nil {
			return err
		}

		ctx.WriteString(";")

		return nil
	}

	if letDecl := stmt.LetDecl; letDecl != nil {
		letDeclType, err := inferType(ctx, &letDecl.Value)
		if err != nil {
			return err
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
			return err
		}

		ctx.WriteString(";")

		ctx.CurrentScope.AddIdentifer(letDecl.Name, *letDeclType)

		return nil
	}

	if assignmentStmt := stmt.AssignmentStmt; assignmentStmt != nil {
		ctx.WriteString(fmt.Sprintf("%s = ", assignmentStmt.Ident))

		err := writeExpression(ctx, &assignmentStmt.Value)
		if err != nil {
			return err
		}

		ctx.WriteString(";")

		return nil
	}

	if returnStmt := stmt.ReturnStmt; returnStmt != nil {
		ctx.WriteString("return ")

		err := writeExpression(ctx, returnStmt)
		if err != nil {
			return err
		}

		ctx.WriteString(";")

		return nil
	}

	return fmt.Errorf("unknown statement type: %#v", stmt)
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
		return writeObjectInstantiation(ctx, objInst)
	}

	if chainedObjOperation := expr.ChainedObjectOperation; chainedObjOperation != nil {
		accessee := chainedObjOperation.Accessee

		for _, operation := range chainedObjOperation.Operations {
			if objAccess := operation.Access; objAccess != nil {
				err := writeAccessedObjectAccess(ctx, &accessee, operation.Access)
				if err != nil {
					return err
				}

				continue
			}

			if objInvoc := operation.Invocation; objInvoc != nil {
				err := writeAccessedObjectInvocation(ctx, &accessee, operation.Invocation)
				if err != nil {
					return err
				}

				continue
			}

			return fmt.Errorf("unknown operation in chained object operation: %#v", operation)
		}

		return nil
	}

	if expr.Ident != nil {
		return writeIdent(ctx, *expr.Ident)
	}

	return fmt.Errorf("unknown expression type: %#v", expr)
}

func writeAccessedObjectAccess(ctx *Context, accessee *ast.Accessable, access *ast.ObjectAccess) error {
	// TODO: this will _not_ work for multiple chained object accessess
	// because of how ts_object_get_field calls will be emitted!

	accesseeType, err := inferAccessableType(ctx, *accessee)
	if err != nil {
		return err
	}

	// Figure out how to access the accessee.
	if nonUnionType := accesseeType.NonUnionType; nonUnionType != nil {
		if literalType := nonUnionType.LiteralType; literalType != nil {
			// Check if this is an object type.
			if objType := literalType.ObjectType; objType != nil {
				fieldName := access.AccessedIdent
				var field *ast.ObjectTypeField

				for _, f := range objType.Fields {
					if f.Name == fieldName {
						field = &f

						break
					}
				}

				if field == nil {
					return fmt.Errorf("failed to find field %s in %#v", fieldName, objType)
				}

				// DO NOT SUBMIT: this is super not always true!
				fieldType := *field.Type.NonUnionType.TypeReference
				ctx.WriteString(fmt.Sprintf("(%s)ts_object_get_field(", mangleTypeName(fieldType)))

				err = writeAccessable(ctx, *accessee)
				if err != nil {
					return err
				}

				ctx.WriteString(fmt.Sprintf(", \"%s\")", field.Name))
				return nil
			}
		}
	}

	return fmt.Errorf("unknown type in object access: %#v", accesseeType)
}

func writeAccessedObjectInvocation(ctx *Context, accessee *ast.Accessable, invocation *ast.ObjectInvocation) error {
	accesseeType, err := inferAccessableType(ctx, *accessee)
	if err != nil {
		return err
	}

	// Figure out how to access the accessee.
	if nonUnionType := accesseeType.NonUnionType; nonUnionType != nil {
		if literalType := nonUnionType.LiteralType; literalType != nil {
			if objType := literalType.ObjectType; objType != nil {
				ctx.WriteString(*accessee.Ident)

				return writeObjectInvocation(ctx, invocation)
			}

			if functionType := literalType.FunctionType; functionType != nil {
				ctx.WriteString(*accessee.Ident)

				return writeObjectInvocation(ctx, invocation)
			}
		}
	}

	return fmt.Errorf("unknown type in object invocation: %#v", accesseeType)
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

	return nil, fmt.Errorf("unknown accessable type: %#v", accessable)
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

func writeAccessable(ctx *Context, accessable ast.Accessable) error {
	if accessable.Ident != nil {
		ctx.WriteString(*accessable.Ident)

		return nil
	}

	return fmt.Errorf("unknown accessable type: %#v", accessable)
}

func getTypeIdFor(ctx *Context, t *ast.Type) (int, error) {
	// DO NOT SUBMIT; need to actually make this work!

	return 0, nil
}
