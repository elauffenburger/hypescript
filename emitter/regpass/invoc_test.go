package regpass_test

import "testing"

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
