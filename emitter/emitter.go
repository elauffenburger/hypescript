package emitter

import (
	"bufio"
	"bytes"
	"elauffenburger/hypescript/ast"
	"embed"
	"io"

	"github.com/pkg/errors"
)

//go:embed .runtime/*
var runtimeFiles embed.FS

type primitiveType string

const (
	TsString primitiveType = "string"
	TsNumber primitiveType = "number"
)

type coreType string

const (
	TsObject   coreType = "TsObject"
	TsFunction coreType = "TsFunction"
	TsVoid     coreType = "void"
)

type typeId int

const (
	TypeIdNone       typeId = 0
	TypeIdTsObject   typeId = 2
	TypeIdTsNum      typeId = 2
	TypeIdTsString   typeId = 3
	TypeIdTsFunction typeId = 4
	TypeIdVoid       typeId = 5
	TypeIdIntrinsic  typeId = 6
)

var coreTypes = []coreType{TsObject, TsFunction, TsVoid}

type Emitter interface {
	Emit(ast *ast.TS) ([]EmittedFile, error)
}

type emitter struct{}

type EmittedFile struct {
	Filename string
	Contents io.Reader
}

func (e emitter) Emit(ast *ast.TS) ([]EmittedFile, error) {
	contentBuffer := bytes.Buffer{}
	contentWriter := bufio.NewReadWriter(bufio.NewReader(&contentBuffer), bufio.NewWriter(&contentBuffer))

	runtime, err := emitRuntime()
	if err != nil {
		return nil, errors.Wrap(err, "failed to write runtime")
	}

	ctx := NewContext(contentWriter.Writer)

	err = writePreamble(ctx)
	if err != nil {
		return nil, err
	}

	ctx.WriteString("int main() { ")

	for _, c := range ast.TopLevelConstructs {
		if c.StatementOrExpression != nil {
			err := writeStatementOrExpression(ctx, c.StatementOrExpression)
			if err != nil {
				return nil, err
			}
		}
	}

	ctx.WriteString(`return 0; }`)

	err = ctx.Output.Flush()
	if err != nil {
		return nil, errors.Wrap(err, "failed to write main.cpp")
	}

	return append(runtime, EmittedFile{Filename: "main.cpp", Contents: contentWriter.Reader}), nil
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
