package writepass

import (
	"bufio"
	"elauffenburger/hypescript/emitter/core"
	"strings"

	"github.com/pkg/errors"
)

type Context struct {
	scopeTracker core.ScopeTracker

	Output *bufio.Writer
}

func NewContext(w *bufio.Writer) *Context {
	ctx := Context{Output: w}

	return &ctx
}

func (ctx *Context) Run(global *core.Scope) error {
	ctx.scopeTracker = core.NewScopeTracker(global)

	// Write the preamble.
	ctx.WriteString(`
		#include <stdlib.h>
		#include <stdio.h>
		#include <string>
		#include <vector>
		#include <algorithm>
		#include <memory>
	
		#include "runtime.hpp"
	`)

	// Emit everything inside of main.
	ctx.WriteString("int main() { ")

	for _, stmtOrExpr := range ctx.currentScope().StatementsOrExpressions {
		if err := ctx.writeStatementOrExpression(stmtOrExpr); err != nil {
			return err
		}
	}

	ctx.WriteString(`return 0; }`)

	if err := ctx.Output.Flush(); err != nil {
		return errors.Wrap(err, "failed to write main.cpp")
	}

	return nil
}

func (ctx *Context) currentScope() *core.Scope {
	return ctx.scopeTracker.CurrentScope()
}

func (ctx *Context) EnterScope() *core.Scope {
	return ctx.scopeTracker.EnterScope()
}

func (ctx *Context) ExitScope() {
	ctx.scopeTracker.ExitScope()
}

func (ctx *Context) WithinScope(s *core.Scope, op func() error) error {
	return ctx.scopeTracker.WithinScope(s, op)
}

func (ctx *Context) WithinTempScope(op func() (interface{}, error)) (interface{}, error) {
	return ctx.scopeTracker.WithinTempScope(op)
}

func (ctx *Context) WithinNewScope(op func() error) error {
	return ctx.scopeTracker.WithinNewScope(op)
}

func (ctx *Context) WriteString(str string) {
	ctx.Output.WriteString(str)
}

// TODO: remove this; it's kind of a hack!
func (ctx *Context) WithinPrintContext(operation func(*Context) error) (string, error) {
	output := strings.Builder{}

	printCtx := &Context{
		scopeTracker: ctx.scopeTracker,
		Output:       bufio.NewWriter(&output),
	}

	err := operation(printCtx)
	printCtx.Output.Flush()

	return output.String(), err
}
