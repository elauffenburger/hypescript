package main

import (
	"fmt"

	"github.com/alecthomas/participle"
)

type Function struct {
	Name       string                   `"function" @Ident`
	Arguments  []*FunctionArgument      `"(" (@@ ("," @@)*)? ")"`
	ReturnType string                   `(":" @Ident)?`
	Body       []*ExpressionOrStatement `"{"@@*"}"`
}

type FunctionArgument struct {
	Name string `@Ident`
	Type string `":" @Ident`
}

type ExpressionOrStatement struct {
	Statement  *Statement  `@@`
	Expression *Expression `| @@`
}

type Expression struct {
	WrappedExpression *Expression `"("@@")"`
	Number            *Number     `| @@`
	String            string      `| @String`
	Ident             string      `| @Ident`
}

type Number struct {
	Integer int `@Int`
}

type LetDecl struct {
	Name  string      `"let" @Ident`
	Value *Expression `"=" @@ ";"`
}

type Statement struct {
	LetDecl *LetDecl    `@@`
	Return  *Expression `| "return" @@ ";"`
}

type TS struct {
	Functions []*Function `@@*`
}

func main() {
	parser, err := participle.Build(&TS{})
	if err != nil {
		panic(fmt.Errorf("building parser failed: %w", err))
	}

	ast := &TS{}
	err = parser.ParseString(`
		function foo (bar: string, baz: num): num {
			let foo = 5;
			let bar = "bar";

			return foo;
		}
	`, ast)

	if err != nil {
		panic(fmt.Errorf("parsing failed: %w", err))
	}
}
