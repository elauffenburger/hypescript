package regpass_test

import (
	"elauffenburger/hypescript/emitter/core"
	"elauffenburger/hypescript/emitter/regpass"
	"elauffenburger/hypescript/typeutils"
	"testing"
)

type varTest struct {
	code       string
	spec       *core.TypeSpec
	errChecker func(error)
}

func testVar(t *testing.T, tc varTest) {
	switch {
	case tc.spec != nil:
		ctx := run(tc.code)

		ty := ctx.GlobalScope.IdentTypes["test"]
		if !ty.Equals(tc.spec) {
			t.Error("wrong var type for var under test")
		}

	case tc.errChecker != nil:
		_, err := runError(tc.code)
		if err == nil {
			t.Errorf("expected error")
		}

		tc.errChecker(err)
	}
}

func TestVarAnnotation(t *testing.T) {
	testVar(t, varTest{
		code: `let test: number = 5;`,
		spec: &core.TypeSpec{TypeReference: typeutils.StrRef("number")},
	})
}

func TestVarAnnotationMismatch(t *testing.T) {
	testVar(t, varTest{
		code: `let test: string = 5;`,
		errChecker: func(err error) {
			if _, ok := err.(regpass.TypeMismatchError); !ok {
				t.Errorf("unexpected error type")
			}
		},
	})
}

func TestVarSatisfyingUnionAnnotation(t *testing.T) {
	testVar(t, varTest{
		code: `let test: string | number = 5;`,
		spec: &core.TypeSpec{
			Union: &core.Union{
				Types: map[*core.TypeSpec]bool{
					{TypeReference: typeutils.StrRef("string")}: true,
					{TypeReference: typeutils.StrRef("number")}: true,
				},
			},
		},
	})
}

func TestVarNotSatisfyingUnionAnnotation(t *testing.T) {
	testVar(t, varTest{
		code: `
			interface Foo {}

			let test: string | Foo = 5;
		`,
		errChecker: func(err error) {
			if _, ok := err.(regpass.TypeMismatchError); !ok {
				t.Error("wrong error type")
			}
		},
	})
}

func TestInterfaceWithOptionalFieldsHasFieldsSatisfied(t *testing.T) {
	testVar(t, varTest{
		code: `
			interface Foo {
				name: string;
				age?: number;
			}

			let test: Foo = {
				name: "Eric",
				age: 30,
			};
		`,
		spec: &core.TypeSpec{
			Object: core.NewObject(
				[]*core.Member{
					{
						Field: &core.ObjectTypeField{
							Name: "name",
							Type: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
						},
					},
					{
						Field: &core.ObjectTypeField{
							Name: "age",
							Type: &core.TypeSpec{
								Union: &core.Union{
									Types: map[*core.TypeSpec]bool{
										{TypeReference: typeutils.StrRef("number")}: true,
									},
								},
							},
						},
					},
				},
			),
		},
	})
}

func TestInterfaceWithOptionalFieldsHasFieldsUnsatisfied(t *testing.T) {
	testVar(t, varTest{
		code: `
			interface Foo {
				name: string;
				age?: number;
			}

			let test: Foo = {
				name: "Tommy Wiseau",
			};
		`,
		spec: &core.TypeSpec{
			Object: core.NewObject(
				[]*core.Member{
					{
						Field: &core.ObjectTypeField{
							Name: "name",
							Type: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
						},
					},
				},
			),
		},
	})
}
