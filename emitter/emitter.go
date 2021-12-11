package emitter

import (
	"bufio"
	"elauffenburger/hypescript/ast"
	"fmt"
	"io"
)

type primitiveType string

const (
	TsString primitiveType = "string"
	TsNum    primitiveType = "num"
	TsVoid   primitiveType = "void"
)

type Emitter interface {
	Emit(ast *ast.TS) error
}

type emitter struct {
	writer io.Writer
}

func (e emitter) Emit(ast *ast.TS) error {
	context := NewContext(bufio.NewWriter(e.writer))

	writeRuntime(context)

	for _, function := range ast.Functions {
		err := writeFunction(context, &function)
		if err != nil {
			return err
		}

		context.WriteString("\n\n")
	}

	return context.Output.Flush()
}

func New(writer io.Writer) Emitter {
	return &emitter{writer: writer}
}

func writeRuntime(ctx *Context) {
	ctx.WriteString(`
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

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

void ts_Console_log(ts_string* str) {
	printf("%s\n", str->value);
}

typedef struct ts_Console {
	void (*log)(ts_string* str);
} ts_Console;

ts_Console* console;

static void init_globals() {
	console = (ts_Console*)malloc(sizeof(ts_Console));
	console->log = ts_Console_log;
}

void ts_main();

int main() {
	init_globals();

	ts_main();

	return 0;
}

`)
}

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

		ctx.WriteString(fmt.Sprintf("%s %s = ", mangleTypeName(*letDeclType.NonUnionType.TypeReference), letDecl.Name))

		err = writeExpression(ctx, &letDecl.Value)
		if err != nil {
			return err
		}

		ctx.WriteString(";")

		ctx.CurrentScope.AddIdentifer(letDecl.Name, *letDeclType)

		return nil
	}

	if assignmentStmt := stmt.AssignmentStmt; assignmentStmt != nil {
		ctx.WriteString(fmt.Sprintf("%s = ", assignmentStmt.Ident))

		err := writeExpression(ctx, &assignmentStmt.Value)
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

	if expr.Invocation != nil {
		err := writeInvocable(ctx, expr.Invocation.Invoked)
		if err != nil {
			return err
		}

		ctx.WriteString("(")

		numArgs := len(expr.Invocation.Arguments)
		for i, arg := range expr.Invocation.Arguments {
			err := writeExpression(ctx, &arg)
			if err != nil {
				return err
			}

			if i != numArgs-1 {
				ctx.WriteString(", ")
			}
		}

		ctx.WriteString(")")

		return nil
	}

	if objectAccess := expr.ObjectAccess; objectAccess != nil {
		err := writeAccessable(ctx, objectAccess.Accessee)
		if err != nil {
			return nil
		}

		ctx.WriteString("->")

		return writeExpression(ctx, &objectAccess.AccessedValue)
	}

	if expr.Ident != nil {
		return writeIdent(ctx, *expr.Ident)
	}

	return fmt.Errorf("unknown expression type: %#v", expr)
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

func writeInvocable(ctx *Context, invocable ast.Invocable) error {
	if invocable.Ident != nil {
		return writeIdent(ctx, *invocable.Ident)
	}

	return fmt.Errorf("unknown invocable type: %#v", invocable)
}

func writeAccessable(ctx *Context, accessable ast.Accessable) error {
	if accessable.Ident != nil {
		ctx.WriteString(*accessable.Ident)

		return nil
	}

	return fmt.Errorf("unknown accessable type: %#v", accessable)
}

func mangleTypeName(name string) string {
	if name == "void" {
		return name
	}

	return fmt.Sprintf("ts_%s*", name)
}

func mangleFunctionName(name string) string {
	return fmt.Sprintf("ts_%s", name)
}

func mangleIdentName(name string, identType *ast.Type) string {
	if identType.NonUnionType != nil {
		if identType.NonUnionType.LiteralType != nil {
			if identType.NonUnionType.LiteralType.FunctionType != nil {
				return mangleFunctionName(name)
			}
		}
	}

	return name
}

func inferType(ctx *Context, expr *ast.Expression) (*ast.Type, error) {
	// DO NOT SUBMIT -- need to actually impl!

	if expr.String != nil {
		t := string(TsString)
		return &ast.Type{NonUnionType: &ast.NonUnionType{TypeReference: &t}}, nil
	}

	if expr.Number != nil {
		t := string(TsNum)
		return &ast.Type{NonUnionType: &ast.NonUnionType{TypeReference: &t}}, nil
	}

	if expr.Ident != nil {
		return ctx.TypeOf(*expr.Ident)
	}

	return nil, fmt.Errorf("could not infer type of %#v", *expr)
}
