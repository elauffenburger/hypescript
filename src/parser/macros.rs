
macro_rules! assert_rule {
    ($pair:ident, $expected:path) => {{
        let rule = $pair.as_rule();
        let expected = $expected;
        if rule != $expected {
            panic!("expected {expected:?}, go: {rule:?}")
        }
    }}
}