package regpass_test

import (
	"elauffenburger/hypescript/emitter/core"
	"elauffenburger/hypescript/typeutils"
	"testing"
)

func TestInvokeFnWithoutOptionalParams(t *testing.T) {
	run(`
		function fn(foo: string) {}

		fn("hello, world!");
	`)
}

func TestInvokeFnWithOptionalParamsSatisfyingParams(t *testing.T) {
	run(`
		function fn(foo: string, bar?: number) {}

		fn("hello, world!", 42);
	`)
}

func TestInvokeFnWithOptionalParamsWihtoutSatisfyingParams(t *testing.T) {
	run(`
		function fn(foo: string, bar?: number) {}

		fn("hello, world!");
	`)
}

func TestInvokeFnWithoutOptionalParamsWithoutSatisfyingParams(t *testing.T) {
	_, err := runError(`
		function fn(foo: string, bar: number) {}

		fn("hello, world!");
	`)
	if err == nil {
		t.Error("expected err")
	}
}

func TestInvokeObjectFnWithThis(t *testing.T) {
	ctx := run(`
		let obj = {
			name: "fmulder",
			getName: function(): string {
				return this.name;
			}
		};
	`)

	obj, err := ctx.GlobalScope.IdentType("obj")
	if err != nil {
		t.Error(err)
	}

	spec := &core.TypeSpec{
		Object: core.NewObject(
			[]*core.Member{
				{
					Field: &core.ObjectTypeField{Name: "name", Type: &core.TypeSpec{TypeReference: typeutils.StrRef("string")}},
				}, {
					Field: &core.ObjectTypeField{
						Name: "getName",
						Type: &core.TypeSpec{
							Function: &core.Function{
								ExplicitReturnType: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
								ImplicitReturnType: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
							},
						},
					},
				},
			},
		),
	}

	if !obj.Equals(spec) {
		t.Errorf("wrong type for obj")
	}
}

func TestInvokeObjectFnWithFnInvokeOnThisBeforeFnDef(t *testing.T) {
	ctx := run(`
		let obj = {
			name: "fmulder",
			sayName: function() {
				console.log(this.getName());
			},
			getName: function(): string {
				return this.name;
			},
		};
	`)

	obj, err := ctx.GlobalScope.IdentType("obj")
	if err != nil {
		t.Error(err)
	}

	spec := &core.TypeSpec{
		Object: core.NewObject(
			[]*core.Member{
				{
					Field: &core.ObjectTypeField{Name: "name", Type: &core.TypeSpec{TypeReference: typeutils.StrRef("string")}},
				},
				{
					Field: &core.ObjectTypeField{
						Name: "sayName",
						Type: &core.TypeSpec{
							Function: &core.Function{
								ImplicitReturnType: &core.TypeSpec{TypeReference: typeutils.StrRef("void")},
							},
						},
					},
				},
				{
					Field: &core.ObjectTypeField{
						Name: "getName",
						Type: &core.TypeSpec{
							Function: &core.Function{
								ExplicitReturnType: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
								ImplicitReturnType: &core.TypeSpec{TypeReference: typeutils.StrRef("string")},
							},
						},
					},
				},
			},
		),
	}

	if !obj.Equals(spec) {
		t.Errorf("wrong type for obj")
	}
}
