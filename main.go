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
			let foo = 5;
			let bar = "bar";

			return foo;
		}

		function blah() {
			let foo = "asdf";

			return foo;
		}

		function main(): num {
			return 0;
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
