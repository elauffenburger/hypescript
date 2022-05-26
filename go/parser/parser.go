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

type parser struct{}

func (p parser) ParseString(str string) (*ast.TS, error) {
	lex := lexer.MustSimple([]lexer.Rule{
		{Name: "Int", Pattern: `\d+`, Action: nil},
		{Name: "Ident", Pattern: `[a-zA-Z_$][a-zA-Z_$0-9]*`, Action: nil},
		{Name: "String", Pattern: `"[^"]*"`, Action: nil},
		{Name: "Whitespace", Pattern: `(?:[\s\t]|\n|(?:\r\n))+`, Action: nil},
		{Name: "Punct", Pattern: `[?,.|<>(){}=:;]`, Action: nil},
		{Name: "Comment", Pattern: `//.*`, Action: nil},
		{Name: "Reserved", Pattern: `(let|function)`, Action: nil},
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
