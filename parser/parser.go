package parser

import (
	"elauffenburger/hypescript/ast"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/pkg/errors"
)

type Parser interface {
	ParseString(str string) (*ast.TS, error)
}

type parser struct {
}

func (p parser) ParseString(str string) (*ast.TS, error) {
	lex := lexer.MustSimple([]lexer.Rule{
		{"Int", `\d+`, nil},
		{"Ident", `[a-zA-Z_$][a-zA-Z_$0-9]*`, nil},
		{"String", `"[^"]*"`, nil},
		{"Whitespace", `(?:[\s\t]|\n|(?:\r\n))+`, nil},
		{"Punct", `[,.<>(){}=:;]`, nil},
		{"Comment", `//.*`, nil},
		{"Reserved", `(let|function)`, nil},
	})

	parser := participle.MustBuild(
		&ast.TS{},
		participle.Lexer(lex),
		participle.Elide("Whitespace", "Comment"),
	)

	ast := &ast.TS{}
	err := parser.ParseString("temp.ts", str, ast)

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse program")
	}

	return ast, nil
}

func New() Parser {
	return &parser{}
}
