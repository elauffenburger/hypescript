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
	Primitive     *PrimitiveType
	Union         *Union

	Unresolved bool
	Source     interface{}
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

func (t *TypeSpec) Equals(other *TypeSpec) bool {
	return reflect.DeepEqual(t, other)
}
