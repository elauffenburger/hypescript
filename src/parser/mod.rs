use pest::{iterators::Pair, Parser};

use crate::ast;
use crate::ast::Rule;

mod types;
pub use types::*;

pub struct ParseResult {
    pub top_level_constructs: Vec<TopLevelConstruct>,
}

pub fn parse(src: &str) -> Result<ParseResult, String> {
    let root_pairs = ast::TsParser::parse(Rule::ts, src).map_err(|e| format!("{e}"))?;

    let mut top_level_constructs = vec![];
    for pair in root_pairs {
        if let Rule::EOI = pair.as_rule() {
            break;
        }

        top_level_constructs.push(match pair.as_rule() {
            Rule::iface_defn => TopLevelConstruct::Interface(parse_interface(pair)?),
            Rule::stmt_or_expr => TopLevelConstruct::StmtOrExpr(parse_stmt_or_expr(pair)?),
            _ => unreachable!(),
        })
    }

    Ok(ParseResult {
        top_level_constructs,
    })
}

fn parse_interface(pair: Pair<Rule>) -> Result<Interface, String> {
    assert_rule(&pair, Rule::iface_defn)?;

    let mut inner = pair.into_inner();

    let name: String = inner.next().unwrap().as_str().into();
    let mut fields = vec![];

    for body_pair in inner.next().unwrap().into_inner() {
        match body_pair.as_rule() {
            Rule::iface_field_defn => {
                let mut inner = body_pair.into_inner();

                let name = inner.next().unwrap().as_str().into();
                let (optional, field_type_pair) = {
                    let pair = inner.next().unwrap();

                    match pair.as_rule() {
                        Rule::optional => (true, inner.next().unwrap()),
                        Rule::type_ident => (false, pair),
                        _ => unreachable!(),
                    }
                };

                fields.push(InterfaceField {
                    name,
                    optional,
                    typ: parse_type_ident(field_type_pair)?,
                });
            }
            Rule::iface_method_defn => todo!("iface_method_defn"),
            _ => unreachable!(),
        }
    }

    Ok(Interface {
        name,
        fields,
        methods: vec![],
    })
}

fn parse_stmt_or_expr(pair: Pair<Rule>) -> Result<StmtOrExpr, String> {
    assert_rule(&pair, Rule::stmt_or_expr)?;

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::stmt => StmtOrExpr::Stmt(parse_stmt(inner)?),
        Rule::expr => StmtOrExpr::Expr(parse_expr(inner)?),
        _ => unreachable!(),
    })
}

fn parse_expr(pair: Pair<Rule>) -> Result<Expr, String> {
    assert_rule(&pair, Rule::expr)?;

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::num => Expr::Num(inner.as_str().parse::<f32>().map_err(|e| format!("{e}"))?),
        Rule::string => todo!("string"),
        Rule::ident_assignment => todo!("ident_assignment"),
        Rule::fn_inst => todo!("fn_inst"),
        Rule::chained_obj_op => todo!("chained_obj_op"),
        Rule::obj_inst => todo!("obj_inst"),
        Rule::ident => todo!("ident"),
        _ => unreachable!(),
    })
}

fn parse_stmt(pair: Pair<Rule>) -> Result<Stmt, String> {
    assert_rule(&pair, Rule::stmt)?;

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::let_decl => todo!("let_decl"),
        Rule::fn_inst => todo!("fn_inst"),
        Rule::expr => todo!("expr"),
        Rule::return_expr => todo!("return_expr"),
        _ => unreachable!(),
    })
}

fn parse_type_ident(pair: Pair<Rule>) -> Result<TypeIdent, String> {
    assert_rule(&pair, Rule::type_ident)?;

    let mut inner = pair.into_inner();

    let head = parse_type_ident_type(inner.next().unwrap())?;
    let rest = {
        let mut parts = vec![];
        while let Some(pair) = inner.next() {
            let mut inner = pair.into_inner();

            let op_pair = inner.next().unwrap();
            let typ = parse_type_ident_type(inner.next().unwrap())?;

            parts.push(match op_pair.as_rule() {
                Rule::union => TypeIdentPart::Union(typ),
                Rule::sum => TypeIdentPart::Sum(typ),
                _ => unreachable!(),
            });
        }

        parts
    };

    Ok(TypeIdent { head, rest })
}

fn parse_type_ident_type(pair: Pair<Rule>) -> Result<TypeIdentType, String> {
    assert_rule(&pair, Rule::type_ident_type)?;

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::literal_type => parse_literal_type(inner)?,
        Rule::ident => TypeIdentType::Name(inner.as_str().into()),
        _ => unreachable!(),
    })
}

fn parse_literal_type(pair: Pair<Rule>) -> Result<TypeIdentType, String> {
    assert_rule(&pair, Rule::literal_type)?;

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::fn_type => {
            let mut inner = inner.into_inner();

            let params = parse_fn_params(inner.next().unwrap())?;
            let return_type = parse_type_ident(inner.next().unwrap())?;

            TypeIdentType::LiteralType(Box::new(LiteralType::FnType {
                params,
                return_type,
            }))
        }
        Rule::obj_type => parse_obj_type(inner)?,
        _ => unreachable!(),
    })
}

fn parse_obj_type(pair: Pair<Rule>) -> Result<TypeIdentType, String> {
    assert_rule(&pair, Rule::obj_type)?;

    let mut fields = vec![];

    // Parse each obj_type_field.
    for pair in pair.into_inner() {
        let mut inner = pair.into_inner();

        let name = inner.next().unwrap().as_str().into();
        let (optional, type_pair) = {
            let pair = inner.next().unwrap();

            match pair.as_rule() {
                Rule::optional => (true, inner.next().unwrap()),
                Rule::type_ident => (false, pair),
                _ => unreachable!(),
            }
        };

        fields.push(ObjTypeField {
            name,
            optional,
            typ: parse_type_ident(type_pair)?,
        })
    }

    Ok(TypeIdentType::LiteralType(Box::new(LiteralType::ObjType {
        fields,
    })))
}

fn parse_fn_params(pair: Pair<Rule>) -> Result<Vec<FnParam>, String> {
    assert_rule(&pair, Rule::fn_param_list)?;

    let mut fn_params = vec![];

    for pair in pair.into_inner() {
        let mut inner = pair.into_inner();

        let name = inner.next().unwrap().as_str().into();
        let (optional, typ) = {
            let opt_pair = inner.next();
            let type_pair = inner.next();

            match (opt_pair, type_pair) {
                // No optionality, no type:
                (None, None) => (false, None),

                // Optionality or type:
                (Some(next), None) => match next.as_rule() {
                    Rule::optional => (true, None),
                    Rule::type_ident => (false, Some(parse_type_ident(next)?)),
                    _ => unreachable!(),
                },

                // Optionality and type:
                (Some(_), Some(type_pair)) => (true, Some(parse_type_ident(type_pair)?)),
                _ => unreachable!(),
            }
        };

        fn_params.push(FnParam {
            name,
            optional,
            typ,
        });
    }

    return Ok(fn_params);
}

fn assert_rule(pair: &Pair<Rule>, expected: Rule) -> Result<(), String> {
    let rule = pair.as_rule();
    if rule == expected {
        Ok(())
    } else {
        Err(format!("expected {expected:?}, got: {rule:?}"))
    }
}
