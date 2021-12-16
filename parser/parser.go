package parser

import (
	"elauffenburger/hypescript/ast"

	"github.com/alecthomas/participle/v2"
	"github.com/pkg/errors"
)

type Parser interface {
	ParseString(str string) (*ast.TS, error)
}

type parser struct {
}

func (p parser) ParseString(str string) (*ast.TS, error) {
	parser, err := participle.Build(&ast.TS{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to build AST")
	}

	ast := &ast.TS{}
	err = parser.ParseString("temp.ts", str, ast)

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse program")
	}

	return ast, nil
}

func New() Parser {
	return &parser{}
}
