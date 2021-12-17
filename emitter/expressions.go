package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"

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
				typeName = mangleTypeNamePtr(*letDeclType.NonUnionType.TypeReference)
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

	ctx.WriteString(" = ")

	err = writeExpression(ctx, &asign.Assignment.Value)
	return errors.Wrap(err, "failed to write ident assignment")
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

func inferAccessableType(ctx *Context, accessable ast.Accessable) (*ast.Type, error) {
	if accessable.Ident != nil {
		return ctx.TypeOf(*accessable.Ident)
	}

	if accessable.LiteralType != nil {
		return &ast.Type{NonUnionType: &ast.NonUnionType{LiteralType: accessable.LiteralType}}, nil
	}

	return nil, fmt.Errorf("unknown accessable type: %#v", accessable)
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
