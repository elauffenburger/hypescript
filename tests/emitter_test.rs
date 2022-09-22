use hypescript::{emitter, parser};

#[test]
fn can_emit_src() {
    let parsed = parser::parse(
        r#"
        function main() {
            console.log("hello, world!");
        } 

        main();
        "#,
    )
    .unwrap();

    insta::assert_debug_snapshot!(emitter::Emitter::new().emit(parsed).unwrap());
}

#[test]
fn can_emit_complex_src() {
    let parsed = parser::parse(
        r#"
        interface Foo {
            str: string;
            num: number;
            bar: Bar;
            baz: Baz;
            qux?: string | number & Baz;
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

    insta::assert_debug_snapshot!(emitter::Emitter::new().emit(parsed).unwrap());
}

#[test]
fn can_use_for_loop() {
    let parsed = parser::parse(
        r#"
        for (let i = 0; i < 10; i++) {
            console.log(i);
        }
        "#,
    )
    .unwrap();

    insta::assert_debug_snapshot!(emitter::Emitter::new().emit(parsed).unwrap());
}

#[test]
fn can_emit_fizzbuzz() {
    let parsed = parser::parse(
        r#"
            function fizzbuzz(n: number): void {
                for (let i = 1; i < n + 1; i++) {
                    let fizz = i % 3 == 0;
                    let buzz = i % 5 == 0;

                    if (fizz && buzz) {
                        console.log("fizzbuzz");
                    } else if (fizz) {
                        console.log("fizz");
                    } else if (buzz) {
                        console.log("buzz");
                    } else {
                        console.log(i);
                    }
                }
            }
        "#,
    )
    .unwrap();

    insta::assert_debug_snapshot!(emitter::Emitter::new().emit(parsed).unwrap());
}

#[test]
fn can_iife() {
    let parsed = parser::parse(
        r#"
            (function(){ return 42; })()
        "#,
    )
    .unwrap();

    insta::assert_debug_snapshot!(emitter::Emitter::new().emit(parsed).unwrap());
}

#[test]
fn can_emit_expr_with_ops() {
    let parsed = parser::parse(
        r#"
            let foo = {
                bar: {
                    str: "hello, world!"
                },
                baz: {
                    name: "i'm a baz!"
                }
            }

            console.log(foo.bar.str);
            console.log(foo.baz.name);
        "#,
    )
    .unwrap();

    insta::assert_debug_snapshot!(emitter::Emitter::new().emit(parsed).unwrap());
}

#[test]
fn can_emit_expr_with_assignment() {
    let parsed = parser::parse(
        r#"
            let foo = {
                name: "i'm a bar!"
            };

            foo.name = "i'm a foo!";
        "#,
    )
    .unwrap();

    insta::assert_debug_snapshot!(emitter::Emitter::new().emit(parsed).unwrap());
}

#[test]
fn can_emit_subexpr() {
    let parsed = parser::parse(
        r#"
            (42);
        "#,
    )
    .unwrap();

    insta::assert_debug_snapshot!(emitter::Emitter::new().emit(parsed).unwrap());
}