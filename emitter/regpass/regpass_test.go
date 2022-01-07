package regpass_test

import (
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/regpass"
	"elauffenburger/hypescript/parser"
)

func runError(code string) (*regpass.Context, error) {
	ctx := regpass.NewContext()
	err := ctx.Run(parse(code))

	return ctx, err
}

func run(code string) *regpass.Context {
	ctx, err := runError(code)
	if err != nil {
		panic(err)
	}

	return ctx
}

func parse(code string) *ast.TS {
	res, err := parser.New().ParseString(code)
	if err != nil {
		panic(err)
	}

	return res
}
