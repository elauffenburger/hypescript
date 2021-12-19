package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
	"strings"
)

type functionInfo struct {
	Function *ast.FunctionInstantiation

	ExplicitReturnType *ast.Type
	ImplicitReturnType *ast.Type
}

func buildFunctionInfo(context *Context, function *ast.FunctionInstantiation) (*functionInfo, error) {
	functionInfo := functionInfo{Function: function, ExplicitReturnType: function.ReturnType}

	context.EnterScope()

	// TODO: need to add the function as a known identifier to the current scope.

	for _, param := range function.Parameters {
		context.CurrentScope.AddIdentifer(param.Name, &param.Type)
	}

	for _, stmtOrExpr := range function.Body {
		// Expressions can't produce identifiers, so we can skip them.
		if stmtOrExpr.Statement == nil {
			continue
		}

		stmt := stmtOrExpr.Statement

		// If this is a let decl, add the ident to the current scope.
		if stmt.LetDecl != nil {
			inferredType, err := inferType(context, &stmt.LetDecl.Value)
			if err != nil {
				return nil, err
			}

			context.CurrentScope.AddIdentifer(stmt.LetDecl.Name, inferredType)

			continue
		}

		// If this is a return stmt, update the implicit return type.
		if stmt.ReturnStmt != nil {
			returnStmtType, err := inferType(context, stmt.ReturnStmt)
			if err != nil {
				return nil, err
			}

			// If we don't have an implied type yet, use this return statement's.
			if functionInfo.ImplicitReturnType == nil {
				functionInfo.ImplicitReturnType = returnStmtType
				continue
			}

			// ...otherwise, if the return types match, bail out.
			if functionInfo.ImplicitReturnType == returnStmtType {
				continue
			}

			// ...otherwise, we need to treat this as a union of the existing type and this type.
			functionInfo.ImplicitReturnType = ast.CreateUnionType(functionInfo.ImplicitReturnType, returnStmtType)
		}
	}

	if functionInfo.ImplicitReturnType == nil {
		functionInfo.ImplicitReturnType = &ast.Type{
			NonUnionType: &ast.NonUnionType{TypeReference: strRef("void")},
		}
	}

	context.ExitScope()

	return &functionInfo, nil
}

func (f *functionInfo) validate() error {
	// Make sure the implicit return type matches the implicit one (if any).
	if rtnType := f.ExplicitReturnType; rtnType != nil {
		if !rtnType.Equals(f.ImplicitReturnType) {
			return fmt.Errorf("implicit and explicit return types of function were not the same: %#v", *f)
		}
	}

	return nil
}

func writeFunctionDeclaration(ctx *Context, fn *ast.FunctionInstantiation) error {
	// Build the complete info struct for this function.
	fnInfo, err := buildFunctionInfo(ctx, fn)
	if err != nil {
		return err
	}

	if err = fnInfo.validate(); err != nil {
		return err
	}

	returnType := fnInfo.ImplicitReturnType

	if fn.Name != nil {
		ctx.CurrentScope.AddIdentifer(*fn.Name, &ast.Type{
			NonUnionType: &ast.NonUnionType{
				LiteralType: &ast.LiteralType{
					FunctionType: &ast.FunctionType{
						Parameters: fn.Parameters,
						ReturnType: returnType,
					},
				},
			},
		})
	}

	return ctx.WithinNewScope(func() error {
		var fnName string
		if fn.Name != nil {
			fnName = *fn.Name
		} else {
			fnName = ctx.CurrentScope.NewIdent()
		}

		ctx.WriteString(fmt.Sprintf("TsFunction* %s = ", mangleFunctionName(fnName)))

		err = writeFunction(ctx, fn, fnInfo)
		if err != nil {
			return err
		}

		ctx.WriteString(";")

		return nil
	})
}

func writeFunction(ctx *Context, fn *ast.FunctionInstantiation, fnInfo *functionInfo) error {
	// Format the function params.
	formattedParams := strings.Builder{}
	formattedParams.WriteString("TsCoreHelpers::toVector<TsFunctionParam>({")

	numParams := len(fn.Parameters)
	for i, param := range fn.Parameters {
		typeId, err := getTypeIdFor(ctx, &param.Type)
		if err != nil {
			return err
		}

		formattedParams.WriteString(fmt.Sprintf("TsFunctionParam(\"%s\", %d)", param.Name, typeId))

		if i != numParams-1 {
			formattedParams.WriteString(", ")
		}
	}

	formattedParams.WriteString("})")

	var fnName string
	if fn.Name != nil {
		fnName = *fn.Name
	} else {
		fnName = ctx.CurrentScope.NewIdent()
	}

	ctx.WriteString(
		fmt.Sprintf(
			"new TsFunction(\"%s\", %s, ",
			fnName,
			formattedParams.String(),
		),
	)

	err := writeFunctionLambda(ctx, fn, fnInfo)
	if err != nil {
		return err
	}

	ctx.WriteString(")")

	return nil
}

func writeFunctionLambda(ctx *Context, fn *ast.FunctionInstantiation, fnInfo *functionInfo) error {
	ctx.WriteString("[=](std::vector<TsFunctionArg> args) -> TsObject* {")

	// Unpack each arg into local vars in the function.
	for _, param := range fn.Parameters {
		ctx.WriteString(fmt.Sprintf("auto %s = (TsFunction*)TsFunctionArg::findArg(args, \"%s\").value;", param.Name, param.Name))
	}

	// Write the body.
	for _, exprOrStmt := range fn.Body {
		err := writeStatementOrExpression(ctx, &exprOrStmt)
		if err != nil {
			return err
		}
	}

	if t := fnInfo.ImplicitReturnType.NonUnionType; t != nil {
		if t := t.TypeReference; t != nil && *t == string(TsVoid) {
			ctx.WriteString("return NULL;")
		}
	}

	ctx.WriteString("}")

	return nil
}
