package core

import (
	"reflect"
)

type PrimitiveType string

const (
	TsString PrimitiveType = "string"
	TsNumber PrimitiveType = "number"
	TsVoid   PrimitiveType = "void"
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

type TypeSpec struct {
	Function      *Function
	Object        *Object
	Interface     *Interface
	TypeReference *string
	Union         *Union

	unresolved bool
	resolver   func()

	// Redirect is a redirect to another TypeSpec
	Redirect *TypeSpec
}

func (t *TypeSpec) Unresolved() bool {
	return t.unresolved
}

func (t *TypeSpec) MarkResolved() {
	t.resolver()
}

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
	Fields []*ObjectTypeField
}

type ObjectTypeField struct {
	Name string
	Type *TypeSpec
}

type Interface struct {
	Name    string
	Members []*InterfaceMember
}

type InterfaceMember struct {
	Field  *InterfaceField
	Method *InterfaceMethod
}

type InterfaceField struct {
	Name string
	Type *TypeSpec
}

type InterfaceMethod struct {
	Name       string
	Parameters []*FunctionParameter
	ReturnType *TypeSpec
}

type FunctionParameter struct {
	Name string
	Type *TypeSpec
}

type Union struct {
	Head *TypeSpec
	Tail []*TypeSpec
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
}

type Accessable struct {
	Ident *string
	Type  *TypeSpec
}

type Number struct {
	Integer *int
}

type LetDecl struct {
	Name  string
	Value *Expression
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

/// EqualsReferencing returns true if t is a reference to other, or other is a reference
/// to t.
func (t *TypeSpec) EqualsReferencing(other *TypeSpec) bool {
	return t.referenceTo(other) || other.referenceTo(t)
}

func (t *TypeSpec) referenceTo(other *TypeSpec) bool {
	if t.TypeReference != nil {
		if other.TypeReference == t.TypeReference {
			return true
		}

		if other.Interface != nil && other.Interface.Name == *t.TypeReference {
			return true
		}
	}

	return false
}

/// EqualsStrict returns true if t is deeply equal to other.
///
/// If you want to test if the TypeSpecs are loosely equal via references,
/// use EqualsReferencing.
func (t *TypeSpec) EqualsStrict(other *TypeSpec) bool {
	return reflect.DeepEqual(t, other)
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
