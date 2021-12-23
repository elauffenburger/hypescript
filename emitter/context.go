package emitter

import (
	"bufio"
	"elauffenburger/hypescript/ast"
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

func (ctx *Context) WithinNewScope(operation func() error) error {
	ctx.EnterScope()

	err := operation()

	ctx.ExitScope()

	return err
}

func (ctx *Context) TypeOf(ident string) (*TypeSpec, error) {
	return ctx.CurrentScope.TypeOf(ident)
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

func NewContext(output *bufio.Writer) *Context {
	ctx := Context{
		Output: output,
	}

	global := ctx.EnterScope()

	global.AddType(&TypeSpec{
		InterfaceDefinition: &ast.InterfaceDefinition{
			Name: "Console",
			Members: []*ast.InterfaceMemberDefinition{
				{
					Method: &ast.InterfaceMethodDefinition{
						Name: "log",
						Parameters: []ast.FunctionParameter{
							{
								Name: "fmt",
								Type: ast.TypeIdentifier{
									NonUnionType: &ast.NonUnionType{
										TypeReference: strRef("any"),
									},
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
