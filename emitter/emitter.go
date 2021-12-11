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
	bufferedWriter := bufio.NewWriter(e.writer)

	context := Context{
		Output: bufferedWriter,
	}

	context.EnterScope()

	writeRuntime(&context)

	for _, function := range ast.Functions {
		err := writeFunction(&context, &function)
		if err != nil {
			return err
		}

		context.WriteString("\n\n")
	}

	context.ExitScope()

	return bufferedWriter.Flush()
}

func New(writer io.Writer) Emitter {
	return &emitter{writer: writer}
}

func writeRuntime(ctx *Context) {
	ctx.WriteString(`
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

func writeStatementOrExpression(ctx *Context, stmtOrExpr *ast.StatementOrExpression) error {
	if stmtOrExpr.Statement != nil {
		return writeStatement(ctx, stmtOrExpr.Statement)
	}

	if stmtOrExpr.Expression != nil {
		return writeExpression(ctx, stmtOrExpr.Expression)
	}

	return fmt.Errorf("Unknown StatementOrExpression: %#v", stmtOrExpr)
}

func writeStatement(ctx *Context, stmt *ast.Statement) error {
	if letDecl := stmt.LetDecl; letDecl != nil {
		letDeclType, err := inferType(ctx, &letDecl.Value)
		if err != nil {
			return err
		}

		ctx.WriteString(fmt.Sprintf("%s* %s = ", mangleTypeName(*letDeclType.NonUnionType.TypeReference), letDecl.Name))

		err = writeExpression(ctx, &letDecl.Value)
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

	return fmt.Errorf("Unknown statement type: %#v", stmt)
}

func writeExpression(ctx *Context, expr *ast.Expression) error {
	if expr.WrappedExpression != nil {
		return writeExpression(ctx, expr.WrappedExpression)
	}

	if expr.Ident != nil {
		ctx.WriteString(*expr.Ident)
		return nil
	}

	if expr.Number != nil {
		if expr.Number.Integer != nil {
			ctx.WriteString(fmt.Sprintf("ts_num_new(%d)", *expr.Number.Integer))
			return nil
		}

		return fmt.Errorf("Unknown number expression: %#v", *expr)
	}

	if expr.String != nil {
		ctx.WriteString(fmt.Sprintf("ts_string_new(\"%s\")", *expr.String))
		return nil
	}

	return fmt.Errorf("Unknown expression type: %#v", expr)
}

func mangleTypeName(name string) string {
	return fmt.Sprintf("ts_%s", name)
}

func mangleFunctionName(name string) string {
	return fmt.Sprintf("%s", name)
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

	if expr.WrappedExpression != nil {
		return inferType(ctx, expr.WrappedExpression)
	}

	if expr.Ident != nil {
		return ctx.TypeOf(*expr.Ident), nil
	}

	return nil, fmt.Errorf("Could not infer type of %#v", *expr)
}
