package emitter

import (
	"bufio"
	"strings"
)

type Context struct {
	scopes       []*Scope
	CurrentScope *Scope

	Output *bufio.Writer
}

func (ctx *Context) WriteString(str string) {
	ctx.Output.WriteString(str)
}

// TODO: remove this; it's kind of a hack!
func (ctx *Context) WithinPrintContext(operation func(*Context) error) (string, error) {
	output := strings.Builder{}

	printCtx := &Context{
		scopes:       ctx.scopes,
		CurrentScope: ctx.CurrentScope,
		Output:       bufio.NewWriter(&output),
	}

	err := operation(printCtx)
	printCtx.Output.Flush()

	return output.String(), err
}

func (ctx *Context) WithinScope(s *Scope, op func() error) error {
	oldScope := ctx.CurrentScope

	ctx.CurrentScope = s
	err := op()
	ctx.CurrentScope = oldScope

	return err
}

func (ctx *Context) WithinTempScope(operation func() (interface{}, error)) (interface{}, error) {
	var scope *Scope
	if ctx.CurrentScope != nil {
		scope = ctx.CurrentScope.Clone()
		scope.Parent = ctx.CurrentScope
	} else {
		scope = NewScope()
	}

	ctx.CurrentScope = scope

	result, err := operation()

	ctx.CurrentScope = scope.Parent

	return result, err
}

func (ctx *Context) WithinNewScope(operation func() error) error {
	ctx.EnterScope()

	err := operation()

	ctx.ExitScope()

	return err
}

func (ctx *Context) EnterScope() *Scope {
	var newScope *Scope
	if ctx.CurrentScope != nil {
		newScope = ctx.CurrentScope.Clone()

		newScope.Parent = ctx.CurrentScope
		ctx.CurrentScope.Children = append(ctx.CurrentScope.Children, newScope)
	} else {
		newScope = NewScope()
	}

	ctx.scopes = append(ctx.scopes, newScope)
	ctx.CurrentScope = newScope

	return newScope
}

func (ctx *Context) ExitScope() {
	ctx.CurrentScope = ctx.CurrentScope.Parent
}

func NewContext(w *bufio.Writer) *Context {
	ctx := Context{
		Output: w,
	}

	global := ctx.EnterScope()

	global.AddType(&TypeSpec{
		Interface: &Interface{
			Name: "Console",
			Members: []*InterfaceMember{
				{
					Method: &InterfaceMethod{
						Name: "log",
						Parameters: []*FunctionParameter{
							{
								Name: "fmt",
								Type: &TypeSpec{
									TypeReference: strRef("any"),
								},
							},
						},
					},
				},
			},
		},
	})

	global.AddIdentifer("console", &TypeSpec{TypeReference: strRef("Console")})

	for _, t := range primitiveTypes {
		global.addPrimitiveType(t)
	}

	return &ctx
}
