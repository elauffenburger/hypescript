package core

import "sync"

type PrimitiveType string

const (
	TsString PrimitiveType = "string"
	TsNumber PrimitiveType = "number"
	TsVoid   PrimitiveType = "void"
	TsAny    PrimitiveType = "any"
)

var PrimitiveTypes = []PrimitiveType{TsString, TsNumber, TsVoid}

type RuntimeType string

const (
	RtTsObject   RuntimeType = "TsObject"
	RtTsFunction RuntimeType = "TsFunction"
	RtTsVoid     RuntimeType = "void"
)

type TypeId int

const (
	TypeIdNone       TypeId = 0
	TypeIdTsObject   TypeId = 1
	TypeIdTsNum      TypeId = 2
	TypeIdTsString   TypeId = 3
	TypeIdTsFunction TypeId = 4
	TypeIdVoid       TypeId = 5
	TypeIdIntrinsic  TypeId = 6
)

type Function struct {
	Name               *string
	Parameters         []*FunctionParameter
	ImplicitReturnType *TypeSpec
	ExplicitReturnType *TypeSpec
	Body               []*StatementOrExpression
}

type StatementOrExpression struct {
	Statement  *Statement
	Expression *Expression

	Scope *Scope
}

type Expression struct {
	Number                 *Number
	String                 *string
	IdentAssignment        *IdentAssignment
	FunctionInstantiation  *Function
	ChainedObjectOperation *ChainedObjectOperation
	ObjectInstantiation    *ObjectInstantiation
	Ident                  *string
}

type Object struct {
	memberTracker *memberTracker
}

func NewObject(members []*Member) *Object {
	obj := &Object{}
	for _, m := range members {
		obj.AddMember(m)
	}

	return obj
}

func (o *Object) AddMember(m *Member) {
	if o.memberTracker == nil {
		o.memberTracker = newMemberTracker()
	}

	o.memberTracker.AddMember(m)
}

func (o *Object) AllMembers() map[string]*Member {
	if o.memberTracker == nil {
		o.memberTracker = newMemberTracker()
	}

	return o.memberTracker.AllMembers()
}

func (o *Object) MemberResolved(name string) <-chan *Member {
	if o.memberTracker == nil {
		o.memberTracker = newMemberTracker()
	}

	return o.memberTracker.MemberResolved(name)
}

type ObjectTypeField struct {
	Name string
	Type *TypeSpec
}

type Interface struct {
	memberTracker *memberTracker

	Name string
}

func NewInterface(name string, members []*Member) *Interface {
	i := &Interface{Name: name}
	for _, m := range members {
		i.AddMember(m)
	}

	return i
}

func (i *Interface) AddMember(m *Member) {
	if i.memberTracker == nil {
		i.memberTracker = newMemberTracker()
	}

	i.memberTracker.AddMember(m)
}

func (i *Interface) AllMembers() map[string]*Member {
	if i.memberTracker == nil {
		i.memberTracker = newMemberTracker()
	}

	return i.memberTracker.AllMembers()
}

func (i *Interface) MemberResolved(name string) <-chan *Member {
	if i.memberTracker == nil {
		i.memberTracker = newMemberTracker()
	}

	return i.memberTracker.MemberResolved(name)
}

type Member struct {
	Field *ObjectTypeField
}

func (m *Member) Name() *string {
	if m.Field != nil {
		return &m.Field.Name
	}

	panic("unknown member type")
}

func (m *Member) Type() *TypeSpec {
	if m.Field != nil {
		return m.Field.Type
	}

	panic("unknown member type")
}

type FunctionParameter struct {
	Name     string
	Optional bool
	Type     *TypeSpec
}

type Union struct {
	Types map[*TypeSpec]bool
}

type IdentAssignment struct {
	Ident      string
	Assignment Assignment
}

type ChainedObjectOperation struct {
	First *ObjectOperation
	Last  *ObjectOperation
}

type ObjectOperation struct {
	Accessee *Accessable

	Access     *ObjectAccess
	Invocation *ObjectInvocation
	Assignment *Assignment

	Next *ObjectOperation
	Prev *ObjectOperation
}

type ObjectInvocation struct {
	Accessee  *Accessable
	Arguments []*Expression
}

type ObjectAccess struct {
	AccessedIdent string
	Type          *TypeSpec
}

type Accessable struct {
	Ident *string
	Type  *TypeSpec
}

type Number struct {
	Integer *int
}

type LetDecl struct {
	Name         string
	ExplicitType *TypeSpec
	Value        *Expression
}

type Statement struct {
	FunctionInstantiation *Function
	ExpressionStmt        *Expression
	LetDecl               *LetDecl
	ReturnStmt            *Expression
}

type Assignment struct {
	Value *Expression
}

type ObjectInstantiation struct {
	Fields []*ObjectFieldInstantiation
}

type ObjectFieldInstantiation struct {
	Name  string
	Type  *TypeSpec
	Value *Expression
}

type TopLevelConstruct struct {
	InterfaceDefinition   *Interface
	StatementOrExpression *StatementOrExpression
}

type TS struct {
	TopLevelConstructs []TopLevelConstruct
}

/// ContainsAllTypeSpecs returns true if right contains all specs in left.
func ContainsAllTypeSpecs(left, right []*TypeSpec) bool {
	unseen := left[:]

	for _, t1 := range unseen {
		for _, t2 := range right {
			if t1.EqualsStrict(t2) {
				unseen = unseen[1:]
				break
			}
		}
	}

	return len(unseen) == 0
}

type HasMembers interface {
	AddMember(m *Member)
	AllMembers() map[string]*Member
	MemberResolved(name string) <-chan *Member
}

type memberTracker struct {
	lock      *sync.Mutex
	listeners map[string][]chan *Member

	members map[string]*Member
}

func newMemberTracker() *memberTracker {
	return &memberTracker{
		lock:      &sync.Mutex{},
		listeners: map[string][]chan *Member{},
		members:   map[string]*Member{},
	}
}

func (t *memberTracker) AddMember(m *Member) {
	t.members[*m.Name()] = m

	t.resolveMember(m)
}

func (t *memberTracker) AllMembers() map[string]*Member {
	return t.members
}

func (t *memberTracker) MemberResolved(name string) <-chan *Member {
	c := make(chan *Member)

	t.lock.Lock()

	if t.listeners[name] == nil {
		t.listeners[name] = []chan *Member{}
	}

	t.listeners[name] = append(t.listeners[name], c)

	t.lock.Unlock()

	if m := t.members[name]; m != nil {
		go func() {
			t.resolveMember(m)
		}()
	}

	return c
}

func (t *memberTracker) resolveMember(m *Member) {
	listeners := t.listeners[*m.Name()]
	if listeners == nil {
		return
	}

	for _, l := range listeners {
		l := l

		go func() {
			l <- m
			close(l)
		}()
	}
}
