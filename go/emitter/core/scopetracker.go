package core

type scopeTracker struct {
	currentScope *Scope
}

type ScopeTracker interface {
	CurrentScope() *Scope
	EnterScope() *Scope
	ExitScope()
	WithinScope(s *Scope, op func() error) error
	WithinTempScope(op func() (interface{}, error)) (interface{}, error)
	WithinNewScope(op func() error) error
	Clone() ScopeTracker
}

func NewScopeTracker(scope *Scope) ScopeTracker {
	return &scopeTracker{currentScope: scope}
}

func (s *scopeTracker) CurrentScope() *Scope {
	return s.currentScope
}

func (s *scopeTracker) EnterScope() *Scope {
	var newScope *Scope
	if s.currentScope != nil {
		newScope = s.currentScope.Clone()

		newScope.Parent = s.currentScope
		s.currentScope.Children = append(s.currentScope.Children, newScope)
	} else {
		newScope = NewScope()
	}

	s.currentScope = newScope

	return newScope
}

func (s *scopeTracker) ExitScope() {
	s.currentScope = s.currentScope.Parent
}

func (s *scopeTracker) WithinScope(scope *Scope, op func() error) error {
	oldScope := s.currentScope

	s.currentScope = scope
	err := op()
	s.currentScope = oldScope

	return err
}

func (s *scopeTracker) WithinTempScope(op func() (interface{}, error)) (interface{}, error) {
	var scope *Scope
	if s.currentScope != nil {
		scope = s.currentScope.Clone()
		scope.Parent = s.currentScope
	} else {
		scope = NewScope()
	}

	s.currentScope = scope

	result, err := op()

	s.currentScope = scope.Parent

	return result, err
}

func (s *scopeTracker) WithinNewScope(op func() error) error {
	s.EnterScope()

	err := op()

	s.ExitScope()

	return err
}

func (s *scopeTracker) Clone() ScopeTracker {
	return &scopeTracker{
		currentScope: s.currentScope.Clone(),
	}
}
