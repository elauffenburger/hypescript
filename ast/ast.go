package ast

import "fmt"

type Function struct {
	Name       string                  `"function" @Ident`
	Arguments  []FunctionArgument      `"(" (@@ ("," @@)*)? ")"`
	ReturnType *Type                   `(":" @@)?`
	Body       []StatementOrExpression `"{"@@*"}"`
}

type Type struct {
	NonUnionType *NonUnionType `@@`
	UnionType    *UnionType    `| @@`
}

type NonUnionType struct {
	TypeReference *string `@Ident`
}

type UnionType struct {
	Head *NonUnionType  `@@`
	Tail []NonUnionType `("|" @@)*`
}

type FunctionArgument struct {
	Name string `@Ident`
	Type Type   `":" @@`
}

type StatementOrExpression struct {
	Statement  *Statement  `@@`
	Expression *Expression `| @@`
}

type Expression struct {
	WrappedExpression *Expression `"("@@")"`
	Number            *Number     `| @@`
	String            *string     `| @String`
	Ident             *string     `| @Ident`
}

type Number struct {
	Integer *int `@Int`
}

type LetDecl struct {
	Name  string     `"let" @Ident`
	Value Expression `"=" @@ ";"`
}

type Statement struct {
	LetDecl    *LetDecl    `@@`
	ReturnStmt *Expression `| "return" @@ ";"`
}

type TS struct {
	Functions []Function `@@*`
}

func (left *Type) Equals(right *Type) bool {
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
