use pest::{iterators::Pair, Parser};

use crate::ast;
use crate::ast::Rule;

mod types;
pub use types::*;

pub type ParseError = String;

#[derive(Debug, PartialEq)]
pub struct ParserResult {
    pub top_level_constructs: Vec<TopLevelConstruct>,
}

pub fn parse(src: &str) -> Result<ParserResult, ParseError> {
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

    Ok(ParserResult {
        top_level_constructs,
    })
}

fn parse_interface(pair: Pair<Rule>) -> Result<Interface, ParseError> {
    assert_rule(&pair, Rule::iface_defn)?;

    let mut inner = pair.into_inner();

    let name: String = parse_ident(inner.next().unwrap())?;
    let mut fields = vec![];

    for body_pair in inner.next().unwrap().into_inner() {
        match body_pair.as_rule() {
            Rule::iface_field_defn => {
                let mut inner = body_pair.into_inner();

                let name = parse_ident(inner.next().unwrap())?;
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

fn parse_stmt_or_expr(pair: Pair<Rule>) -> Result<StmtOrExpr, ParseError> {
    assert_rule(&pair, Rule::stmt_or_expr)?;

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::stmt => StmtOrExpr::Stmt(parse_stmt(inner)?),
        Rule::expr => StmtOrExpr::Expr(parse_expr(inner)?),
        _ => unreachable!(),
    })
}

fn parse_expr(pair: Pair<Rule>) -> Result<Expr, ParseError> {
    assert_rule(&pair, Rule::expr)?;

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::num => Expr::Num(inner.as_str().parse::<f32>().map_err(|e| format!("{e}"))?),
        Rule::string => Expr::Str(inner.into_inner().next().unwrap().as_str().into()),
        Rule::ident_assignment => {
            let mut inner = inner.into_inner();

            let ident = parse_ident(inner.next().unwrap())?;
            let assignment = parse_assignment(inner.next().unwrap())?;

            Expr::IdentAssignment(Box::new(IdentAssignment { ident, assignment }))
        }
        Rule::fn_inst => Expr::FnInst(parse_fn_inst(inner)?),
        Rule::chained_obj_op => {
            let mut inner = inner.into_inner();

            let accessable = {
                let next = inner.next().unwrap();
                match next.as_rule() {
                    Rule::ident => Accessable::Ident(parse_ident(next)?),
                    Rule::literal_type => Accessable::LiteralType(parse_literal_type(next)?),
                    _ => unreachable!(),
                }
            };

            let obj_ops = {
                let mut ops = vec![];

                while let Some(peeked) = inner.peek() {
                    if peeked.as_rule() != Rule::obj_op {
                        break;
                    }

                    let next = inner.next().unwrap().into_inner().next().unwrap();
                    ops.push(match next.as_rule() {
                        Rule::obj_access => {
                            ObjOp::Access(parse_ident(next.into_inner().next().unwrap())?)
                        }
                        Rule::obj_invoc => {
                            let mut args = vec![];

                            for pair in next.into_inner() {
                                args.push(parse_expr(pair)?);
                            }

                            ObjOp::Invoc { args }
                        }
                        _ => unreachable!(),
                    });
                }

                ops
            };

            let assignment = {
                match inner.next() {
                    Some(next) => Some(Box::new(parse_expr(next.into_inner().next().unwrap())?)),
                    None => None,
                }
            };

            Expr::ChainedObjOp(ChainedObjOp {
                accessable,
                obj_ops,
                assignment,
            })
        }
        Rule::obj_inst => {
            let mut fields = vec![];
            for pair in inner.into_inner() {
                let mut inner = pair.into_inner();

                let name = parse_ident(inner.next().unwrap())?;
                let value = parse_expr(inner.next().unwrap())?;

                fields.push(ObjFieldInst { name, value });
            }

            Expr::ObjInst(ObjInst { fields })
        }
        Rule::ident => Expr::Ident(parse_ident(inner)?),
        _ => unreachable!(),
    })
}

fn parse_ident(pair: Pair<Rule>) -> Result<String, String> {
    assert_rule(&pair, Rule::ident)?;

    Ok(pair.as_str().into())
}

fn parse_fn_inst(pair: Pair<Rule>) -> Result<FnInst, ParseError> {
    assert_rule(&pair, Rule::fn_inst)?;

    let (mut name, mut params, mut return_type, mut body) = (None, None, None, vec![]);
    for inner in pair.into_inner() {
        match inner.as_rule() {
            Rule::ident => name = Some(parse_ident(inner)?),
            Rule::fn_param_list => params = Some(parse_fn_params(inner)?),
            Rule::type_ident => return_type = Some(parse_type_ident(inner)?),
            Rule::stmt_or_expr => body.push(parse_stmt_or_expr(inner)?),
            _ => unreachable!(),
        }
    }

    Ok(FnInst {
        name,
        params: params.unwrap_or(vec![]),
        return_type,
        body,
    })
}

fn parse_stmt(pair: Pair<Rule>) -> Result<Stmt, ParseError> {
    assert_rule(&pair, Rule::stmt)?;

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::let_decl => {
            let mut inner = inner.into_inner();

            let (name, mut typ, mut assignment) = (parse_ident(inner.next().unwrap())?, None, None);
            for pair in inner {
                match pair.as_rule() {
                    Rule::type_ident => typ = Some(parse_type_ident(pair)?),
                    Rule::assignment => assignment = Some(parse_assignment(pair)?),
                    _ => unreachable!(),
                }
            }

            Stmt::LetDecl {
                name,
                typ,
                assignment,
            }
        }
        Rule::fn_inst => todo!("fn_inst"),
        Rule::expr => Stmt::Expr(parse_expr(inner)?),
        Rule::return_expr => Stmt::ReturnExpr(parse_expr(inner.into_inner().next().unwrap())?),
        _ => unreachable!(),
    })
}

fn parse_assignment(pair: Pair<Rule>) -> Result<Expr, String> {
    assert_rule(&pair, Rule::assignment)?;

    Ok(parse_expr(pair.into_inner().next().unwrap())?)
}

fn parse_type_ident(pair: Pair<Rule>) -> Result<TypeIdent, ParseError> {
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

    Ok(TypeIdent {
        head,
        rest: if rest.is_empty() { None } else { Some(rest) },
    })
}

fn parse_type_ident_type(pair: Pair<Rule>) -> Result<TypeIdentType, ParseError> {
    assert_rule(&pair, Rule::type_ident_type)?;

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::literal_type => TypeIdentType::LiteralType(Box::new(parse_literal_type(inner)?)),
        Rule::ident => TypeIdentType::Name(parse_ident(inner)?),
        _ => unreachable!(),
    })
}

fn parse_literal_type(pair: Pair<Rule>) -> Result<LiteralType, ParseError> {
    assert_rule(&pair, Rule::literal_type)?;

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::fn_type => {
            let mut inner = inner.into_inner();

            let params = parse_fn_params(inner.next().unwrap())?;
            let return_type = parse_type_ident(inner.next().unwrap())?;

            LiteralType::FnType {
                params,
                return_type: Some(return_type),
            }
        }
        Rule::obj_type => {
            let mut fields = vec![];

            // Parse each obj_type_field.
            for pair in inner.into_inner() {
                let mut inner = pair.into_inner();

                let name = parse_ident(inner.next().unwrap())?;
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

            LiteralType::ObjType { fields }
        }
        _ => unreachable!(),
    })
}

fn parse_fn_params(pair: Pair<Rule>) -> Result<Vec<FnParam>, ParseError> {
    assert_rule(&pair, Rule::fn_param_list)?;

    let mut fn_params = vec![];

    for pair in pair.into_inner() {
        let mut inner = pair.into_inner();

        let name = parse_ident(inner.next().unwrap())?;
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

fn assert_rule(pair: &Pair<Rule>, expected: Rule) -> Result<(), ParseError> {
    let rule = pair.as_rule();
    if rule == expected {
        Ok(())
    } else {
        panic!("expected {expected:?}, got: {rule:?}")
    }
}
