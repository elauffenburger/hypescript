package regpass

import (
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/core"
)

func (ctx *Context) registerFunctionDeclaration(astFn *ast.FunctionInstantiation) (*core.StatementOrExpression, error) {
	explRtnType, err := ctx.typeSpecFromAst(astFn.ReturnType)
	if err != nil {
		return nil, err
	}

	// Figure out our params.
	params, err := ctx.functionParameterFromAst(astFn.Parameters)
	if err != nil {
		return nil, err
	}

	// Build our type spec.
	fn := &core.Function{
		Name:               astFn.Name,
		Parameters:         params,
		ImplicitReturnType: nil,
		ExplicitReturnType: explRtnType,
	}

	// Add the type spec.
	if astFn.Name != nil {
		ctx.currentScope().AddIdentifer(*astFn.Name, &core.TypeSpec{Function: fn})
	}

	// Add the statement.
	//
	// We'll return this as the registered stmt/expr.
	result := ctx.currentScope().AddStmtOrExpr(&core.StatementOrExpression{
		Statement: &core.Statement{
			FunctionInstantiation: fn,
		},
	})

	ctx.EnterScope()

	// Add the params as idents in the current scope.
	for _, p := range fn.Parameters {
		ctx.currentScope().AddIdentifer(p.Name, p.Type)
	}

	for _, stmtOrExpr := range astFn.Body {
		stmtOrExpr, err := ctx.registerStatementOrExpression(stmtOrExpr)
		if err != nil {
			return nil, err
		}

		// Add the statement or expression to the fn body.
		fn.Body = append(fn.Body, stmtOrExpr)

		// Expressions can't produce identifiers, so we can skip them.
		if stmtOrExpr.Statement == nil {
			continue
		}

		stmt := stmtOrExpr.Statement

		// If this is a let decl, add the ident to the current scope.
		if let := stmt.LetDecl; let != nil {
			infType, err := ctx.currentScope().InferType(let.Value)
			if err != nil {
				return nil, err
			}

			ctx.currentScope().AddIdentifer(let.Name, infType)

			continue
		}

		// If this is a return stmt, update the implicit return type.
		if stmt.ReturnStmt != nil {
			rtnStmtType, err := ctx.currentScope().InferType(stmt.ReturnStmt)
			if err != nil {
				return nil, err
			}

			// If we don't have an implied type yet, use this return statement's.
			if fn.ImplicitReturnType == nil {
				fn.ImplicitReturnType = rtnStmtType
				continue
			}

			// ...otherwise, if the return types match, bail out.
			if fn.ImplicitReturnType == rtnStmtType {
				continue
			}

			// ...otherwise, we need to treat this as a union of the existing type and this type.
			union, err := createUnionType(fn.ImplicitReturnType, rtnStmtType)
			if err != nil {
				return nil, err
			}

			fn.ImplicitReturnType = union
		}
	}

	if fn.ImplicitReturnType == nil {
		fn.ImplicitReturnType = &core.TypeSpec{TypeReference: strRef("void")}
	}

	ctx.ExitScope()

	return result, nil
}

func (ctx *Context) functionParameterFromAst(params []*ast.FunctionParameter) ([]*core.FunctionParameter, error) {
	results := make([]*core.FunctionParameter, 0)
	for _, p := range params {
		pType, err := ctx.typeSpecFromAst(&p.Type)
		if err != nil {
			return nil, err
		}

		results = append(results, &core.FunctionParameter{
			Name: p.Name,
			Type: pType,
		})
	}

	return results, nil
}
