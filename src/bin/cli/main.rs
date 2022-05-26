use hypescript::ast;
use pest::{iterators::Pair, Parser};

const SRC: &'static str = r#"
    interface Foo {
        str: string;
        num: number;
        bar: Bar;
        baz: Baz;
        qux: string | number;
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
    let root_pairs = ast::IdentParser::parse(ast::Rule::ts, SRC).map_err(|e| format!("{e}"))?;

    let mut top_level_constructs = vec![];
    for pair in root_pairs {
        match pair.as_rule() {
            ast::Rule::iface_defn => {
                let iface = TopLevelConstruct::Interface(parse_interface(pair)?);
                println!("{iface:?}");

                top_level_constructs.push(iface);
            }
            ast::Rule::stmt_or_expr => {}
            ast::Rule::EOI => {}
            r @ _ => return Err(format!("unknown top-level construct: {r:?}")),
        }
    }

    Ok(())
}

fn parse_interface(pair: Pair<ast::Rule>) -> Result<Interface, String> {
    let mut inner = pair.into_inner();

    let name: String = inner.next().ok_or("expected iface name")?.as_str().into();

    let mut fields = vec![];
    let mut methods = vec![];

    for body_pair in inner.next().ok_or("expected iface body")?.into_inner() {
        match body_pair.as_rule() {
            ast::Rule::iface_field_defn => {
                let mut inner = body_pair.into_inner();

                let name = inner
                    .next()
                    .ok_or("expected iface field name")?
                    .as_str()
                    .to_owned();
                let (optional, field_type_pair) = {
                    let pair = inner.next().ok_or("expected optionality or type ident")?;

                    match pair.as_rule() {
                        ast::Rule::optional => (true, inner.next().ok_or("expected type ident")?),
                        ast::Rule::type_ident => (false, pair),
                        r @ _ => return Err(format!("unknown iface field part: {r:?}")),
                    }
                };

                fields.push(InterfaceField {
                    name,
                    optional,
                    typ: parse_type_ident(field_type_pair)?.into(),
                });
            }
            ast::Rule::iface_method_defn => {
                let mut inner = body_pair.into_inner();

                let name = inner.next().ok_or("expected iface method name")?.as_str();
                let fn_params = inner.next().ok_or("expected fn param list")?.into_inner();
            }
            r @ _ => return Err(format!("unknown iface inner construct: {r:?}")),
        }
    }

    Ok(Interface {
        name,
        fields,
        methods,
    })
}

fn parse_type_ident(pair: Pair<ast::Rule>) -> Result<TypeIdent, String> {
    match pair.as_rule() {
        ast::Rule::type_ident => {
            let mut inner = pair.into_inner();

            let head = parse_type_ident_type(inner.next().ok_or("expected type_ident head")?)?;

            let rest = {
                let mut parts = vec![];
                while let Some(pair) = inner.next() {
                    let mut inner = pair.into_inner();

                    let op_pair = inner.next().ok_or("expected type_ident_part_operator")?;
                    let typ =
                        parse_type_ident_type(inner.next().ok_or("expected type_ident_type")?)?;

                    parts.push(match op_pair.as_rule() {
                        ast::Rule::union => TypeIdentPart::Union(typ),
                        ast::Rule::sum => TypeIdentPart::Sum(typ),
                        r @ _ => return Err(format!("expected union or sum: {r:?}")),
                    });
                }

                parts
            };

            Ok(TypeIdent { head, rest })
        }
        r @ _ => return Err(format!("expected type_ident: {r:?}")),
    }
}

fn parse_type_ident_type(pair: Pair<ast::Rule>) -> Result<TypeIdentType, String> {
    Ok(match pair.as_rule() {
        ast::Rule::type_ident_type => {
            let inner = pair
                .into_inner()
                .next()
                .ok_or("expected literal_type or ident")?;

            match inner.as_rule() {
                ast::Rule::literal_type => {
                    let inner = inner
                        .into_inner()
                        .next()
                        .ok_or("expected literal_type part")?;

                    match inner.as_rule() {
                        ast::Rule::fn_type => {
                            let mut inner = inner.into_inner();

                            let params =
                                parse_fn_params(inner.next().ok_or("expected fn_param_list")?)?;

                            let return_type =
                                parse_type_ident(inner.next().ok_or("expected type_ident")?)?;

                            TypeIdentType::LiteralType(Box::new(LiteralType::FnType {
                                params,
                                return_type,
                            }))
                        }
                        ast::Rule::obj_type => {
                            let mut fields = vec![];

                            for pair in inner.into_inner() {
                                let mut inner = pair.into_inner();

                                let name = inner.next().ok_or("expected ident")?.as_str().into();
                                let (optional, type_pair) = {
                                    let pair =
                                        inner.next().ok_or("expected optional or type_ident")?;

                                    match pair.as_rule() {
                                        ast::Rule::optional => {
                                            (true, inner.next().ok_or("expected type_ident")?)
                                        }
                                        ast::Rule::type_ident => (false, pair),
                                        r @ _ => {
                                            return Err(format!(
                                                "unknown object_type_field part: {r:?}"
                                            ))
                                        }
                                    }
                                };

                                fields.push(ObjTypeField {
                                    name,
                                    optional,
                                    typ: parse_type_ident(type_pair)?,
                                })
                            }

                            TypeIdentType::LiteralType(Box::new(LiteralType::ObjType { fields }))
                        }
                        r @ _ => return Err(format!("expected fn_type or obj_type: {r:?}")),
                    }
                }
                ast::Rule::ident => TypeIdentType::Name(inner.as_str().into()),
                r @ _ => return Err(format!("expected literal_type or ident: {r:?}")),
            }
        }
        r @ _ => return Err(format!("expected type_ident_type: {r:?}")),
    })
}

fn parse_fn_params(pair: Pair<ast::Rule>) -> Result<Vec<FnParam>, String> {
    let mut fn_params = vec![];

    for pair in pair.into_inner() {
        let mut inner = pair.into_inner();

        let name = inner.next().ok_or("expected ident")?.as_str().into();
        let (optional, type_pair) = {
            let pair = inner.next().ok_or("expected optional or type_ident")?;

            match pair.as_rule() {
                ast::Rule::optional => (true, inner.next().ok_or("expected type_ident")?),
                ast::Rule::type_ident => (false, pair),
                r @ _ => return Err(format!("unknown fn_param part: {r:?}")),
            }
        };

        fn_params.push(FnParam {
            name,
            optional,
            typ: parse_type_ident(type_pair)?,
        });
    }

    return Ok(fn_params);
}

#[derive(Debug)]
enum TopLevelConstruct {
    Interface(Interface),
    StmtOrExpr(StmtOrExpr),
}

#[derive(Debug)]
struct Interface {
    name: String,
    fields: Vec<InterfaceField>,
    methods: Vec<InterfaceMethod>,
}

#[derive(Debug)]
struct InterfaceMethod {
    name: String,
    params: Vec<FnParam>,
}

#[derive(Debug)]
struct InterfaceField {
    name: String,
    optional: bool,
    typ: TypeIdent,
}

#[derive(Debug)]
struct FnParam {
    name: String,
    optional: bool,
    typ: TypeIdent,
}

#[derive(Debug)]
struct TypeIdent {
    head: TypeIdentType,
    rest: Vec<TypeIdentPart>,
}

#[derive(Debug)]
enum TypeIdentPart {
    Union(TypeIdentType),
    Sum(TypeIdentType),
}

#[derive(Debug)]
enum TypeIdentType {
    Name(String),
    LiteralType(Box<LiteralType>),
}

#[derive(Debug)]
enum LiteralType {
    FnType {
        params: Vec<FnParam>,
        return_type: TypeIdent,
    },
    ObjType {
        fields: Vec<ObjTypeField>,
    },
}

#[derive(Debug)]
struct ObjTypeField {
    name: String,
    optional: bool,
    typ: TypeIdent,
}

#[derive(Debug)]
enum StmtOrExpr {
    Stmt(Stmt),
    Expr(Expr),
}

#[derive(Debug)]
enum Stmt {}

#[derive(Debug)]
enum Expr {}
