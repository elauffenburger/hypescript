package regpass_test

import (
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/core"
	"elauffenburger/hypescript/emitter/regpass"
	"elauffenburger/hypescript/parser"
	"elauffenburger/hypescript/typeutils"
	"testing"
)

func TestTypeAnnotationOnLetBinding(t *testing.T) {
	ctx := run(`let foo: number = 5;`)

	ty := ctx.GlobalScope.IdentTypes["foo"]

	if *ty.TypeReference != "number" {
		t.Errorf("foo was expected to have type number")
	}
}

func TestMismatchedTypeAnnotationOnLetBinding(t *testing.T) {
	_, err := runWithError(`let foo: string = 5;`)

	if err == nil {
		t.Errorf("expected error")
	}

	if _, ok := err.(regpass.TypeMismatchError); !ok {
		t.Errorf("unexpected error type")
	}
}

func TestFunctionReturnTypeInferenceFromVarType(t *testing.T) {
	ctx := run(`
		function fn() {
			let result = {
				name: "foo",
			};

			return result;
		}	
	`)

	fn := ctx.GlobalScope.IdentTypes["fn"]

	expectedType := &core.TypeSpec{
		Object: &core.Object{
			Fields: map[string]*core.ObjectTypeField{
				"name": {
					Name: "name",
					Type: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
				},
			},
		},
	}

	if !fn.Function.ImplicitReturnType.EqualsReferencing(expectedType) {
		t.Errorf("wrong type inferred")
	}
}

func runWithError(code string) (*regpass.Context, error) {
	ctx := regpass.NewContext()
	err := ctx.Run(parse(code))

	return ctx, err
}

func run(code string) *regpass.Context {
	ctx, err := runWithError(code)
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
