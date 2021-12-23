package core

import (
	"fmt"

	"github.com/pkg/errors"
)

type Scope struct {
	Parent   *Scope
	Children []*Scope

	IdentTypes      map[string]*TypeSpec
	Types           []*TypeSpec
	UnresolvedTypes []*TypeSpec

	StatementsOrExpressions []*StatementOrExpression

	nextIdentNum int
}

func NewScope() *Scope {
	return &Scope{
		IdentTypes:              make(map[string]*TypeSpec),
		Types:                   make([]*TypeSpec, 0),
		Children:                make([]*Scope, 0),
		StatementsOrExpressions: make([]*StatementOrExpression, 0),
		UnresolvedTypes:         make([]*TypeSpec, 0),
	}
}

func NewGlobalScope() *Scope {
	scope := NewScope()

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

	for _, t := range PrimitiveTypes {
		scope.AddPrimitiveType(t)
	}

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

func (s *Scope) RegisteredTypeOf(ident string) *TypeSpec {
	// Try to get the type.
	t, err := s.TypeOf(ident)

	// If it is registered, return it!
	if t != nil && err == nil {
		return t
	}

	// ...otherwise, create and register a placeholder we'll
	// fill out later (hopefully).

	t = s.NewPlaceholderTypeExpression(&Expression{Ident: &ident})
	s.UnresolvedTypes = append(s.UnresolvedTypes, t)
	s.IdentTypes[ident] = t

	return t
}

func (s *Scope) RegisteredType(t *TypeSpec) *TypeSpec {
	// Try to get the type.
	for _, registered := range s.Types {
		// If we found the referenced type, return it!
		if t.TypeReference != nil && registered.TypeReference == t.TypeReference {
			return registered
		}

		// If it's registered, return it!
		if registered.Equals(t) {
			return registered
		}
	}

	// ...otherwise, add the type and hopefully resolve it later.
	s.UnresolvedTypes = append(s.UnresolvedTypes, t)

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

		if t.Primitive != nil && string(*t.Primitive) == name {
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
	t := &TypeSpec{Unresolved: true, Source: source}

	return t
}

func (s *Scope) AddPrimitiveType(t PrimitiveType) {
	s.AddType(&TypeSpec{Primitive: &t})
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
		fields := make([]*ObjectTypeField, len(objInst.Fields))
		for i, f := range objInst.Fields {
			fields[i] = &ObjectTypeField{
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

func strRef(str string) *string {
	return &str
}
