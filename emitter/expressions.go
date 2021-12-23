package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"

	"github.com/pkg/errors"
)

func writeStatementOrExpression(ctx *Context, stmtOrExpr *StatementOrExpression) error {
	return ctx.WithinScope(stmtOrExpr.Scope, func() error {
		if stmtOrExpr.Statement != nil {
			return writeStatement(ctx, stmtOrExpr.Statement)
		}

		if stmtOrExpr.Expression != nil {
			return writeExpression(ctx, stmtOrExpr.Expression)
		}

		return fmt.Errorf("unknown StatementOrExpression: %#v", stmtOrExpr)
	})
}

func writeStatement(ctx *Context, stmt *Statement) error {
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
		ctx.WriteString(fmt.Sprintf("auto %s = ", letDecl.Name))

		err := writeExpression(ctx, letDecl.Value)
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

	return errors.WithStack(fmt.Errorf("unknown statement type: %#v", stmt))
}

func writeExpression(ctx *Context, expr *Expression) error {
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
		return writeFunction(ctx, fnInst)
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

func writeIdentAssignment(ctx *Context, assign *IdentAssignment) error {
	err := writeIdent(ctx, assign.Ident)
	if err != nil {
		return err
	}

	ctx.WriteString(" = ")

	return writeExpression(ctx, assign.Assignment.Value)
}

func writeIdent(ctx *Context, ident string) error {
	identType, err := ctx.CurrentScope.TypeOf(ident)

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

func (ctx *Context) registerStatementOrExpression(stmtOrExpr *ast.StatementOrExpression) (*StatementOrExpression, error) {
	if stmtOrExpr.Statement != nil {
		return ctx.registerStatement(stmtOrExpr.Statement)
	}

	if stmtOrExpr.Expression != nil {
		return ctx.registerExpression(stmtOrExpr.Expression)
	}

	return nil, fmt.Errorf("unknown StatementOrExpression: %#v", stmtOrExpr)
}

func (ctx *Context) registerExpression(expr *ast.Expression) (*StatementOrExpression, error) {
	if expr.FunctionInstantiation != nil {
		return ctx.registerFunctionDeclaration(expr.FunctionInstantiation)
	}

	e, err := expressionFromAst(ctx, expr)
	if err != nil {
		return nil, err
	}

	return ctx.CurrentScope.AddStmtOrExpr(&StatementOrExpression{Expression: e}), nil
}

func (ctx *Context) registerStatement(stmt *ast.Statement) (*StatementOrExpression, error) {
	if fnInst := stmt.FunctionInstantiation; fnInst != nil {
		return ctx.registerFunctionDeclaration(fnInst)
	}

	if letDecl := stmt.LetDecl; letDecl != nil {
		letDeclType, err := inferTypeFromAst(ctx, &letDecl.Value)
		if err != nil {
			return nil, err
		}

		ctx.CurrentScope.AddIdentifer(letDecl.Name, letDeclType)

		stmt, err := statementFromAst(ctx, stmt)
		if err != nil {
			return nil, err
		}

		return ctx.CurrentScope.AddStmt(stmt), nil
	}

	if stmt.ReturnStmt != nil {
		expr, err := expressionFromAst(ctx, stmt.ReturnStmt)
		if err != nil {
			return nil, err
		}

		return ctx.CurrentScope.AddStmt(&Statement{ReturnStmt: expr}), nil
	}

	if stmt.ExpressionStmt != nil {
		expr, err := expressionFromAst(ctx, stmt.ExpressionStmt)
		if err != nil {
			return nil, err
		}

		return ctx.CurrentScope.AddStmt(&Statement{ExpressionStmt: expr}), nil
	}

	return nil, errors.WithStack(fmt.Errorf("unknown statement type: %#v", stmt))
}

func fromAstFunctionParameters(ctx *Context, params []*ast.FunctionParameter) ([]*FunctionParameter, error) {
	results := make([]*FunctionParameter, 0)
	for _, p := range params {
		pType, err := fromAstTypeIdentifier(ctx, &p.Type)
		if err != nil {
			return nil, err
		}

		results = append(results, &FunctionParameter{
			Name: p.Name,
			Type: pType,
		})
	}

	return results, nil
}

func (ctx *Context) registerFunctionDeclaration(astFn *ast.FunctionInstantiation) (*StatementOrExpression, error) {
	explRtnType, err := fromAstTypeIdentifier(ctx, astFn.ReturnType)
	if err != nil {
		return nil, err
	}

	// Figure out our params.
	params, err := fromAstFunctionParameters(ctx, astFn.Parameters)
	if err != nil {
		return nil, err
	}

	// Build our type spec.
	fn := &Function{
		Name:               astFn.Name,
		Parameters:         params,
		ImplicitReturnType: nil,
		ExplicitReturnType: explRtnType,
	}

	// Add the type spec.
	if astFn.Name != nil {
		ctx.CurrentScope.AddIdentifer(*astFn.Name, &TypeSpec{Function: fn})
	}

	// Add the statement.
	//
	// We'll return this as the registered stmt/expr.
	result := ctx.CurrentScope.AddStmtOrExpr(&StatementOrExpression{
		Statement: &Statement{
			FunctionInstantiation: fn,
		},
	})

	ctx.EnterScope()

	// Add the params as idents in the current scope.
	for _, p := range fn.Parameters {
		ctx.CurrentScope.AddIdentifer(p.Name, p.Type)
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
			infType, err := inferType(ctx, let.Value)
			if err != nil {
				return nil, err
			}

			ctx.CurrentScope.AddIdentifer(let.Name, infType)

			continue
		}

		// If this is a return stmt, update the implicit return type.
		if stmt.ReturnStmt != nil {
			rtnStmtType, err := inferType(ctx, stmt.ReturnStmt)
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
		fn.ImplicitReturnType = &TypeSpec{TypeReference: strRef("void")}
	}

	ctx.ExitScope()

	return result, nil
}

func statementOrExpressionFromAst(ctx *Context, stmtOrExpr *ast.StatementOrExpression) (*StatementOrExpression, error) {
	if stmtOrExpr.Statement != nil {
		stmt, err := statementFromAst(ctx, stmtOrExpr.Statement)
		if err != nil {
			return nil, err
		}

		return &StatementOrExpression{Statement: stmt}, nil
	}

	if stmtOrExpr.Expression != nil {
		expr, err := expressionFromAst(ctx, stmtOrExpr.Expression)
		if err != nil {
			return nil, err
		}

		return &StatementOrExpression{Expression: expr}, nil
	}

	return nil, fmt.Errorf("unknown stmt or expression value: %#v", stmtOrExpr)
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

func fromAstAccessible(ctx *Context, accessable ast.Accessable) (*Accessable, error) {
	if accessable.Ident != nil {
		return &Accessable{Ident: accessable.Ident}, nil
	}

	if lit := accessable.LiteralType; lit != nil {
		if f := lit.FunctionType; f != nil {
			fn, err := functionFromAst(ctx, &ast.FunctionInstantiation{
				Parameters: f.Parameters,
				ReturnType: f.ReturnType,
			})

			if err != nil {
				return nil, err
			}

			return &Accessable{Type: &TypeSpec{Function: fn}}, nil
		}

		if lit.ObjectType != nil {
			obj, err := objectFromAst(ctx, lit.ObjectType.Fields)
			if err != nil {
				return nil, err
			}

			return &Accessable{Type: &TypeSpec{Object: obj}}, nil
		}
	}

	return nil, fmt.Errorf("unknown accessable type: %#v", accessable)
}

func expressionFromAst(ctx *Context, expr *ast.Expression) (*Expression, error) {
	if expr.Number != nil {
		if expr.Number.Integer != nil {
			return &Expression{Number: &Number{Integer: expr.Number.Integer}}, nil
		}

		return nil, fmt.Errorf("unknown number expression: %#v", *expr)
	}

	if expr.String != nil {
		return &Expression{String: expr.String}, nil
	}

	if objInst := expr.ObjectInstantiation; objInst != nil {
		fields := make([]*ObjectFieldInstantiation, len(objInst.Fields))
		for i, f := range objInst.Fields {
			valueExpr, err := expressionFromAst(ctx, &f.Value)
			if err != nil {
				return nil, err
			}

			t, err := inferTypeFromAst(ctx, &f.Value)
			if err != nil {
				return nil, err
			}

			fields[i] = &ObjectFieldInstantiation{
				Name:  f.Name,
				Type:  t,
				Value: valueExpr,
			}
		}

		return &Expression{ObjectInstantiation: &ObjectInstantiation{Fields: fields}}, nil
	}

	if fnInst := expr.FunctionInstantiation; fnInst != nil {
		fn, err := functionFromAst(ctx, fnInst)
		if err != nil {
			return nil, err
		}

		return &Expression{FunctionInstantiation: fn}, nil
	}

	if chainedObjOperation := expr.ChainedObjectOperation; chainedObjOperation != nil {
		op, err := chainedObjOperationFromAst(ctx, chainedObjOperation)
		if err != nil {
			return nil, err
		}

		return &Expression{ChainedObjectOperation: op}, nil
	}

	if expr.Ident != nil {
		return &Expression{Ident: expr.Ident}, nil
	}

	if expr.IdentAssignment != nil {
		value, err := expressionFromAst(ctx, &expr.IdentAssignment.Assignment.Value)
		if err != nil {
			return nil, err
		}

		return &Expression{
			IdentAssignment: &IdentAssignment{
				Ident: expr.IdentAssignment.Ident,
				Assignment: Assignment{
					Value: value,
				},
			},
		}, nil
	}

	return nil, fmt.Errorf("unknown expression type: %#v", expr)
}

func (ctx *Context) registerInterface(i *ast.InterfaceDefinition) error {
	members := make([]*InterfaceMember, len(i.Members))
	for i, m := range i.Members {
		var member *InterfaceMember
		if m.Field != nil {
			t, err := fromAstTypeIdentifier(ctx, &m.Field.Type)
			if err != nil {
				return err
			}

			member = &InterfaceMember{
				Field: &InterfaceField{
					Name: m.Field.Name,
					Type: t,
				},
			}
		} else if m.Method != nil {
			t, err := fromAstTypeIdentifier(ctx, m.Method.ReturnType)
			if err != nil {
				return err
			}

			member = &InterfaceMember{
				Method: &InterfaceMethod{
					Name:       m.Field.Name,
					ReturnType: t,
				},
			}
		}

		members[i] = member
	}

	ctx.CurrentScope.AddType(&TypeSpec{
		Interface: &Interface{
			Name:    i.Name,
			Members: members,
		},
	})

	return nil
}

func statementFromAst(ctx *Context, stmt *ast.Statement) (*Statement, error) {
	if stmt.ExpressionStmt != nil {
		expr, err := expressionFromAst(ctx, stmt.ExpressionStmt)
		if err != nil {
			return nil, err
		}

		return &Statement{ExpressionStmt: expr}, nil
	}

	if fn := stmt.FunctionInstantiation; fn != nil {
		fn, err := functionFromAst(ctx, fn)
		if err != nil {
			return nil, err
		}

		return &Statement{FunctionInstantiation: fn}, nil
	}

	if stmt.LetDecl != nil {
		value, err := expressionFromAst(ctx, &stmt.LetDecl.Value)
		if err != nil {
			return nil, err
		}

		return &Statement{
			LetDecl: &LetDecl{
				Name:  stmt.LetDecl.Name,
				Value: value,
			},
		}, nil
	}

	if stmt.ReturnStmt != nil {
		rtn, err := expressionFromAst(ctx, stmt.ReturnStmt)
		if err != nil {
			return nil, err
		}

		return &Statement{ReturnStmt: rtn}, nil
	}

	return nil, fmt.Errorf("unknown statement: %#v", stmt)
}
