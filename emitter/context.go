package emitter

import (
	"bufio"
	"elauffenburger/hypescript/ast"
	"fmt"
)

type Context struct {
	scopes       []Scope
	CurrentScope *Scope

	Output *bufio.Writer
}

func (context *Context) WriteString(str string) {
	context.Output.WriteString(str)
}

type Scope struct {
	IdentTypes map[string]ast.Type
}

func (scope *Scope) TypeOf(ident string) *ast.Type {
	t, ok := scope.IdentTypes[ident]
	if !ok {
		panic(fmt.Sprintf("Unknown identifier %s in scope: %#v", ident, scope))
	}

	return &t
}

func (context *Context) TypeOf(ident string) *ast.Type {
	return context.CurrentScope.TypeOf(ident)
}

func (context *Context) EnterScope() {
	newScope := Scope{
		IdentTypes: make(map[string]ast.Type),
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

func (scope *Scope) AddIdentifer(ident string, identType ast.Type) {
	scope.IdentTypes[ident] = identType
}
