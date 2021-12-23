package emitter

import (
	"bufio"
	"bytes"
	"elauffenburger/hypescript/ast"
	"embed"
	"io"
	"reflect"

	"github.com/pkg/errors"
)

//go:embed .runtime/*
var runtimeFiles embed.FS

type primitiveType string

const (
	TsString primitiveType = "string"
	TsNumber primitiveType = "number"
	TsVoid   primitiveType = "void"
)

var primitiveTypes = []primitiveType{TsString, TsNumber, TsVoid}

type runtimeType string

const (
	RtTsObject   runtimeType = "TsObject"
	RtTsFunction runtimeType = "TsFunction"
	RtTsVoid     runtimeType = "void"
)

type typeId int

const (
	TypeIdNone       typeId = 0
	TypeIdTsObject   typeId = 1
	TypeIdTsNum      typeId = 2
	TypeIdTsString   typeId = 3
	TypeIdTsFunction typeId = 4
	TypeIdVoid       typeId = 5
	TypeIdIntrinsic  typeId = 6
)

type Emitter interface {
	Emit(ast *ast.TS) ([]EmittedFile, error)
}

type TypeSpec struct {
	Function      *Function
	Object        *Object
	Interface     *Interface
	TypeReference *string
	Primitive     *primitiveType
	Union         *Union

	Unresolved bool
	Source     interface{}
}

type Function struct {
	Name               *string
	Parameters         []*FunctionParameter
	ImplicitReturnType *TypeSpec
	ExplicitReturnType *TypeSpec
	Body               []*StatementOrExpression
}

type StatementOrExpression struct {
	Statement  *Statement
	Expression *Expression

	Scope *Scope
}

type Expression struct {
	Number                 *Number
	String                 *string
	IdentAssignment        *IdentAssignment
	FunctionInstantiation  *Function
	ChainedObjectOperation *ChainedObjectOperation
	ObjectInstantiation    *ObjectInstantiation
	Ident                  *string
}

type Object struct {
	Fields []*ObjectTypeField
}

type ObjectTypeField struct {
	Name string
	Type *TypeSpec
}

type Interface struct {
	Name    string
	Members []*InterfaceMember
}

type InterfaceMember struct {
	Field  *InterfaceField
	Method *InterfaceMethod
}

type InterfaceField struct {
	Name string
	Type *TypeSpec
}

type InterfaceMethod struct {
	Name       string
	Parameters []*FunctionParameter
	ReturnType *TypeSpec
}

type FunctionParameter struct {
	Name string
	Type *TypeSpec
}

type Union struct {
	Head *TypeSpec
	Tail []*TypeSpec
}

type IdentAssignment struct {
	Ident      string
	Assignment Assignment
}

type ChainedObjectOperation struct {
	First *ObjectOperation
	Last  *ObjectOperation
}

type ObjectOperation struct {
	Accessee *Accessable

	Access     *ObjectAccess
	Invocation *ObjectInvocation
	Assignment *Assignment

	Next *ObjectOperation
	Prev *ObjectOperation
}

type ObjectInvocation struct {
	Accessee  *Accessable
	Arguments []*Expression
}

type ObjectAccess struct {
	AccessedIdent string
}

type Accessable struct {
	Ident *string
	Type  *TypeSpec
}

type Number struct {
	Integer *int
}

type LetDecl struct {
	Name  string
	Value *Expression
}

type Statement struct {
	FunctionInstantiation *Function
	ExpressionStmt        *Expression
	LetDecl               *LetDecl
	ReturnStmt            *Expression
}

type Assignment struct {
	Value *Expression
}

type ObjectInstantiation struct {
	Fields []*ObjectFieldInstantiation
}

type ObjectFieldInstantiation struct {
	Name  string
	Type  *TypeSpec
	Value *Expression
}

type TopLevelConstruct struct {
	InterfaceDefinition   *Interface
	StatementOrExpression *StatementOrExpression
}

type TS struct {
	TopLevelConstructs []TopLevelConstruct
}

func (t *TypeSpec) Equals(other *TypeSpec) bool {
	return reflect.DeepEqual(t, other)
}

type emitter struct{}

type EmittedFile struct {
	Filename string
	Contents io.Reader
}

func (e emitter) buildContext(ast *ast.TS, output *bufio.Writer) (*Context, error) {
	ctx := NewContext(output)

	for _, c := range ast.TopLevelConstructs {
		if c.StatementOrExpression != nil {
			_, err := ctx.registerStatementOrExpression(c.StatementOrExpression)
			if err != nil {
				return nil, err
			}

			continue
		}

		if intdef := c.InterfaceDefinition; intdef != nil {
			ctx.registerInterface(intdef)

			continue
		}

		return nil, errors.Errorf("unknown top-level construct: %v", c)
	}

	return ctx, nil
}

func (e emitter) Emit(ast *ast.TS) ([]EmittedFile, error) {
	buf := bytes.Buffer{}
	rw := bufio.NewReadWriter(bufio.NewReader(&buf), bufio.NewWriter(&buf))

	runtime, err := emitRuntime()
	if err != nil {
		return nil, errors.Wrap(err, "failed to write runtime")
	}

	ctx, err := e.buildContext(ast, rw.Writer)
	if err != nil {
		return nil, err
	}

	if err = writePreamble(ctx); err != nil {
		return nil, err
	}

	ctx.WriteString("int main() { ")

	err = writeScope(ctx, ctx.CurrentScope)
	if err != nil {
		return nil, err
	}

	ctx.WriteString(`return 0; }`)

	if err = ctx.Output.Flush(); err != nil {
		return nil, errors.Wrap(err, "failed to write main.cpp")
	}

	return append(runtime, EmittedFile{Filename: "main.cpp", Contents: rw.Reader}), nil
}

func writeScope(ctx *Context, s *Scope) error {
	// Write statements & expressions.
	for _, stmtOrExpr := range s.StatementsOrExpressions {
		if err := writeStatementOrExpression(ctx, stmtOrExpr); err != nil {
			return err
		}
	}

	return nil
}

func New() Emitter {
	return &emitter{}
}

func emitRuntime() ([]EmittedFile, error) {
	header, err := runtimeFiles.Open(".runtime/runtime.hpp")
	if err != nil {
		return nil, errors.Wrap(err, "could not write runtime header")
	}

	runtime, err := runtimeFiles.Open(".runtime/runtime.cpp")
	if err != nil {
		return nil, errors.Wrap(err, "could not write runtime code")
	}

	return []EmittedFile{
		{Filename: "runtime.hpp", Contents: header},
		{Filename: "runtime.cpp", Contents: runtime},
	}, nil
}

func writePreamble(ctx *Context) error {
	ctx.WriteString(`
		#include <stdlib.h>
		#include <stdio.h>
		#include <string>
		#include <vector>
		#include <algorithm>
		#include <memory>
	
		#include "runtime.hpp"
	`)

	return nil
}
