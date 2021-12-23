package emitter

import (
	"bufio"
	"bytes"
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/core"
	"elauffenburger/hypescript/emitter/regpass"
	"elauffenburger/hypescript/emitter/writepass"
	"embed"
	"io"

	"github.com/pkg/errors"
)

//go:embed .runtime/*
var runtimeFiles embed.FS

type Emitter interface {
	Emit(ast *ast.TS) ([]*EmittedFile, error)
}

type emitter struct{}

type EmittedFile struct {
	Filename string
	Contents io.Reader
}

func New() Emitter {
	return &emitter{}
}

func (e emitter) Emit(ast *ast.TS) ([]*EmittedFile, error) {
	// Run the registration pass.
	regpassCtx, err := runRegPass(ast)
	if err != nil {
		return nil, err
	}

	// Run the write pass.
	result, err := runWritePass(regpassCtx.GlobalScope)
	if err != nil {
		return nil, err
	}

	// Emit the runtime.
	runtime, err := emitRuntime()
	if err != nil {
		return nil, errors.Wrap(err, "failed to write runtime")
	}

	// Return all emitted files.
	return append(runtime, result), nil
}

func runRegPass(ast *ast.TS) (*regpass.Context, error) {
	ctx := regpass.NewContext()

	return ctx, ctx.Run(ast)

}

func runWritePass(global *core.Scope) (*EmittedFile, error) {
	buf := bytes.Buffer{}
	rw := bufio.NewReadWriter(bufio.NewReader(&buf), bufio.NewWriter(&buf))

	ctx := writepass.NewContext(rw.Writer)
	err := ctx.Run(global)
	if err != nil {
		return nil, err
	}

	return &EmittedFile{Filename: "main.cpp", Contents: rw.Reader}, nil
}

func emitRuntime() ([]*EmittedFile, error) {
	header, err := runtimeFiles.Open(".runtime/runtime.hpp")
	if err != nil {
		return nil, errors.Wrap(err, "could not write runtime header")
	}

	runtime, err := runtimeFiles.Open(".runtime/runtime.cpp")
	if err != nil {
		return nil, errors.Wrap(err, "could not write runtime code")
	}

	return []*EmittedFile{
		{Filename: "runtime.hpp", Contents: header},
		{Filename: "runtime.cpp", Contents: runtime},
	}, nil
}
