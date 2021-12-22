package emitter

import (
	"bufio"
	"elauffenburger/hypescript/ast"
	"fmt"
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

type Scope struct {
	Parent *Scope

	IdentTypes map[string]*TypeDefinition
	Types      []*TypeDefinition

	nextIdentNum int
}

func NewScope() *Scope {
	return &Scope{
		IdentTypes: make(map[string]*TypeDefinition),
		Types:      make([]*TypeDefinition, 0),
	}
}

func (scope *Scope) Clone() *Scope {
	newScope := NewScope()

	for k, v := range scope.IdentTypes {
		newScope.IdentTypes[k] = v
	}

	newScope.Types = append(newScope.Types, scope.Types...)

	return newScope
}

func (scope *Scope) TypeOf(ident string) (*TypeDefinition, error) {
	t, ok := scope.IdentTypes[ident]
	if ok {
		return t, nil
	}

	if scope.Parent == nil {
		return nil, fmt.Errorf("unknown identifier %s in scope: %#v", ident, scope)
	}

	return scope.Parent.TypeOf(ident)
}

func (scope *Scope) NewIdent() string {
	ident := scope.nextIdentNum
	scope.nextIdentNum++

	return fmt.Sprintf("ident%d", ident)
}

func (context *Context) TypeOf(ident string) (*TypeDefinition, error) {
	return context.CurrentScope.TypeOf(ident)
}

func (context *Context) EnterScope() *Scope {
	var newScope *Scope
	if context.CurrentScope != nil {
		newScope = context.CurrentScope.Clone()

		newScope.Parent = context.CurrentScope
	} else {
		newScope = NewScope()
	}

	context.scopes = append(context.scopes, newScope)
	context.CurrentScope = newScope

	return newScope
}

func (context *Context) ExitScope() {
	context.scopes = context.scopes[:len(context.scopes)-1]

	if len(context.scopes) == 0 {
		context.CurrentScope = nil
		return
	}

	context.CurrentScope = context.scopes[len(context.scopes)-1]
}

func (scope *Scope) AddIdentifer(ident string, identType *TypeDefinition) {
	scope.IdentTypes[ident] = identType
}

func NewContext(output *bufio.Writer) *Context {
	ctx := Context{
		Output: output,
	}

	global := ctx.EnterScope()

	global.AddIdentifer("Console", &TypeDefinition{
		ObjectType: &ast.ObjectType{
			Fields: []ast.ObjectTypeField{
				{
					Name: "log",
					Type: ast.TypeIdentifier{
						NonUnionType: &ast.NonUnionType{
							LiteralType: &ast.LiteralType{
								FunctionType: &ast.FunctionType{
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
				},
			},
		},
	})

	global.AddIdentifer("console", &TypeDefinition{TypeReference: strRef("Console")})

	return &ctx
}

func strRef(str string) *string {
	return &str
}
