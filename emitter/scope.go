package emitter

import (
	"fmt"

	"github.com/pkg/errors"
)

type Scope struct {
	Parent   *Scope
	Children []*Scope

	IdentTypes map[string]*TypeSpec
	Types      []*TypeSpec

	nextIdentNum int
}

func NewScope() *Scope {
	return &Scope{
		IdentTypes: make(map[string]*TypeSpec),
		Types:      make([]*TypeSpec, 0),
		Children:   make([]*Scope, 0),
	}
}

func (s *Scope) Clone() *Scope {
	newScope := NewScope()

	for k, v := range s.IdentTypes {
		newScope.IdentTypes[k] = v
	}

	newScope.Parent = s.Parent
	newScope.Types = append(newScope.Types, s.Types...)
	newScope.Children = s.Children

	return newScope
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

func (s *Scope) AddType(t *TypeSpec) {
	s.Types = append(s.Types, t)
}

func (s *Scope) addPrimitiveType(t primitiveType) {
	s.AddType(&TypeSpec{PrimitiveType: &t})
}

func (s *Scope) AddIdentifer(ident string, identType *TypeSpec) {
	s.IdentTypes[ident] = identType
}
