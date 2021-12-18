package main

import (
	"elauffenburger/hypescript/emitter"
	"elauffenburger/hypescript/parser"
	"strings"
	"testing"

	_ "embed"

	"github.com/sergi/go-diff/diffmatchpatch"
)

//go:embed snapshots/emitter_test/complex.cpp
var testEmitForComplexCodeSnapshot string

func TestEmitForComplexCode(t *testing.T) {
	code := `
		function foo(a: string, b: num): num {
			let ay = 5;
			let bee = "bar";

			return ay;
		}

		function blah() {
			let foo = "asdf";

			let bar = console;
			bar = console;

			return foo;
		}

		function blah2() {}

		function main(): void {
			let obj = { 
				foo: "bar", 
				baz: 5, 
				qux: { 
					a: "a"
				} 
			};

			obj.qux.a = "b";

			blah();
		}
	`

	assertCodeMatchesSnapshot(t, code, testEmitForComplexCodeSnapshot)
}

func assertCodeMatchesSnapshot(t *testing.T, code, snapshot string) {
	emitted := emitForString(t, code)

	differ := diffmatchpatch.New()
	diffs := differ.DiffMain(normalizeCode(snapshot), normalizeCode(emitted), false)
	for _, diff := range diffs {
		if diff.Type != diffmatchpatch.DiffEqual {
			t.Errorf("\n%s\n", differ.DiffPrettyText(diffs))
		}
	}
}

func emitForString(t *testing.T, code string) string {
	result := strings.Builder{}
	e := emitter.New(&result)

	ast, err := parser.New().ParseString(code)
	if err != nil {
		t.Error(err)
	}

	err = e.Emit(ast)
	if err != nil {
		t.Error(err)
	}

	return result.String()
}

func normalizeCode(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
}
