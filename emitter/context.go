package emitter

import (
	"bufio"
	"elauffenburger/hypescript/ast"
	"fmt"
	"strings"

	"github.com/pkg/errors"
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

	IdentTypes map[string]*TypeSpec
	Types      []*TypeSpec

	nextIdentNum int
}

func NewScope() *Scope {
	return &Scope{
		IdentTypes: make(map[string]*TypeSpec),
		Types:      make([]*TypeSpec, 0),
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

func (scope *Scope) TypeOf(ident string) (*TypeSpec, error) {
	t, ok := scope.IdentTypes[ident]
	if ok {
		return t, nil
	}

	if scope.Parent == nil {
		return nil, errors.WithStack(fmt.Errorf("unknown identifier %s in scope: %#v", ident, scope))
	}

	return scope.Parent.TypeOf(ident)
}

func (s *Scope) ContainsIdent(ident string) bool {
	_, ok := s.IdentTypes[ident]

	if !ok && s.Parent != nil {
		return s.Parent.ContainsIdent(ident)
	}

	return false
}

func (scope *Scope) NewIdent() string {
	ident := scope.nextIdentNum
	scope.nextIdentNum++

	return fmt.Sprintf("ident%d", ident)
}

func (s *Scope) GetNamedType(name string) (*TypeSpec, error) {
	for _, t := range s.Types {
		if t.InterfaceDefinition != nil && t.InterfaceDefinition.Name == name {
			return t, nil
		}

		if t.PrimitiveType != nil && string(*t.PrimitiveType) == name {
			return t, nil
		}
	}

	if s.Parent != nil {
		return s.Parent.GetNamedType(name)
	}

	return nil, fmt.Errorf("failed to find type %s in scope", name)
}

func (s *Scope) ContainsType(t *TypeSpec) bool {
	if t.InterfaceDefinition != nil {
		i, _ := s.GetNamedType(t.InterfaceDefinition.Name)

		return i != nil
	}

	if t.TypeReference != nil {
		i, _ := s.GetNamedType(*t.TypeReference)

		return i != nil
	}

	return false
}

func (s *Scope) ValidateHasType(t *TypeSpec) error {
	if !s.ContainsType(t) {
		return errors.WithStack(fmt.Errorf("unknown type %#v in current scope", t))
	}

	return nil
}

func (context *Context) TypeOf(ident string) (*TypeSpec, error) {
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

func (scope *Scope) AddIdentifer(ident string, identType *TypeSpec) {
	scope.IdentTypes[ident] = identType
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

func (s *Scope) AddType(t *TypeSpec) {
	s.Types = append(s.Types, t)
}

func (s *Scope) addPrimitiveType(t primitiveType) {
	s.AddType(&TypeSpec{PrimitiveType: &t})
}

func strRef(str string) *string {
	return &str
}
