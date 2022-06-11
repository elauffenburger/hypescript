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
