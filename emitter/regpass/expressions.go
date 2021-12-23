package regpass

import (
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/core"
	"fmt"

	"github.com/pkg/errors"
)

func (ctx *Context) registerStatementOrExpression(stmtOrExpr *ast.StatementOrExpression) (*core.StatementOrExpression, error) {
	if stmtOrExpr.Statement != nil {
		return ctx.registerStatement(stmtOrExpr.Statement)
	}

	if stmtOrExpr.Expression != nil {
		return ctx.registerExpression(stmtOrExpr.Expression)
	}

	return nil, fmt.Errorf("unknown StatementOrExpression: %#v", stmtOrExpr)
}

func (ctx *Context) registerExpression(expr *ast.Expression) (*core.StatementOrExpression, error) {
	if expr.FunctionInstantiation != nil {
		return ctx.registerFunctionDeclaration(expr.FunctionInstantiation)
	}

	e, err := ctx.expressionFromAst(expr)
	if err != nil {
		return nil, err
	}

	return ctx.currentScope().AddStmtOrExpr(&core.StatementOrExpression{Expression: e}), nil
}

func (ctx *Context) registerStatement(stmt *ast.Statement) (*core.StatementOrExpression, error) {
	if fnInst := stmt.FunctionInstantiation; fnInst != nil {
		return ctx.registerFunctionDeclaration(fnInst)
	}

	if letDecl := stmt.LetDecl; letDecl != nil {
		value, err := ctx.expressionFromAst(&letDecl.Value)
		if err != nil {
			return nil, err
		}

		letDeclType, err := ctx.currentScope().InferType(value)
		if err != nil {
			return nil, err
		}

		ctx.currentScope().AddIdentifer(letDecl.Name, letDeclType)

		stmt, err := ctx.statementFromAst(stmt)
		if err != nil {
			return nil, err
		}

		return ctx.currentScope().AddStmt(stmt), nil
	}

	if stmt.ReturnStmt != nil {
		expr, err := ctx.expressionFromAst(stmt.ReturnStmt)
		if err != nil {
			return nil, err
		}

		return ctx.currentScope().AddStmt(&core.Statement{ReturnStmt: expr}), nil
	}

	if stmt.ExpressionStmt != nil {
		expr, err := ctx.expressionFromAst(stmt.ExpressionStmt)
		if err != nil {
			return nil, err
		}

		return ctx.currentScope().AddStmt(&core.Statement{ExpressionStmt: expr}), nil
	}

	return nil, errors.WithStack(fmt.Errorf("unknown statement type: %#v", stmt))
}

func (ctx *Context) statementFromAst(stmt *ast.Statement) (*core.Statement, error) {
	if stmt.ExpressionStmt != nil {
		expr, err := ctx.expressionFromAst(stmt.ExpressionStmt)
		if err != nil {
			return nil, err
		}

		return &core.Statement{ExpressionStmt: expr}, nil
	}

	if fn := stmt.FunctionInstantiation; fn != nil {
		fn, err := ctx.functionFromAst(fn)
		if err != nil {
			return nil, err
		}

		return &core.Statement{FunctionInstantiation: fn}, nil
	}

	if stmt.LetDecl != nil {
		value, err := ctx.expressionFromAst(&stmt.LetDecl.Value)
		if err != nil {
			return nil, err
		}

		return &core.Statement{
			LetDecl: &core.LetDecl{
				Name:  stmt.LetDecl.Name,
				Value: value,
			},
		}, nil
	}

	if stmt.ReturnStmt != nil {
		rtn, err := ctx.expressionFromAst(stmt.ReturnStmt)
		if err != nil {
			return nil, err
		}

		return &core.Statement{ReturnStmt: rtn}, nil
	}

	return nil, fmt.Errorf("unknown statement: %#v", stmt)
}

func (ctx *Context) expressionFromAst(expr *ast.Expression) (*core.Expression, error) {
	if expr.Number != nil {
		if expr.Number.Integer != nil {
			return &core.Expression{Number: &core.Number{Integer: expr.Number.Integer}}, nil
		}

		return nil, fmt.Errorf("unknown number expression: %#v", *expr)
	}

	if expr.String != nil {
		return &core.Expression{String: expr.String}, nil
	}

	if objInst := expr.ObjectInstantiation; objInst != nil {
		fields := make([]*core.ObjectFieldInstantiation, len(objInst.Fields))
		for i, f := range objInst.Fields {
			value, err := ctx.expressionFromAst(&f.Value)
			if err != nil {
				return nil, err
			}

			t, err := ctx.currentScope().InferType(value)
			if err != nil {
				return nil, err
			}

			fields[i] = &core.ObjectFieldInstantiation{
				Name:  f.Name,
				Type:  t,
				Value: value,
			}
		}

		return &core.Expression{ObjectInstantiation: &core.ObjectInstantiation{Fields: fields}}, nil
	}

	if fnInst := expr.FunctionInstantiation; fnInst != nil {
		fn, err := ctx.functionFromAst(fnInst)
		if err != nil {
			return nil, err
		}

		return &core.Expression{FunctionInstantiation: fn}, nil
	}

	if chainedObjOperation := expr.ChainedObjectOperation; chainedObjOperation != nil {
		op, err := ctx.chainedObjOperationFromAst(chainedObjOperation)
		if err != nil {
			return nil, err
		}

		return &core.Expression{ChainedObjectOperation: op}, nil
	}

	if expr.Ident != nil {
		return &core.Expression{Ident: expr.Ident}, nil
	}

	if expr.IdentAssignment != nil {
		value, err := ctx.expressionFromAst(&expr.IdentAssignment.Assignment.Value)
		if err != nil {
			return nil, err
		}

		return &core.Expression{
			IdentAssignment: &core.IdentAssignment{
				Ident: expr.IdentAssignment.Ident,
				Assignment: core.Assignment{
					Value: value,
				},
			},
		}, nil
	}

	return nil, fmt.Errorf("unknown expression type: %#v", expr)
}
