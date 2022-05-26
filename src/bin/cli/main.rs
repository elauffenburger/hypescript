use hypescript::parser;

const SRC: &'static str = r#"
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
"#;

fn main() -> Result<(), String> {
    parser::parse(SRC).and(Ok(()))
}

#[cfg(test)]
mod test {
    use super::parser::*;
    use super::*;

    #[test]
    fn can_parse_src() {
        let parsed = parser::parse(SRC).unwrap();

        assert_eq!(
            parsed,
            ParseResult {
                top_level_constructs: vec![
                    TopLevelConstruct::Interface(Interface {
                        name: "Foo".into(),
                        fields: vec![
                            InterfaceField {
                                name: "str".into(),
                                optional: false,
                                typ: TypeIdent {
                                    head: TypeIdentType::Name("string".into(),),
                                    rest: vec![],
                                },
                            },
                            InterfaceField {
                                name: "num".into(),
                                optional: false,
                                typ: TypeIdent {
                                    head: TypeIdentType::Name("number".into(),),
                                    rest: vec![],
                                },
                            },
                            InterfaceField {
                                name: "bar".into(),
                                optional: false,
                                typ: TypeIdent {
                                    head: TypeIdentType::Name("Bar".into(),),
                                    rest: vec![],
                                },
                            },
                            InterfaceField {
                                name: "baz".into(),
                                optional: false,
                                typ: TypeIdent {
                                    head: TypeIdentType::Name("Baz".into(),),
                                    rest: vec![],
                                },
                            },
                            InterfaceField {
                                name: "qux".into(),
                                optional: false,
                                typ: TypeIdent {
                                    head: TypeIdentType::Name("string".into(),),
                                    rest: vec![
                                        TypeIdentPart::Union(TypeIdentType::Name("number".into(),),),
                                        TypeIdentPart::Sum(TypeIdentType::Name("Baz".into(),),),
                                    ],
                                },
                            },
                        ],
                        methods: vec![],
                    },),
                    TopLevelConstruct::Interface(Interface {
                        name: "Bar".into(),
                        fields: vec![InterfaceField {
                            name: "str".into(),
                            optional: false,
                            typ: TypeIdent {
                                head: TypeIdentType::Name("string".into(),),
                                rest: vec![],
                            },
                        },],
                        methods: vec![],
                    },),
                    TopLevelConstruct::Interface(Interface {
                        name: "Baz".into(),
                        fields: vec![InterfaceField {
                            name: "name".into(),
                            optional: false,
                            typ: TypeIdent {
                                head: TypeIdentType::Name("string".into(),),
                                rest: vec![],
                            },
                        },],
                        methods: vec![],
                    },),
                    TopLevelConstruct::StmtOrExpr(StmtOrExpr::Stmt(Stmt::Expr(Expr::FnInst(
                        FnInst {
                            name: Some("foo".into(),),
                            params: vec![
                                FnParam {
                                    name: "a".into(),
                                    optional: false,
                                    typ: Some(TypeIdent {
                                        head: TypeIdentType::Name("string".into(),),
                                        rest: vec![],
                                    },),
                                },
                                FnParam {
                                    name: "b".into(),
                                    optional: false,
                                    typ: Some(TypeIdent {
                                        head: TypeIdentType::Name("number".into(),),
                                        rest: vec![],
                                    },),
                                },
                            ],
                            body: vec![
                                StmtOrExpr::Stmt(Stmt::LetDecl {
                                    name: "ay".into(),
                                    typ: None,
                                    assignment: Some(Expr::Num(5.0,),),
                                },),
                                StmtOrExpr::Stmt(Stmt::LetDecl {
                                    name: "bee".into(),
                                    typ: None,
                                    assignment: Some(Expr::Str("bar".into(),),),
                                },),
                                StmtOrExpr::Stmt(Stmt::ReturnExpr(Expr::Ident("ay".into(),),),),
                            ],
                            return_type: Some(TypeIdent {
                                head: TypeIdentType::Name("number".into(),),
                                rest: vec![],
                            },),
                        },
                    ),),),),
                    TopLevelConstruct::StmtOrExpr(StmtOrExpr::Stmt(Stmt::Expr(Expr::FnInst(
                        FnInst {
                            name: Some("blah".into(),),
                            params: vec![],
                            body: vec![
                                StmtOrExpr::Stmt(Stmt::LetDecl {
                                    name: "foo".into(),
                                    typ: None,
                                    assignment: Some(Expr::Str("asdf".into(),),),
                                },),
                                StmtOrExpr::Stmt(Stmt::LetDecl {
                                    name: "bar".into(),
                                    typ: None,
                                    assignment: Some(Expr::Ident("foo".into(),),),
                                },),
                                StmtOrExpr::Stmt(Stmt::Expr(Expr::IdentAssignment(Box::new(
                                    IdentAssignment {
                                        ident: "bar".into(),
                                        assignment: Expr::Str("bar".into(),),
                                    },
                                )),),),
                                StmtOrExpr::Stmt(Stmt::ReturnExpr(Expr::Ident("foo".into(),),),),
                            ],
                            return_type: None,
                        },
                    ),),),),
                    TopLevelConstruct::StmtOrExpr(StmtOrExpr::Stmt(Stmt::Expr(Expr::FnInst(
                        FnInst {
                            name: Some("blah2".into(),),
                            params: vec![],
                            body: vec![],
                            return_type: None,
                        },
                    ),),),),
                    TopLevelConstruct::StmtOrExpr(StmtOrExpr::Stmt(Stmt::Expr(Expr::FnInst(
                        FnInst {
                            name: Some("blah3".into(),),
                            params: vec![],
                            body: vec![
                                StmtOrExpr::Stmt(Stmt::LetDecl {
                                    name: "foo".into(),
                                    typ: Some(TypeIdent {
                                        head: TypeIdentType::Name("number".into(),),
                                        rest: vec![],
                                    },),
                                    assignment: Some(Expr::Num(5.0,),),
                                },),
                                StmtOrExpr::Stmt(Stmt::ReturnExpr(Expr::Ident("foo".into(),),),),
                            ],
                            return_type: None,
                        },
                    ),),),),
                    TopLevelConstruct::StmtOrExpr(StmtOrExpr::Stmt(Stmt::Expr(Expr::FnInst(
                        FnInst {
                            name: Some("returnsFn".into(),),
                            params: vec![],
                            body: vec![StmtOrExpr::Stmt(Stmt::ReturnExpr(Expr::FnInst(FnInst {
                                name: None,
                                params: vec![],
                                body: vec![StmtOrExpr::Stmt(Stmt::Expr(Expr::ChainedObjOp(
                                    ChainedObjOp {
                                        accessable: Accessable::Ident("console".into(),),
                                        obj_ops: vec![
                                            ObjOp::Access("log".into(),),
                                            ObjOp::Invoc {
                                                args: vec![Expr::Str("in nested!".into(),),],
                                            },
                                        ],
                                        assignment: None,
                                    },
                                ),),),],
                                return_type: None,
                            },),),),],
                            return_type: None,
                        },
                    ),),),),
                    TopLevelConstruct::StmtOrExpr(StmtOrExpr::Stmt(Stmt::Expr(Expr::FnInst(
                        FnInst {
                            name: Some("run".into(),),
                            params: vec![],
                            body: vec![
                                StmtOrExpr::Stmt(Stmt::Expr(Expr::FnInst(FnInst {
                                    name: Some("baz".into(),),
                                    params: vec![],
                                    body: vec![StmtOrExpr::Stmt(Stmt::Expr(Expr::ChainedObjOp(ChainedObjOp {
                                        accessable: Accessable::Ident("console".into(),),
                                        obj_ops: vec![
                                            ObjOp::Access("log".into(),),
                                            ObjOp::Invoc {
                                                args: vec![Expr::Str("in baz!".into(),),],
                                            },
                                        ],
                                        assignment: None,
                                    },),),),],
                                    return_type: None,
                                },),),),
                                StmtOrExpr::Stmt(Stmt::LetDecl {
                                    name: "obj".into(),
                                    typ: None,
                                    assignment: Some(Expr::ObjInst(ObjInst {
                                        fields: vec![
                                            ObjFieldInst {
                                                name: "foo".into(),
                                                value: Expr::Str("bar".into(),),
                                            },
                                            ObjFieldInst {
                                                name: "baz".into(),
                                                value: Expr::Num(5.0,),
                                            },
                                            ObjFieldInst {
                                                name: "qux".into(),
                                                value: Expr::ObjInst(ObjInst {
                                                    fields: vec![
                                                        ObjFieldInst {
                                                            name: "a".into(),
                                                            value: Expr::Str("a".into(),),
                                                        },
                                                        ObjFieldInst {
                                                            name: "foo".into(),
                                                            value: Expr::FnInst(FnInst {
                                                                name: None,
                                                                params: vec![],
                                                                body: vec![StmtOrExpr::Stmt(Stmt::Expr(
                                                                    Expr::ChainedObjOp(ChainedObjOp {
                                                                        accessable: Accessable::Ident(
                                                                            "console".into(),
                                                                        ),
                                                                        obj_ops: vec![
                                                                            ObjOp::Access("log".into(),),
                                                                            ObjOp::Invoc {
                                                                                args: vec![Expr::Str(
                                                                                    "in foo!"
                                                                                        .into(),
                                                                                ),],
                                                                            },
                                                                        ],
                                                                        assignment: None,
                                                                    },),
                                                                ),),],
                                                                return_type: None,
                                                            },),
                                                        },
                                                    ],
                                                },),
                                            },
                                        ],
                                    },),),
                                },),
                                StmtOrExpr::Stmt(Stmt::Expr(Expr::ChainedObjOp(ChainedObjOp {
                                    accessable: Accessable::Ident("obj".into(),),
                                    obj_ops: vec![ObjOp::Access("qux".into(),), ObjOp::Access("a".into(),),],
                                    assignment: Some(Box::new(Expr::Str("hello, world!".into(),),)),
                                },),),),
                                StmtOrExpr::Stmt(Stmt::Expr(Expr::ChainedObjOp(ChainedObjOp {
                                    accessable: Accessable::Ident("console".into(),),
                                    obj_ops: vec![
                                        ObjOp::Access("log".into(),),
                                        ObjOp::Invoc {
                                            args: vec![Expr::ChainedObjOp(ChainedObjOp {
                                                accessable: Accessable::Ident("obj".into(),),
                                                obj_ops: vec![
                                                    ObjOp::Access("qux".into(),),
                                                    ObjOp::Access("a".into(),),
                                                ],
                                                assignment: None,
                                            },),],
                                        },
                                    ],
                                    assignment: None,
                                },),),),
                                StmtOrExpr::Stmt(Stmt::Expr(Expr::ChainedObjOp(ChainedObjOp {
                                    accessable: Accessable::Ident("obj".into(),),
                                    obj_ops: vec![
                                        ObjOp::Access("qux".into(),),
                                        ObjOp::Access("foo".into(),),
                                        ObjOp::Invoc { args: vec![] },
                                    ],
                                    assignment: None,
                                },),),),
                                StmtOrExpr::Stmt(Stmt::Expr(Expr::ChainedObjOp(ChainedObjOp {
                                    accessable: Accessable::Ident("baz".into(),),
                                    obj_ops: vec![ObjOp::Invoc { args: vec![] },],
                                    assignment: None,
                                },),),),
                                StmtOrExpr::Stmt(Stmt::ReturnExpr(Expr::ChainedObjOp(ChainedObjOp {
                                    accessable: Accessable::Ident("sFn".into(),),
                                    obj_ops: vec![ObjOp::Invoc { args: vec![] }, ObjOp::Invoc { args: vec![] },],
                                    assignment: None,
                                },),),),
                                StmtOrExpr::Stmt(Stmt::LetDecl {
                                    name: "foo".into(),
                                    typ: Some(TypeIdent {
                                        head: TypeIdentType::Name("Foo".into(),),
                                        rest: vec![],
                                    },),
                                    assignment: Some(Expr::ObjInst(ObjInst {
                                        fields: vec![
                                            ObjFieldInst {
                                                name: "str".into(),
                                                value: Expr::Str("str".into(),),
                                            },
                                            ObjFieldInst {
                                                name: "num".into(),
                                                value: Expr::Num(5.0,),
                                            },
                                            ObjFieldInst {
                                                name: "bar".into(),
                                                value: Expr::ObjInst(ObjInst {
                                                    fields: vec![ObjFieldInst {
                                                        name: "str".into(),
                                                        value: Expr::Str("hello".into(),),
                                                    },],
                                                },),
                                            },
                                            ObjFieldInst {
                                                name: "baz".into(),
                                                value: Expr::ObjInst(ObjInst {
                                                    fields: vec![ObjFieldInst {
                                                        name: "name".into(),
                                                        value: Expr::Str("world!".into(),),
                                                    },],
                                                },),
                                            },
                                        ],
                                    },),),
                                },),
                                StmtOrExpr::Stmt(Stmt::Expr(Expr::ChainedObjOp(ChainedObjOp {
                                    accessable: Accessable::Ident("console".into(),),
                                    obj_ops: vec![
                                        ObjOp::Access("log".into(),),
                                        ObjOp::Invoc {
                                            args: vec![Expr::ChainedObjOp(ChainedObjOp {
                                                accessable: Accessable::Ident("foo".into(),),
                                                obj_ops: vec![
                                                    ObjOp::Access("bar".into(),),
                                                    ObjOp::Access("str".into(),),
                                                ],
                                                assignment: None,
                                            },),],
                                        },
                                    ],
                                    assignment: None,
                                },),),),
                                StmtOrExpr::Stmt(Stmt::Expr(Expr::ChainedObjOp(ChainedObjOp {
                                    accessable: Accessable::Ident("console".into(),),
                                    obj_ops: vec![
                                        ObjOp::Access("log".into(),),
                                        ObjOp::Invoc {
                                            args: vec![Expr::ChainedObjOp(ChainedObjOp {
                                                accessable: Accessable::Ident("foo".into(),),
                                                obj_ops: vec![
                                                    ObjOp::Access("baz".into(),),
                                                    ObjOp::Access("name".into(),),
                                                ],
                                                assignment: None,
                                            },),],
                                        },
                                    ],
                                    assignment: None,
                                },),),),
                                StmtOrExpr::Stmt(Stmt::LetDecl {
                                    name: "bar".into(),
                                    typ: None,
                                    assignment: Some(Expr::ObjInst(ObjInst {
                                        fields: vec![
                                            ObjFieldInst {
                                                name: "name".into(),
                                                value: Expr::Str("foo".into(),),
                                            },
                                            ObjFieldInst {
                                                name: "sayName".into(),
                                                value: Expr::FnInst(FnInst {
                                                    name: None,
                                                    params: vec![],
                                                    body: vec![StmtOrExpr::Stmt(Stmt::Expr(Expr::ChainedObjOp(
                                                        ChainedObjOp {
                                                            accessable: Accessable::Ident("console".into(),),
                                                            obj_ops: vec![
                                                                ObjOp::Access("log".into(),),
                                                                ObjOp::Invoc {
                                                                    args: vec![Expr::ChainedObjOp(
                                                                        ChainedObjOp {
                                                                            accessable: Accessable::Ident(
                                                                                "this".into(),
                                                                            ),
                                                                            obj_ops: vec![
                                                                                ObjOp::Access(
                                                                                    "getName"
                                                                                        .into(),
                                                                                ),
                                                                                ObjOp::Invoc {
                                                                                    args: vec![]
                                                                                },
                                                                            ],
                                                                            assignment: None,
                                                                        },
                                                                    ),],
                                                                },
                                                            ],
                                                            assignment: None,
                                                        },
                                                    ),
                                                ),
                                            ),
                                            ],
                                                    return_type: None,
                                                },),
                                            },
                                            ObjFieldInst {
                                                name: "getName".into(),
                                                value: Expr::FnInst(FnInst {
                                                    name: None,
                                                    params: vec![],
                                                    body: vec![StmtOrExpr::Stmt(Stmt::ReturnExpr(Expr::ChainedObjOp(
                                                        ChainedObjOp {
                                                            accessable: Accessable::Ident("this".into(),),
                                                            obj_ops: vec![ObjOp::Access("name".into(),),],
                                                            assignment: None,
                                                        },
                                                    ),),),],
                                                    return_type: Some(TypeIdent {
                                                        head: TypeIdentType::Name("string".into(),),
                                                        rest: vec![],
                                                    },),
                                                },),
                                            },
                                        ],
                                    },),),
                                },),
                                StmtOrExpr::Stmt(Stmt::Expr(Expr::ChainedObjOp(ChainedObjOp {
                                    accessable: Accessable::Ident("bar".into(),),
                                    obj_ops: vec![
                                        ObjOp::Access("sayName".into(),),
                                        ObjOp::Invoc { args: vec![] },
                                    ],
                                    assignment: None,
                                },),),),
                            ],
                            return_type: Some(TypeIdent {
                                head: TypeIdentType::Name("void".into(),),
                                rest: vec![],
                            },),
                        },
                    ),),),),
                    TopLevelConstruct::StmtOrExpr(StmtOrExpr::Stmt(Stmt::Expr(Expr::ChainedObjOp(ChainedObjOp {
                        accessable: Accessable::Ident("run".into(),),
                        obj_ops: vec![ObjOp::Invoc { args: vec![] },],
                        assignment: None,
                    },),),),),
                ],
            }
        );
    }
}
