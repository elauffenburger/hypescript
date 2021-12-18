package main

import (
	"bufio"
	"elauffenburger/hypescript/emitter"
	"elauffenburger/hypescript/parser"
	"io"
	"strings"
	"testing"

	_ "embed"

	"github.com/pkg/errors"
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
			let bar = foo;
			bar = "bar";

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
	e := emitter.New()

	ast, err := parser.New().ParseString(code)
	if err != nil {
		t.Error(err)
	}

	files, err := e.Emit(ast)
	if err != nil {
		t.Error(err)
	}

	for _, file := range files {
		if file.Filename == "main.cpp" {
			reader := bufio.NewReader(file.Contents)

			contents, err := io.ReadAll(reader)
			if err != nil {
				panic(errors.Wrap(err, "could not read main.cpp"))
			}

			return string(contents)
		}
	}

	panic("no main.cpp emitted")
}

func normalizeCode(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
}
