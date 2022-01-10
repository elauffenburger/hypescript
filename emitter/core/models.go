package core

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
	Members map[string]*Member
}

func (o *Object) AllMembers() map[string]*Member {
	return o.Members
}

type ObjectTypeField struct {
	Name string
	Type *TypeSpec
}

type Interface struct {
	Name    string
	Members map[string]*Member
}

type Member struct {
	Name     string
	Field    *ObjectTypeField
	Function *Function
}

func (m *Member) Type() *TypeSpec {
	if m.Field != nil {
		return m.Field.Type
	} else if m.Function != nil {
		return &TypeSpec{Function: m.Function}
	}

	panic("unknown member type")
}

func (i *Interface) AllMembers() map[string]*Member {
	return i.Members
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
	AllMembers() map[string]*Member
}
