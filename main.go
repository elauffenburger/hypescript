package main

import (
	"elauffenburger/hypescript/emitter"
	"elauffenburger/hypescript/parser"
	"fmt"
	"strings"
)

func main() {
	parser := parser.New()

	ast, err := parser.ParseString(`
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

		function main(): void {
			let obj = { foo: "bar", baz: 5, qux: { a: "a" } };
			obj.qux.a;

			blah();
		}
	`)

	if err != nil {
		panic(err)
	}

	output := strings.Builder{}
	emitter := emitter.New(&output)

	err = emitter.Emit(ast)
	if err != nil {
		panic(err)
	}

	fmt.Println(output.String())
}
