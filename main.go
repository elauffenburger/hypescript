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
	Body       []ExpressionOrStatement `"{"@@*"}"`
}

type Type struct {
	TypeName string `@Ident`
}

type FunctionArgument struct {
	Name string `@Ident`
	Type Type   `":" @@`
}

type ExpressionOrStatement struct {
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

func (function *Function) ReturnStatements() []*Statement {
	var stmts []*Statement

	for _, statementOrExpression := range function.Body {
		if statementOrExpression.Statement == nil || statementOrExpression.Statement.ReturnStmt == nil {
			continue
		}

		stmts = append(stmts, statementOrExpression.Statement)
	}

	return stmts
}

func (function *Function) InferReturnType(context *Context) *Type {
	var impliedReturnType *Type
	if function.ReturnType != nil {
		impliedReturnType = function.ReturnType
	}

	returnStmts := function.ReturnStatements()

	// If the function didn't specify a return value and there aren't any returns, assume the type is void.
	if impliedReturnType == nil && len(returnStmts) == 0 {
		return &Type{TypeName: string(TsVoid)}
	}

	// Otherwise, we need to make sure the return types line up.
	for _, stmt := range returnStmts {
		inferredType := inferType(context, stmt.ReturnStmt)

		if impliedReturnType == nil {
			impliedReturnType = inferredType
			continue
		}

		if inferredType != impliedReturnType {
			panic(fmt.Sprintf("Function return types didn't line up for: %#v", function))
		}
	}

	return impliedReturnType
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

func writeFunction(context *Context, function *Function) {
	returnType := function.InferReturnType(context)
	returnTypeName, functionName := mangleTypeName(returnType.TypeName), mangleFunctionName(function.Name)

	// DO NOT SUBMIT: need to add the function as a known identifier to the current scope.

	context.EnterScope()

	formattedArgs := strings.Builder{}
	numArgs := len(function.Arguments)
	for i, arg := range function.Arguments {
		typeName, argName := mangleTypeName(arg.Type.TypeName), arg.Name

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

func writeStatementOrExpression(context *Context, expressionOrStatement *ExpressionOrStatement) {
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

		context.WriteString(fmt.Sprintf("%s* %s = ", mangleTypeName(letDeclType.TypeName), letDecl.Name))
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
		return &Type{TypeName: string(TsString)}
	}

	if expression.Number != nil {
		return &Type{TypeName: string(TsNum)}
	}

	if expression.WrappedExpression != nil {
		return inferType(context, expression.WrappedExpression)
	}

	if expression.Ident != nil {
		return context.TypeOf(*expression.Ident)
	}

	panic(fmt.Sprintf("Could not infer type of %#v", *expression))
}
