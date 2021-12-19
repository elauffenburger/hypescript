package ast

import "fmt"

type TypeDefinition struct {
	InterfaceDefinition *InterfaceDefinition `@@`
}

type InterfaceDefinition struct {
	Name    string                       `"interface" @Ident "{"`
	Members []*InterfaceMemberDefinition `(@@";")* "}"`
}

type InterfaceMemberDefinition struct {
	Field  *InterfaceFieldDefinition  `@@`
	Method *InterfaceMethodDefinition `| @@`
}

type InterfaceFieldDefinition struct {
	Name string `@Ident`
	Type Type   `":" @@`
}

type InterfaceMethodDefinition struct {
	Name       string              `@Ident`
	Parameters []FunctionParameter `"(" (@@ ("," @@)*)? ")"`
	ReturnType *Type               `(":" @@)?;`
}

type FunctionInstantiation struct {
	Name       *string                 `"function" @Ident?`
	Parameters []FunctionParameter     `"(" (@@ ("," @@)*)? ")"`
	ReturnType *Type                   `(":" @@)?`
	Body       []StatementOrExpression `"{"@@*"}"`
}

type Type struct {
	NonUnionType *NonUnionType `@@`
	UnionType    *UnionType    `| @@`
}

type NonUnionType struct {
	LiteralType   *LiteralType `@@`
	TypeReference *string      `| @Ident`
}

type LiteralType struct {
	FunctionType *FunctionType `@@`
	ObjectType   *ObjectType   `| @@`
}

type FunctionType struct {
	Parameters []FunctionParameter `"(" (@@ ("," @@)*)? ")"`
	ReturnType *Type               `"=>" @@`
}

type ObjectType struct {
	Fields []ObjectTypeField `"{" @@? ("," @@)* ","? "}"`
}

type ObjectTypeField struct {
	Name string `(@Ident | ("\""@Ident"\"")) ":"`
	Type Type   `":" @@`
}

type UnionType struct {
	Head *NonUnionType  `@@`
	Tail []NonUnionType `("|" @@)*`
}

type FunctionParameter struct {
	Name string `@Ident`
	Type Type   `":" @@`
}

type StatementOrExpression struct {
	Statement  *Statement  `@@`
	Expression *Expression `| @@`
}

type Expression struct {
	Number                 *Number                 `@@`
	String                 *string                 `| @String`
	IdentAssignment        *IdentAssignment        `| @@`
	FunctionInstantiation  *FunctionInstantiation  `| @@`
	ChainedObjectOperation *ChainedObjectOperation `| @@`
	ObjectInstantiation    *ObjectInstantiation    `| @@`
	Ident                  *string                 `| @Ident`
}

type IdentAssignment struct {
	Ident      string     `@Ident`
	Assignment Assignment `@@`
}

type ChainedObjectOperation struct {
	Accessee   Accessable        `@@`
	Operations []ObjectOperation `@@ @@*`
	Assignment *Assignment       `@@?`
}

type ObjectOperation struct {
	Access     *ObjectAccess     `@@`
	Invocation *ObjectInvocation `| @@`
}

type ObjectInvocation struct {
	Arguments []Expression `"("@@? ("," @@)* ")"`
}

type ObjectAccess struct {
	AccessedIdent string `"."@Ident`
}

type Accessable struct {
	Ident       *string      `@Ident`
	LiteralType *LiteralType `| "(" @@ ")"`
}

type Number struct {
	Integer *int `@Int`
}

type LetDecl struct {
	Name  string     `"let" @Ident`
	Value Expression `"=" @@`
}

type Statement struct {
	FunctionInstantiation *FunctionInstantiation `@@ ";"?`
	ExpressionStmt        *Expression            `| @@ ";"`
	LetDecl               *LetDecl               `| @@ ";"`
	ReturnStmt            *Expression            `| "return" @@ ";"`
}

type Assignment struct {
	Value Expression `"=" @@`
}

type ObjectInstantiation struct {
	Fields []ObjectFieldInstantiation `"{" @@? ("," @@)* ","? "}"`
}

type ObjectFieldInstantiation struct {
	Name  string     `(@Ident | ("\""@Ident"\"")) ":"`
	Value Expression `@@`
}

type TopLevelConstruct struct {
	StatementOrExpression *StatementOrExpression `@@`
}

type TS struct {
	TopLevelConstructs []TopLevelConstruct `@@*`
}

func (left *Type) Equals(right *Type) bool {
	if right == nil {
		return false
	}

	if left.NonUnionType != nil && right.NonUnionType != nil {
		left, right := left.NonUnionType, right.NonUnionType

		if left.TypeReference != nil && right.TypeReference != nil {
			return *left.TypeReference == *right.TypeReference
		}
	}

	panic(fmt.Sprintf("Unsupported type comparison between %#v and %#v", left, right))
}

func CreateUnionType(left, right *Type) *Type {
	if left.NonUnionType != nil && right.NonUnionType != nil {
		left, right := left.NonUnionType, right.NonUnionType

		return &Type{UnionType: &UnionType{Head: left, Tail: []NonUnionType{*right}}}
	}

	panic(fmt.Errorf("Union type creation for %#v and %#v not implemented!", left, right))
}
