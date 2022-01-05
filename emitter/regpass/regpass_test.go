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
	_, err := runError(`let foo: string = 5;`)

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

	if !fn.Function.ImplicitReturnType.Equals(expectedType) {
		t.Errorf("wrong type inferred")
	}
}

func TestFunctionReturnTypeObjectLiteral(t *testing.T) {
	ctx := run(`
		function fn(): { msg: string } {
			return {
				msg: "hello world!",
			};
		}	
	`)

	fn := ctx.GlobalScope.IdentTypes["fn"].Function

	expectedType := &core.TypeSpec{
		Object: &core.Object{
			Fields: map[string]*core.ObjectTypeField{
				"msg": {
					Name: "msg",
					Type: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
				},
			},
		},
	}

	if !fn.ExplicitReturnType.Equals(expectedType) {
		t.Errorf("wrong explicit return type")
	}
}

func TestFunctionReturnTypeObjectLiteralMissingFields(t *testing.T) {
	_, err := runError(`
		function fn(): { name: string, age: number } {
			return {
				name: "Tommy Wiseau",
			};
		}	
	`)

	if err == nil {
		t.Errorf("expected error")
	}

	if _, ok := err.(regpass.FnRtnTypeMismatchError); !ok {
		t.Errorf("wrong error")
	}
}

func TestFunctionReturnTypeObjectLiteralMatchingInterface(t *testing.T) {
	ctx := run(`
		interface Foo {
			msg: string;
		}

		function fn(): Foo {
			return {
				msg: "hello world!",
			};
		}	
	`)

	fn := ctx.GlobalScope.IdentTypes["fn"].Function

	expectedExplRtnType := &core.TypeSpec{TypeReference: typeutils.StrRef("Foo")}
	if !fn.ExplicitReturnType.Equals(expectedExplRtnType) {
		t.Errorf("wrong explicit return type")
	}

	expectedImplRtnType := &core.TypeSpec{
		Object: &core.Object{
			Fields: map[string]*core.ObjectTypeField{
				"msg": {
					Name: "msg",
					Type: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
				},
			},
		},
	}
	if !fn.ImplicitReturnType.Equals(expectedImplRtnType) {
		t.Errorf("wrong impllicit return type")
	}
}

func TestFunctionReturnTypeObjectLiteralMatchingComplexInterface(t *testing.T) {
	ctx := run(`
		interface Foo {
			msg: string;
			bar: Bar;
		}

		interface Bar {
			name: string;
		}

		function fn(): Foo {
			return {
				msg: "hello world!",
				bar: {
					name: "sdlfkj",
				}
			};
		}	
	`)

	fn := ctx.GlobalScope.IdentTypes["fn"].Function

	expectedExplRtnType := &core.TypeSpec{TypeReference: typeutils.StrRef("Foo")}
	if !fn.ExplicitReturnType.Equals(expectedExplRtnType) {
		t.Errorf("wrong explicit return type")
	}

	expectedImplRtnType := &core.TypeSpec{
		Object: &core.Object{
			Fields: map[string]*core.ObjectTypeField{
				"msg": {
					Name: "msg",
					Type: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
				},
				"bar": {
					Name: "bar",
					Type: &core.TypeSpec{
						Object: &core.Object{
							Fields: map[string]*core.ObjectTypeField{
								"name": {
									Name: "name",
									Type: &core.TypeSpec{
										TypeReference: typeutils.StrRef("string"),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if !fn.ImplicitReturnType.Equals(expectedImplRtnType) {
		t.Errorf("wrong impllicit return type")
	}
}

func TestFunctionReturnTypeObjectLiteralSupersetOfInterface(t *testing.T) {
	ctx := run(`
		interface Foo {
			msg: string;
		}

		function fn(): Foo {
			return {
				msg: "The truth is out there",
				name: "Fox Mulder",
			};
		}	
	`)

	fn := ctx.GlobalScope.IdentTypes["fn"].Function

	expectedExplRtnType := &core.TypeSpec{TypeReference: typeutils.StrRef("Foo")}
	if !fn.ExplicitReturnType.Equals(expectedExplRtnType) {
		t.Errorf("wrong explicit return type")
	}

	expectedImplRtnType := &core.TypeSpec{
		Object: &core.Object{
			Fields: map[string]*core.ObjectTypeField{
				"msg": {
					Name: "msg",
					Type: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
				},
				"name": {
					Name: "name",
					Type: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
				},
			},
		},
	}
	if !fn.ImplicitReturnType.Equals(expectedImplRtnType) {
		t.Errorf("wrong impllicit return type")
	}
}

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
