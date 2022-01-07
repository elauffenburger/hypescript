package regpass_test

import (
	"elauffenburger/hypescript/emitter/core"
	"elauffenburger/hypescript/emitter/regpass"
	"elauffenburger/hypescript/typeutils"
	"testing"
)

type fnTest struct {
	code       string
	explSpec   *core.TypeSpec
	implSpec   *core.TypeSpec
	errChecker func(error)
}

func testFn(t *testing.T, tc fnTest) {
	switch {
	case tc.implSpec != nil, tc.explSpec != nil:
		ctx := run(tc.code)
		fn := ctx.GlobalScope.IdentTypes["fn"]

		if tc.implSpec != nil && !fn.Function.ImplicitReturnType.Equals(tc.implSpec) {
			t.Error("wrong implicit return type inferred")
		}

		if tc.explSpec != nil && !fn.Function.ExplicitReturnType.Equals(tc.explSpec) {
			t.Error("wrong explicit return type inferred")
		}

	case tc.errChecker != nil:
		_, err := runError(tc.code)
		if err == nil {
			t.Error("expected error")
		}

		tc.errChecker(err)
	}
}

func TestFunctionReturnTypeInferredFromVar(t *testing.T) {
	testFn(t, fnTest{
		code: `
			function fn() {
				let result = {
					name: "foo",
				};

				return result;
			}	
		`,
		implSpec: &core.TypeSpec{
			Object: &core.Object{
				Fields: map[string]*core.ObjectTypeField{
					"name": {
						Name: "name",
						Type: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
					},
				},
			},
		},
	})
}

func TestFunctionCanHaveExplicitObjectLiteralReturnType(t *testing.T) {
	testFn(t, fnTest{
		code: `
			function fn(): { msg: string } {
				return {
					msg: "hello world!",
				};
			}	
		`,
		explSpec: &core.TypeSpec{
			Object: &core.Object{
				Fields: map[string]*core.ObjectTypeField{
					"msg": {
						Name: "msg",
						Type: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
					},
				},
			},
		},
	})
}

func TestFunctionImplicitObjectLiteralRequiresAllFields(t *testing.T) {
	testFn(t, fnTest{
		code: `
			function fn(): { name: string, age: number } {
				return {
					name: "Tommy Wiseau",
				};
			}	
		`,
		errChecker: func(err error) {
			if _, ok := err.(regpass.FnRtnTypeMismatchError); !ok {
				t.Errorf("wrong error")
			}
		},
	})
}

func TestFunctionHasObjectLiteralSatisfyInterfaceReturnType(t *testing.T) {
	testFn(t, fnTest{
		code: `
			interface Foo {
				msg: string;
			}

			function fn(): Foo {
				return {
					msg: "hello world!",
				};
			}	
		`,
		explSpec: &core.TypeSpec{TypeReference: typeutils.StrRef("Foo")},
		implSpec: &core.TypeSpec{
			Object: &core.Object{
				Fields: map[string]*core.ObjectTypeField{
					"msg": {
						Name: "msg",
						Type: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
					},
				},
			},
		},
	})
}

func TestFunctionReturnsObjectLiteralSatisfyingComplexInterface(t *testing.T) {
	testFn(t, fnTest{
		code: `
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
		`,
		explSpec: &core.TypeSpec{TypeReference: typeutils.StrRef("Foo")},
		implSpec: &core.TypeSpec{
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
		},
	})
}

func TestFunctionReturnObjectLiteralSubsetOfExplicitInterface(t *testing.T) {
	testFn(t, fnTest{
		code: `
			interface Foo {
				msg: string;
			}

			function fn(): Foo {
				return {
					msg: "The truth is out there",
					name: "Fox Mulder",
				};
			}	
		`,
		explSpec: &core.TypeSpec{TypeReference: typeutils.StrRef("Foo")},
		implSpec: &core.TypeSpec{
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
		},
	})
}
