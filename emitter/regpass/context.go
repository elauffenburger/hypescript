package regpass

import (
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/core"
	"fmt"
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
	// Register all types.
	if err := ctx.registerTypes(ast); err != nil {
		return err
	}

	for _, c := range ast.TopLevelConstructs {
		if c.StatementOrExpression != nil {
			_, err := ctx.registerStatementOrExpression(c.StatementOrExpression)
			if err != nil {
				return err
			}

			continue
		}
	}

	// Validate functions now that we've completely resolved all idents and types.
	for _, fn := range allFunctions(ctx.GlobalScope) {
		if err := validateFn(ctx.GlobalScope, fn); err != nil {
			return err
		}
	}

	// Make sure invocations make sense.
	for scope, invocs := range ctx.Invocations() {
		for _, invoc := range invocs {
			t := invoc.Accessee.Type

			switch {
			case t.Function != nil:
				fn := t.Function
				var fnName string
				if fn.Name != nil {
					fnName = *fn.Name
				} else {
					fnName = "<anonymous>"
				}

				if len(invoc.Arguments) > len(fn.Parameters) {
					return fmt.Errorf("too many args for invocation of fn %s", fnName)
				}

				for i, param := range fn.Parameters {
					if i > len(invoc.Arguments)-1 {
						if param.Optional {
							continue
						}

						return fmt.Errorf("missing arg for %s in call to fn %s", param.Name, fnName)
					}

					arg := invoc.Arguments[i]
					argType, err := scope.ExprType(arg)
					if err != nil {
						return err
					}

					if !argType.Satisfies(param.Type) {
						return fmt.Errorf("cannot use type %s for arg %s in call to %s", argType, param.Name, fnName)
					}
				}
			}

		}
	}

	return nil
}

func (ctx *Context) UnresolvedTypes() map[string]*core.TypeSpec {
	types := make(map[string]*core.TypeSpec, 0)
	addUnresolvedTypesFromScope(types, ctx.GlobalScope)

	return types
}

func addUnresolvedTypesFromScope(types map[string]*core.TypeSpec, s *core.Scope) {
	for k, v := range s.UnresolvedTypes() {
		types[k] = v
	}

	for _, child := range s.Children {
		addUnresolvedTypesFromScope(types, child)
	}
}

func (ctx *Context) registerTypes(ast *ast.TS) error {
	for _, c := range ast.TopLevelConstructs {
		if intdef := c.InterfaceDefinition; intdef != nil {
			if err := ctx.registerInterface(intdef); err != nil {
				return err
			}

			continue
		}
	}

	// Make sure that we can resolve any unresolved types we had pending.
	for name, t := range ctx.UnresolvedTypes() {
		regd, err := ctx.currentScope().ResolvedTypeFromName(name)
		if err != nil {
			return err
		}

		t.Redirect = regd
		t.MarkResolved()
	}

	// If there are still unresolved types, bail out.
	unresolved := ctx.UnresolvedTypes()
	if len(unresolved) != 0 {
		return fmt.Errorf("failed to resolve types: %s", unresolved)
	}

	return nil
}

func (c *Context) Invocations() map[*core.Scope][]*core.ObjectInvocation {
	invocs := make(map[*core.Scope][]*core.ObjectInvocation)
	addInvocationsFromScope(c.GlobalScope, invocs)

	return invocs
}

func addInvocationsFromScope(scope *core.Scope, invocs map[*core.Scope][]*core.ObjectInvocation) {
	if scope == nil {
		return
	}

	scopeInvocs := make([]*core.ObjectInvocation, 0)
	for _, stmtOrExpr := range scope.StatementsOrExpressions {
		// Try to grab the expression directly or fallback to checking for an expression statement.
		expr := stmtOrExpr.Expression
		if expr == nil {
			expr = stmtOrExpr.Statement.ExpressionStmt
		}

		if expr == nil || expr.ChainedObjectOperation == nil {
			continue
		}

		maybeInvoc := expr.ChainedObjectOperation.Last
		if maybeInvoc.Invocation == nil {
			continue
		}

		scopeInvocs = append(scopeInvocs, maybeInvoc.Invocation)
	}

	invocs[scope] = scopeInvocs

	for _, child := range scope.Children {
		addInvocationsFromScope(child, invocs)
	}
}
