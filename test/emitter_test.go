package main

import (
	"bufio"
	"elauffenburger/hypescript/emitter"
	"elauffenburger/hypescript/parser"
	"io"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/pkg/errors"
)

func TestEmitForComplexCode(t *testing.T) {
	code := `
		function foo(a: string, b: number): number {
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

	assertCodeMatchesSnapshot(t, code)
}

func assertCodeMatchesSnapshot(t *testing.T, code string) {
	emitted := emitForString(t, code)

	cupaloy.Snapshot(t, emitted)
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
