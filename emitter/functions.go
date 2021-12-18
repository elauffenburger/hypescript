package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type functionInfo struct {
	Function *ast.Function

	ExplicitReturnType *ast.Type
	ImplicitReturnType *ast.Type
}

func buildFunctionInfo(context *Context, function *ast.Function) (*functionInfo, error) {
	functionInfo := functionInfo{Function: function, ExplicitReturnType: function.ReturnType}

	context.EnterScope()

	// TODO: need to add the function as a known identifier to the current scope.

	for _, param := range function.Parameters {
		context.CurrentScope.AddIdentifer(param.Name, param.Type)
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
				return nil, errors.Wrap(err, "failed to infer type for value of let decl")
			}

			context.CurrentScope.AddIdentifer(stmt.LetDecl.Name, *inferredType)

			continue
		}

		// If this is a return stmt, update the implicit return type.
		if stmt.ReturnStmt != nil {
			returnStmtType, err := inferType(context, stmt.ReturnStmt)
			if err != nil {
				return nil, errors.Wrap(err, "failed to infer type for return statement")
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

func writeFunction(context *Context, function *ast.Function) error {
	functionInfo, err := buildFunctionInfo(context, function)
	if err != nil {
		return errors.Wrap(err, "failed to build function info")
	}

	if functionInfo.ExplicitReturnType != nil {
		if !functionInfo.ExplicitReturnType.Equals(functionInfo.ImplicitReturnType) {
			return fmt.Errorf("implicit and explicit return types of function were not the same: %#v", *functionInfo)
		}
	}

	returnType := functionInfo.ImplicitReturnType

	// TODO: this is super not guaranteed to be correct!
	mangledReturnTypeName := mangleTypeNamePtr(*returnType.NonUnionType.TypeReference)
	mangledFunctionName := mangleFunctionName(function.Name)

	context.CurrentScope.AddIdentifer(function.Name, ast.Type{
		NonUnionType: &ast.NonUnionType{
			LiteralType: &ast.LiteralType{
				FunctionType: &ast.FunctionType{
					Parameters: function.Parameters,
					ReturnType: returnType,
				},
			},
		},
	})

	context.EnterScope()

	formattedArgs := strings.Builder{}
	numArgs := len(function.Parameters)
	for i, arg := range function.Parameters {
		typeName, argName := mangleTypeNamePtr(*arg.Type.NonUnionType.TypeReference), arg.Name

		formattedArgs.WriteString(fmt.Sprintf("%s %s", typeName, argName))

		if i != numArgs-1 {
			formattedArgs.WriteString(", ")
		}

		context.CurrentScope.AddIdentifer(arg.Name, arg.Type)
	}

	context.WriteString(fmt.Sprintf("%s %s(%s) {\n", mangledReturnTypeName, mangledFunctionName, formattedArgs.String()))

	for _, statementOrExpression := range function.Body {
		context.WriteString("\t")

		err := writeStatementOrExpression(context, &statementOrExpression)
		if err != nil {
			return errors.Wrap(err, "failed to write function body")
		}

		context.WriteString("\n")
	}

	context.WriteString("}")

	context.ExitScope()

	return nil
}
