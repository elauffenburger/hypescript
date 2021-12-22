package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
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
	if fnInst := stmt.FunctionInstantiation; fnInst != nil {
		return writeFunctionDeclaration(ctx, fnInst)
	}

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

		ctx.WriteString(fmt.Sprintf("auto %s = ", letDecl.Name))

		err = writeExpression(ctx, &letDecl.Value)
		if err != nil {
			return err
		}

		ctx.WriteString(";")

		ctx.CurrentScope.AddIdentifer(letDecl.Name, letDeclType)

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
			ctx.WriteString(fmt.Sprintf("new TsNum(%d)", *expr.Number.Integer))
			return nil
		}

		return fmt.Errorf("unknown number expression: %#v", *expr)
	}

	if expr.String != nil {
		ctx.WriteString(fmt.Sprintf("new TsString(%s)", *expr.String))
		return nil
	}

	if objInst := expr.ObjectInstantiation; objInst != nil {
		return writeObjectInstantiation(ctx, objInst)
	}

	if fnInst := expr.FunctionInstantiation; fnInst != nil {
		fnInfo, err := buildFunctionInfo(ctx, fnInst)
		if err != nil {
			return err
		}

		return writeFunction(ctx, fnInst, fnInfo)
	}

	if chainedObjOperation := expr.ChainedObjectOperation; chainedObjOperation != nil {
		return writeChainedObjectOperation(ctx, chainedObjOperation)
	}

	if expr.Ident != nil {
		return writeIdent(ctx, *expr.Ident)
	}

	if expr.IdentAssignment != nil {
		return writeIdentAssignment(ctx, expr.IdentAssignment)
	}

	return fmt.Errorf("unknown expression type: %#v", expr)
}

func writeIdentAssignment(ctx *Context, asign *ast.IdentAssignment) error {
	err := writeIdent(ctx, asign.Ident)
	if err != nil {
		return err
	}

	ctx.WriteString(" = ")

	return writeExpression(ctx, &asign.Assignment.Value)
}

func typeToAccessee(t *ast.TypeIdentifier) (*ast.Accessable, error) {
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

func getRuntimeTypeName(t *ast.TypeIdentifier) (string, error) {
	if t.NonUnionType != nil {
		if t.NonUnionType.TypeReference != nil {
			return *t.NonUnionType.TypeReference, nil
		}

		if t := t.NonUnionType.LiteralType; t != nil {
			if t.ObjectType != nil {
				return string(RtTsObject), nil
			}

			if t.FunctionType != nil {
				return string(RtTsFunction), nil
			}
		}
	}

	return "", fmt.Errorf("unknown type: %#v", t)
}

func inferAccessableType(ctx *Context, accessable ast.Accessable) (*TypeSpec, error) {
	if accessable.Ident != nil {
		return ctx.TypeOf(*accessable.Ident)
	}

	if lit := accessable.LiteralType; lit != nil {
		if lit.FunctionType != nil {
			return &TypeSpec{FunctionType: lit.FunctionType}, nil
		}

		if lit.ObjectType != nil {
			return &TypeSpec{ObjectType: lit.ObjectType}, nil
		}
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

func getTypeIdFor(ctx *Context, t *TypeSpec) (int, error) {
	// TODO; need to actually make this work!

	return 0, nil
}
