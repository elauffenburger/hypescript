use hypescript::parser;

#[macro_use]
mod macros {
    macro_rules! insta_test {
        ($($file:expr),*) => {{
            let parser = parser::Parser::new();

            let mut parsed = vec![];
            $(
                parsed.push(parser.parse($file, include_str!($file)).unwrap());
            )*

            insta::assert_debug_snapshot!(parsed);
        }};
    }
}

#[test]
fn all_the_things() {
    insta_test!("./parser_tests/all-the-things.ts");
}

#[test]
fn empty_export() {
    insta_test!("./parser_tests/empty-export.ts");
}

#[test]
fn mods_simple() {
    insta_test!(
        "./parser_tests/mods/simple/foo.ts",
        "./parser_tests/mods/simple/bar.ts"
    );
}