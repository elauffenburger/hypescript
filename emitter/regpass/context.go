package regpass

import (
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/core"

	"github.com/pkg/errors"
)

type Context struct {
	scopeTracker core.ScopeTracker
	GlobalScope  *core.Scope
}

func NewContext() *Context {
	global := core.NewGlobalScope()

	return &Context{
		scopeTracker: core.NewScopeTracker(global),
		GlobalScope:  global,
	}
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

func (ctx *Context) Run(ast *ast.TS) error {
	// Register all constructs.
	for _, c := range ast.TopLevelConstructs {
		if c.StatementOrExpression != nil {
			_, err := ctx.registerStatementOrExpression(c.StatementOrExpression)
			if err != nil {
				return err
			}

			continue
		}

		if intdef := c.InterfaceDefinition; intdef != nil {
			ctx.registerInterface(intdef)

			continue
		}

		return errors.Errorf("unknown top-level construct: %v", c)
	}

	return nil
}

func strRef(str string) *string {
	return &str
}
