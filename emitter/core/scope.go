package core

import (
	"fmt"

	"github.com/pkg/errors"
)

type Scope struct {
	Parent   *Scope
	Children []*Scope

	IdentTypes map[string]*TypeSpec
	Types      []*TypeSpec

	// unresolvedTypes is a map from a type reference to a
	// placeholder type spec for that type.
	unresolvedTypes map[string]*TypeSpec

	StatementsOrExpressions []*StatementOrExpression

	nextIdentNum int
}

func NewScope() *Scope {
	return &Scope{
		IdentTypes:              make(map[string]*TypeSpec),
		Types:                   make([]*TypeSpec, 0),
		unresolvedTypes:         make(map[string]*TypeSpec),
		Children:                make([]*Scope, 0),
		StatementsOrExpressions: make([]*StatementOrExpression, 0),
	}
}

func NewGlobalScope() *Scope {
	scope := NewScope()

	scope.AddType(&TypeSpec{Interface: &Interface{Name: "string"}})
	scope.AddType(&TypeSpec{Interface: &Interface{Name: "number"}})
	// TODO: this doesn't feel right; should be a special type name.
	scope.AddType(&TypeSpec{Interface: &Interface{Name: "void"}})

	scope.AddType(&TypeSpec{
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

	scope.AddIdentifer("console", &TypeSpec{TypeReference: strRef("Console")})

	return scope
}

func (s *Scope) Clone() *Scope {
	newScope := NewScope()

	for k, v := range s.IdentTypes {
		newScope.IdentTypes[k] = v
	}

	newScope.Parent = s.Parent
	newScope.Types = append(newScope.Types, s.Types...)

	return newScope
}

func (s *Scope) AddStmt(stmt *Statement) *StatementOrExpression {
	return s.AddStmtOrExpr(&StatementOrExpression{Statement: stmt})
}

func (s *Scope) AddStmtOrExpr(stmtOrExpr *StatementOrExpression) *StatementOrExpression {
	if stmtOrExpr.Scope == nil {
		stmtOrExpr.Scope = s
	}

	s.StatementsOrExpressions = append(s.StatementsOrExpressions, stmtOrExpr)

	return stmtOrExpr
}

func (s *Scope) AddType(t *TypeSpec) {
	s.Types = append(s.Types, t)
}

func (s *Scope) AddIdentifer(ident string, identType *TypeSpec) {
	s.IdentTypes[ident] = identType
}

func (s *Scope) RegisteredType(typeName string) *TypeSpec {
	// Try to get the type.
	for _, registered := range s.Types {
		// Check if we can resolve the type.
		if i := registered.Interface; i != nil && i.Name == typeName {
			return registered
		}
	}

	// ...otherwise, add a placholder type and hopefully resolve it later.
	t := &TypeSpec{unresolved: true}

	// Attach a function that can be invoked to mark the type as resolved.
	t.resolver = func() {
		t.unresolved = false
		delete(s.unresolvedTypes, typeName)
	}

	s.unresolvedTypes[typeName] = t

	return t
}

func (s *Scope) TypeOf(ident string) (*TypeSpec, error) {
	t, ok := s.IdentTypes[ident]
	if ok {
		return t, nil
	}

	if s.Parent == nil {
		return nil, errors.WithStack(fmt.Errorf("unknown identifier %s in scope: %#v", ident, s))
	}

	return s.Parent.TypeOf(ident)
}

func (s *Scope) ResolveType(t *TypeSpec) (*TypeSpec, error) {
	if t.TypeReference != nil {
		return s.GetNamedType(*t.TypeReference)
	}

	return t, nil
}

func (s *Scope) GetResolvedType(ident string) (*TypeSpec, error) {
	t, err := s.TypeOf(ident)
	if err != nil {
		return nil, err
	}

	return s.ResolveType(t)
}

func (s *Scope) ContainsIdent(ident string) bool {
	_, ok := s.IdentTypes[ident]

	if !ok && s.Parent != nil {
		return s.Parent.ContainsIdent(ident)
	}

	return false
}

func (s *Scope) NewIdent() string {
	ident := s.nextIdentNum
	s.nextIdentNum++

	return fmt.Sprintf("ident%d", ident)
}

func (s *Scope) GetNamedType(name string) (*TypeSpec, error) {
	for _, t := range s.Types {
		if t.Interface != nil && t.Interface.Name == name {
			return t, nil
		}
	}

	if s.Parent != nil {
		return s.Parent.GetNamedType(name)
	}

	return nil, fmt.Errorf("failed to find type %s in scope", name)
}

func (s *Scope) ContainsType(t *TypeSpec) bool {
	if t.Interface != nil {
		i, _ := s.GetNamedType(t.Interface.Name)

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

func (s *Scope) NewPlaceholderTypeExpression(source interface{}) *TypeSpec {
	return &TypeSpec{unresolved: true}
}

func (s *Scope) GetTypeIdFor(t *TypeSpec) (int, error) {
	// TODO; need to actually make this work!

	return 0, nil
}

func (s *Scope) InferType(expr *Expression) (*TypeSpec, error) {
	if expr.String != nil {
		t := string(TsString)

		return &TypeSpec{TypeReference: &t}, nil
	}

	if expr.Number != nil {
		t := string(TsNumber)

		return &TypeSpec{TypeReference: &t}, nil
	}

	if expr.Ident != nil {
		return s.TypeOf(*expr.Ident)
	}

	if fn := expr.FunctionInstantiation; fn != nil {
		return &TypeSpec{Function: fn}, nil
	}

	if objInst := expr.ObjectInstantiation; objInst != nil {
		fields := make(map[string]*ObjectTypeField, len(objInst.Fields))
		for _, f := range objInst.Fields {
			fields[f.Name] = &ObjectTypeField{
				Name: f.Name,
				Type: f.Type,
			}
		}

		return &TypeSpec{Object: &Object{Fields: fields}}, nil
	}

	if chain := expr.ChainedObjectOperation; chain != nil {
		return chain.Last.Accessee.Type, nil
	}

	return nil, fmt.Errorf("unable to infer type")
}

func (s *Scope) UnresolvedTypes() map[string]*TypeSpec {
	types := make(map[string]*TypeSpec, len(s.unresolvedTypes))
	for k, v := range s.unresolvedTypes {
		types[k] = v
	}

	return types
}

func strRef(str string) *string {
	return &str
}
