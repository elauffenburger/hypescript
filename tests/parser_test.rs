use hypescript::{parser::{self, Module}, util::rcref};

#[test]
fn can_parse_src() {
    let parser = parser::Parser::new(rcref(Module{}));

    let parsed = parser.parse(
        r#"
        interface Foo {
            str: string;
            num: number;
            bar: Bar;
            baz: Baz;
            qux: string | number & Baz;
        }

        interface Bar {
            str: string;
        }

        interface Baz {
            name: string;
        }

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

        function blah3() {
            let foo: number = 5;

            return foo;
        }

        function returnsFn() {
            return function() {
                console.log("in nested!");
            };
        }

        function run(): void {
            function baz() {
                console.log("in baz!");
            }

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

            obj.qux.a = "hello, world!";

            console.log(obj.qux.a);
            obj.qux.foo();

            baz();

            returnsFn()();

            let foo: Foo = {
                str: "str",
                num: 5,
                bar: {
                    str: "hello",
                },
                baz: {
                    name: "world!",
                },
            };

            console.log(foo.bar.str);
            console.log(foo.baz.name);

            let bar = {
                name: "foo",
                sayName: function() {
                    console.log(this.getName());
                },
                getName: function(): string {
                    return this.name;
                },
            };

            bar.sayName();
        }

        run();
"#,
    )
    .unwrap();

    insta::assert_debug_snapshot!(parsed);
}

#[test]
fn can_parse_expr_with_ops() {
    let parser = parser::Parser::new(rcref(Module{}));

    let parsed = parser.parse(
        r#"
            console.log(foo.bar.str);
            console.log(foo.baz.name);
        "#,
    )
    .unwrap();

    insta::assert_debug_snapshot!(parsed);
}
