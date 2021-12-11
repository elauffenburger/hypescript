package main

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle"
)

type primitiveType string

const (
	TsString primitiveType = "string"
	TsNum                  = "num"
	TsVoid                 = "void"
)

type Context struct {
	scopes       []Scope
	CurrentScope *Scope

	Output strings.Builder
}

func (context *Context) WriteString(str string) {
	context.Output.WriteString(str)
}

func (context *Context) String() string {
	return context.Output.String()
}

type Scope struct {
	IdentTypes map[string]Type
}

func (scope *Scope) TypeOf(ident string) *Type {
	t, ok := scope.IdentTypes[ident]
	if !ok {
		panic(fmt.Sprintf("Unknown identifier %s in scope: %#v", ident, scope))
	}

	return &t
}

func (context *Context) TypeOf(ident string) *Type {
	return context.CurrentScope.TypeOf(ident)
}

type Function struct {
	Name       string                  `"function" @Ident`
	Arguments  []FunctionArgument      `"(" (@@ ("," @@)*)? ")"`
	ReturnType *Type                   `(":" @@)?`
	Body       []StatementOrExpression `"{"@@*"}"`
}

type Type struct {
	TypeName  *string    `@Ident`
	UnionType *UnionType `| @@`
}

type UnionType struct {
	Types []*Type `@@ | @@ ("|" @@)*`
}

type FunctionArgument struct {
	Name string `@Ident`
	Type Type   `":" @@`
}

type StatementOrExpression struct {
	Statement  *Statement  `@@`
	Expression *Expression `| @@`
}

type Expression struct {
	WrappedExpression *Expression `"("@@")"`
	Number            *Number     `| @@`
	String            *string     `| @String`
	Ident             *string     `| @Ident`
}

type Number struct {
	Integer *int `@Int`
}

type LetDecl struct {
	Name  string     `"let" @Ident`
	Value Expression `"=" @@ ";"`
}

type Statement struct {
	LetDecl    *LetDecl    `@@`
	ReturnStmt *Expression `| "return" @@ ";"`
}

type TS struct {
	Functions []Function `@@*`
}

func main() {
	parser, err := participle.Build(&TS{})
	if err != nil {
		panic(fmt.Errorf("building parser failed: %w", err))
	}

	ast := &TS{}
	err = parser.ParseString(`
		function foo(a: string, b: num): num {
			let foo = 5;
			let bar = "bar";

			return foo;
		}

		function blah() {
			let foo = "asdf";

			return foo;
		}

		function main(): num {
			return 0;
		}
	`, ast)

	if err != nil {
		panic(fmt.Errorf("parsing failed: %w", err))
	}

	context := Context{
		Output: strings.Builder{},
	}

	context.EnterScope()

	writeCoreTypes(&context)

	for _, function := range ast.Functions {
		writeFunction(&context, &function)

		context.WriteString("\n\n")
	}

	context.ExitScope()

	fmt.Println(context.String())
}

func writeCoreTypes(context *Context) {
	context.WriteString(`
#include <stdlib.h>
#include <string.h>

typedef struct ts_num {
	int value;
} ts_num;

typedef struct ts_string {
	char* value;
	int len;
} ts_string;

ts_num* ts_num_new(int num) {
	ts_num* result = (ts_num*)malloc(sizeof(ts_num));
	result->value = num;

	return result;
}

ts_string* ts_string_new(char* str) {
	ts_string* result = (ts_string*)malloc(sizeof(ts_string));
	result->value = str;
	result->len = strlen(str);

	return result;
}

`)
}

func (context *Context) EnterScope() {
	newScope := Scope{
		IdentTypes: make(map[string]Type),
	}

	context.scopes = append(context.scopes, newScope)
	context.CurrentScope = &newScope
}

func (context *Context) ExitScope() {
	context.scopes = context.scopes[:len(context.scopes)-1]

	if len(context.scopes) == 0 {
		context.CurrentScope = nil
		return
	}

	context.CurrentScope = &context.scopes[len(context.scopes)-1]
}

func (scope *Scope) AddIdentifer(ident string, identType Type) {
	scope.IdentTypes[ident] = identType
}

type FunctionBuilder struct {
	Context  *Context
	Function *Function

	ExplicitReturnType *Type
	ImpliedReturnType  *Type
}

type FunctionInfo struct {
	Function *Function

	ExplicitReturnType *Type
	ImplicitReturnType *Type
}

func buildFunctionInfo(context *Context, function *Function) *FunctionInfo {
	functionInfo := FunctionInfo{Function: function, ExplicitReturnType: function.ReturnType}

	context.EnterScope()

	// DO NOT SUBMIT: need to add the function as a known identifier to the current scope.

	for _, arg := range function.Arguments {
		context.CurrentScope.AddIdentifer(arg.Name, arg.Type)
	}

	for _, stmtOrExpr := range function.Body {
		// Expressions can't produce identifiers, so we can skip them.
		if stmtOrExpr.Statement == nil {
			continue
		}

		stmt := stmtOrExpr.Statement

		// If this is a let decl, add the ident to the current scope.
		if stmt.LetDecl != nil {
			context.CurrentScope.AddIdentifer(stmt.LetDecl.Name, *inferType(context, &stmt.LetDecl.Value))

			continue
		}

		// If this is a return stmt, update the implicit return type.
		if stmt.ReturnStmt != nil {
			returnStmtType := inferType(context, stmt.ReturnStmt)

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
			functionInfo.ImplicitReturnType = createUnionType(functionInfo.ImplicitReturnType, returnStmtType)
		}
	}

	context.ExitScope()

	return &functionInfo
}

func createUnionType(left, right *Type) *Type {
	return &Type{UnionType: &UnionType{Types: []*Type{left, right}}}
}

func (left *Type) Equals(right *Type) bool {
	if left.TypeName != nil && right.TypeName != nil {
		return *left.TypeName == *right.TypeName
	}

	panic(fmt.Sprintf("Unsupported type comparison between %#v and %#v", left, right))
}

func writeFunction(context *Context, function *Function) {
	functionInfo := buildFunctionInfo(context, function)

	if functionInfo.ExplicitReturnType != nil {
		if !functionInfo.ExplicitReturnType.Equals(functionInfo.ImplicitReturnType) {
			panic(fmt.Sprintf("implicit and explicit return types of function were not the same: %#v", *functionInfo))
		}
	}

	returnTypeName := mangleTypeName(*functionInfo.ImplicitReturnType.TypeName)
	functionName := mangleFunctionName(function.Name)

	// DO NOT SUBMIT: need to add the function as a known identifier to the current scope.

	context.EnterScope()

	formattedArgs := strings.Builder{}
	numArgs := len(function.Arguments)
	for i, arg := range function.Arguments {
		typeName, argName := mangleTypeName(*arg.Type.TypeName), arg.Name

		formattedArgs.WriteString(fmt.Sprintf("%s* %s", typeName, argName))

		if i != numArgs-1 {
			formattedArgs.WriteString(", ")
		}

		context.CurrentScope.AddIdentifer(arg.Name, arg.Type)
	}

	context.WriteString(fmt.Sprintf("%s* %s(%s) {\n", returnTypeName, functionName, formattedArgs.String()))

	for _, statementOrExpression := range function.Body {
		context.WriteString("\t")
		writeStatementOrExpression(context, &statementOrExpression)
		context.WriteString("\n")
	}

	context.WriteString("}")

	context.ExitScope()
}

func writeStatementOrExpression(context *Context, expressionOrStatement *StatementOrExpression) {
	if expressionOrStatement.Statement != nil {
		writeStatement(context, expressionOrStatement.Statement)
		return
	}

	if expressionOrStatement.Expression != nil {
		writeExpression(context, expressionOrStatement.Expression)
	}
}

func writeStatement(context *Context, statement *Statement) {
	if letDecl := statement.LetDecl; letDecl != nil {
		letDeclType := inferType(context, &letDecl.Value)

		context.WriteString(fmt.Sprintf("%s* %s = ", mangleTypeName(*letDeclType.TypeName), letDecl.Name))
		writeExpression(context, &letDecl.Value)
		context.WriteString(";")

		return
	}

	if returnStmt := statement.ReturnStmt; returnStmt != nil {
		context.WriteString("return ")
		writeExpression(context, returnStmt)
		context.WriteString(";")

		return
	}
}

func writeExpression(context *Context, expression *Expression) {
	if expression.WrappedExpression != nil {
		writeExpression(context, expression.WrappedExpression)
		return
	}

	if expression.Ident != nil {
		context.WriteString(*expression.Ident)
		return
	}

	if expression.Number != nil {
		if expression.Number.Integer != nil {
			context.WriteString(fmt.Sprintf("ts_num_new(%d)", *expression.Number.Integer))
			return
		}

		panic(fmt.Sprintf("Unknown number expression: %#v", *expression))
	}

	if expression.String != nil {
		context.WriteString(fmt.Sprintf("ts_string_new(\"%s\")", *expression.String))
		return
	}
}

func mangleTypeName(typeName string) string {
	return fmt.Sprintf("ts_%s", typeName)
}

func mangleFunctionName(functionName string) string {
	return fmt.Sprintf("%s", functionName)
}

func inferType(context *Context, expression *Expression) *Type {
	// DO NOT SUBMIT -- need to actually impl!

	if expression.String != nil {
		t := string(TsString)
		return &Type{TypeName: &t}
	}

	if expression.Number != nil {
		t := string(TsNum)
		return &Type{TypeName: &t}
	}

	if expression.WrappedExpression != nil {
		return inferType(context, expression.WrappedExpression)
	}

	if expression.Ident != nil {
		return context.TypeOf(*expression.Ident)
	}

	panic(fmt.Sprintf("Could not infer type of %#v", *expression))
}
