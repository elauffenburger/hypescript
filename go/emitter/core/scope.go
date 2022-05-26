package core

import (
	"elauffenburger/hypescript/typeutils"
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

	scope.AddType(&TypeSpec{Interface: NewInterface("string", []*Member{})})
	scope.AddType(&TypeSpec{Interface: NewInterface("number", []*Member{})})

	// TODO: this doesn't feel right; should be special type names maybe?
	scope.AddType(&TypeSpec{Interface: NewInterface("void", []*Member{})})
	scope.AddType(&TypeSpec{Interface: NewInterface("null", []*Member{})})
	scope.AddType(&TypeSpec{Interface: NewInterface("undefined", []*Member{})})

	scope.AddType(&TypeSpec{
		Interface: NewInterface(
			"Console",
			[]*Member{
				{
					Field: &ObjectTypeField{
						Name: "log",
						Type: &TypeSpec{
							Function: &Function{
								Name: typeutils.StrRef("log"),
								Parameters: []*FunctionParameter{
									{
										Name: "fmt",
										Type: &TypeSpec{
											TypeReference: typeutils.StrRef("any"),
										},
									},
								},
							},
						},
					},
				},
			},
		),
	})

	scope.AddIdentifer("console", &TypeSpec{TypeReference: typeutils.StrRef("Console")})

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

func (s *Scope) TypeFromName(typeName string) *TypeSpec {
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

func (s *Scope) ResolveType(t *TypeSpec) (*TypeSpec, error) {
	if t.TypeReference != nil {
		return s.ResolvedTypeFromName(*t.TypeReference)
	}

	return t, nil
}

func (s *Scope) IdentType(ident string) (*TypeSpec, error) {
	t, ok := s.IdentTypes[ident]
	if ok {
		return t, nil
	}

	if s.Parent == nil {
		return nil, errors.WithStack(fmt.Errorf("unknown identifier %s in scope: %#v", ident, s))
	}

	return s.Parent.IdentType(ident)
}

func (s *Scope) ResolveIdentType(ident string) (*TypeSpec, error) {
	t, err := s.IdentType(ident)
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

func (s *Scope) ResolvedTypeFromName(name string) (*TypeSpec, error) {
	for _, t := range s.Types {
		if t.Interface != nil && t.Interface.Name == name {
			return t, nil
		}
	}

	if s.Parent != nil {
		return s.Parent.ResolvedTypeFromName(name)
	}

	return nil, fmt.Errorf("failed to find type %s in scope", name)
}

func (s *Scope) ContainsType(t *TypeSpec) bool {
	if t.Interface != nil {
		i, _ := s.ResolvedTypeFromName(t.Interface.Name)

		return i != nil
	}

	if t.TypeReference != nil {
		i, _ := s.ResolvedTypeFromName(*t.TypeReference)

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

func (s *Scope) ExprType(expr *Expression) (*TypeSpec, error) {
	if expr.String != nil {
		t := string(TsString)

		return &TypeSpec{TypeReference: &t}, nil
	}

	if expr.Number != nil {
		t := string(TsNumber)

		return &TypeSpec{TypeReference: &t}, nil
	}

	if expr.Ident != nil {
		return s.IdentType(*expr.Ident)
	}

	if fn := expr.FunctionInstantiation; fn != nil {
		return &TypeSpec{Function: fn}, nil
	}

	if objInst := expr.ObjectInstantiation; objInst != nil {
		members := make([]*Member, len(objInst.Fields))
		for i, f := range objInst.Fields {
			members[i] = &Member{
				Field: &ObjectTypeField{
					Name: f.Name,
					Type: f.Type,
				},
			}
		}

		return &TypeSpec{Object: NewObject(members)}, nil
	}

	if chain := expr.ChainedObjectOperation; chain != nil {
		op := chain.Last

		if op.Access != nil {
			return op.Access.Type, nil
		}

		if op.Invocation != nil {
			return op.Invocation.Accessee.Type, nil
		}
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
