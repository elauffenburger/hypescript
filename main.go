package main

import (
	"elauffenburger/hypescript/emitter"
	"elauffenburger/hypescript/parser"
	"io"
	"log"
	"os"
	"path"
)

func main() {
	parser := parser.New()

	ast, err := parser.ParseString(`
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
					a: "a",
					foo: function() {
						console.log("in foo!");
					}
				} 
			};

			obj.qux.a = "hello, world!!\n";

			console.log(obj.qux.a);
			obj.qux.foo();
		}
	`)

	if err != nil {
		panic(err)
	}

	emitter := emitter.New()

	files, err := emitter.Emit(ast)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("%+v\n", err)
	}

	outputDir := path.Join(cwd, "./build")
	os.Mkdir("./build", 0777)

	for _, file := range files {
		bytes, err := io.ReadAll(file.Contents)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}

		err = os.WriteFile(path.Join(outputDir, file.Filename), bytes, 0777)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}
	}
}
