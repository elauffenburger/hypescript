package writepass

import (
	"elauffenburger/hypescript/emitter/core"
	"fmt"

	"github.com/pkg/errors"
)

func (ctx *Context) writeStatementOrExpression(stmtOrExpr *core.StatementOrExpression) error {
	return ctx.WithinScope(stmtOrExpr.Scope, func() error {
		if stmtOrExpr.Statement != nil {
			return ctx.writeStatement(stmtOrExpr.Statement)
		}

		if stmtOrExpr.Expression != nil {
			return ctx.writeExpression(stmtOrExpr.Expression)
		}

		return fmt.Errorf("unknown StatementOrExpression: %#v", stmtOrExpr)
	})
}

func (ctx *Context) writeStatement(stmt *core.Statement) error {
	if fnInst := stmt.FunctionInstantiation; fnInst != nil {
		return ctx.writeFunctionDeclaration(fnInst)
	}

	if exprStmt := stmt.ExpressionStmt; exprStmt != nil {
		err := ctx.writeExpression(exprStmt)
		if err != nil {
			return err
		}

		ctx.WriteString(";")

		return nil
	}

	if letDecl := stmt.LetDecl; letDecl != nil {
		ctx.WriteString(fmt.Sprintf("auto %s = ", letDecl.Name))

		err := ctx.writeExpression(letDecl.Value)
		if err != nil {
			return err
		}

		ctx.WriteString(";")

		return nil
	}

	if returnStmt := stmt.ReturnStmt; returnStmt != nil {
		ctx.WriteString("return ")

		err := ctx.writeExpression(returnStmt)
		if err != nil {
			return err
		}

		ctx.WriteString(";")

		return nil
	}

	return errors.WithStack(fmt.Errorf("unknown statement type: %#v", stmt))
}

func (ctx *Context) writeExpression(expr *core.Expression) error {
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
		return ctx.writeObjectInstantiation(objInst)
	}

	if fnInst := expr.FunctionInstantiation; fnInst != nil {
		return ctx.writeFunction(fnInst)
	}

	if chainedObjOperation := expr.ChainedObjectOperation; chainedObjOperation != nil {
		return ctx.writeChainedObjectOperation(chainedObjOperation)
	}

	if expr.Ident != nil {
		return ctx.writeIdent(*expr.Ident)
	}

	if expr.IdentAssignment != nil {
		return ctx.writeIdentAssignment(expr.IdentAssignment)
	}

	return fmt.Errorf("unknown expression type: %#v", expr)
}
