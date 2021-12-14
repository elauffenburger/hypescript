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

	if expr.Invocation != nil {
		err := writeInvocable(ctx, expr.Invocation.Invoked)
		if err != nil {
			return err
		}

		ctx.WriteString("(")

		numArgs := len(expr.Invocation.Arguments)
		for i, arg := range expr.Invocation.Arguments {
			err := writeExpression(ctx, &arg)
			if err != nil {
				return err
			}

			if i != numArgs-1 {
				ctx.WriteString(", ")
			}
		}

		ctx.WriteString(")")

		return nil
	}

	if objInst := expr.ObjectInstantiation; objInst != nil {
		return writeObjectInstantiation(ctx, objInst)
	}

	if objectAccess := expr.ObjectAccess; objectAccess != nil {
		err := writeAccessable(ctx, objectAccess.Accessee)
		if err != nil {
			return err
		}

		ctx.WriteString("->")

		return writeExpression(ctx, &objectAccess.AccessedValue)
	}

	if expr.Ident != nil {
		return writeIdent(ctx, *expr.Ident)
	}

	return fmt.Errorf("unknown expression type: %#v", expr)
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

func writeInvocable(ctx *Context, invocable ast.Invocable) error {
	if invocable.Ident != nil {
		return writeIdent(ctx, *invocable.Ident)
	}

	return fmt.Errorf("unknown invocable type: %#v", invocable)
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
