use pest::{iterators::Pair, Parser};

use crate::ast;
use crate::ast::Rule;

mod types;
#[macro_use]
mod macros;
pub use types::*;

pub type ParseError = Box<dyn std::error::Error>;

#[derive(Debug, PartialEq)]
pub struct ParserResult {
    pub top_level_constructs: Vec<TopLevelConstruct>,
}

pub fn parse(src: &str) -> Result<ParserResult, ParseError> {
    let root_pairs = ast::TsParser::parse(Rule::ts, src)?;

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
    assert_rule!(pair, Rule::iface_defn);

    let mut inner = pair.into_inner();

    let name: String = parse_ident(inner.next().unwrap())?;
    let mut fields = vec![];
    let mut methods = vec![];

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
            Rule::iface_method_defn => {
                let mut inner = body_pair.into_inner();

                let name = parse_ident(inner.next().unwrap())?;
                let (params, typ) = {
                    let next = inner.next();

                    // See if there's anything left.
                    match next {
                        Some(next) => {
                            // If there is, see if it's a list of params or a return type.
                            match next.as_rule() {
                                Rule::fn_param_list => {
                                    // Parse the params.
                                    let params = parse_fn_param_list(next)?;

                                    // See if there's a type.
                                    let typ = match inner.next() {
                                        Some(next) => Some(parse_type_ident(next)?),
                                        None => None,
                                    };

                                    (params, typ)
                                }
                                Rule::type_ident => (vec![], Some(parse_type_ident(next)?)),
                                _ => todo!(),
                            }
                        }
                        None => (vec![], None),
                    }
                };

                methods.push(InterfaceMethod { name, params, typ });
            }
            _ => unreachable!(),
        }
    }

    Ok(Interface {
        name,
        fields,
        methods,
    })
}

fn parse_stmt_or_expr(pair: Pair<Rule>) -> Result<StmtOrExpr, ParseError> {
    assert_rule!(pair, Rule::stmt_or_expr);

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::stmt => StmtOrExpr::Stmt(parse_stmt(inner)?),
        Rule::expr => StmtOrExpr::Expr(parse_expr(inner)?),
        _ => unreachable!(),
    })
}

fn parse_expr(pair: Pair<Rule>) -> Result<Expr, ParseError> {
    fn parse_expr_inner(pair: Pair<Rule>) -> Result<ExprInner, ParseError> {
        let inner = pair.into_inner().next().unwrap();

        Ok(match inner.as_rule() {
            Rule::comparison => ExprInner::Comparison(parse_comparison(inner)?),
            Rule::incr_decr => ExprInner::IncrDecr(parse_incr_decr(inner)?),
            Rule::num => ExprInner::Num(parse_num(inner)?),
            Rule::string => ExprInner::Str(parse_str(inner)?),
            Rule::ident_assignment => {
                let mut inner = inner.into_inner();

                let ident = parse_ident(inner.next().unwrap())?;
                let assignment = parse_assignment(inner.next().unwrap())?;

                ExprInner::IdentAssignment(Box::new(IdentAssignment { ident, assignment }))
            }
            Rule::fn_inst => ExprInner::FnInst(parse_fn_inst(inner)?),
            Rule::obj_inst => {
                let mut fields = vec![];
                for pair in inner.into_inner() {
                    let mut inner = pair.into_inner();

                    let name = parse_ident(inner.next().unwrap())?;
                    let value = parse_expr(inner.next().unwrap())?;

                    fields.push(ObjFieldInst { name, value });
                }

                ExprInner::ObjInst(ObjInst { fields })
            }
            Rule::ident => ExprInner::Ident(parse_ident(inner)?),
            r @ _ => todo!("{:?}", r),
        })
    }

    let mut inner = pair.into_inner();
    let (is_sub_expr, expr_inner) = match inner.peek().unwrap().as_rule() {
        Rule::expr_inner => (false, parse_expr_inner(inner.next().unwrap())?),
        Rule::sub_expr => (
            true,
            parse_expr_inner(inner.next().unwrap().into_inner().next().unwrap())?,
        ),
        r @ _ => todo!("{:?}", r),
    };

    let ops = {
        let mut ops = vec![];

        while let Some(_) = inner.peek() {
            let next = inner.next().unwrap().into_inner().next().unwrap();
            ops.push(match next.as_rule() {
                Rule::obj_access => ObjOp::Access(parse_ident(next.into_inner().next().unwrap())?),
                Rule::obj_invoc => {
                    let mut args = vec![];

                    for pair in next.into_inner() {
                        args.push(parse_expr(pair)?);
                    }

                    ObjOp::Invoc { args }
                }
                Rule::arthm => ObjOp::Arithmetic(parse_arithmetic(next)?),
                Rule::comparison_op => ObjOp::ComparisonOp(parse_comparison_op(next)?),
                Rule::assignment => ObjOp::Assignment(parse_assignment(next)?),
                rule @ _ => todo!("{:?}", rule),
            });
        }

        ops
    };

    Ok(Expr {
        inner: expr_inner,
        is_sub_expr,
        ops,
    })
}

fn parse_comparison(pair: Pair<Rule>) -> Result<Comparison, ParseError> {
    assert_rule!(pair, Rule::comparison);

    let mut inner = pair.into_inner();

    let left = parse_comparison_term(inner.next().unwrap())?;
    let op = parse_comparison_op(inner.next().unwrap())?;
    let right = parse_comparison_term(inner.next().unwrap())?;

    Ok(Comparison { left, op, right })
}

fn parse_comparison_op(pair: Pair<Rule>) -> Result<ComparisonOp, ParseError> {
    assert_rule!(pair, Rule::comparison_op);

    Ok(match pair.as_str() {
        "==" => ComparisonOp::LooseEq,
        "!=" => ComparisonOp::LooseNeq,
        "<" => ComparisonOp::Lt,
        ">" => ComparisonOp::Gt,
        "&&" => ComparisonOp::And,
        op @ _ => todo!("{:?}", op),
    })
}

fn parse_comparison_term(pair: Pair<Rule>) -> Result<ComparisonTerm, ParseError> {
    assert_rule!(pair, Rule::comparison_term);

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::incr_decr => ComparisonTerm::IncrDecr(parse_incr_decr(inner)?),
        Rule::num => ComparisonTerm::Num(parse_num(inner)?),
        Rule::string => ComparisonTerm::Str(parse_str(inner)?),
        Rule::ident => ComparisonTerm::Ident(parse_ident(inner)?),
        Rule::comparison => ComparisonTerm::Comparison(Box::new(parse_comparison(inner)?)),
        Rule::arthm => ComparisonTerm::Arithmetic(parse_arithmetic(inner)?),
        inner @ _ => todo!("{:?}", inner),
    })
}

fn parse_arithmetic(pair: Pair<Rule>) -> Result<Arithmetic, ParseError> {
    assert_rule!(pair, Rule::arthm);

    let mut inner = pair.into_inner();

    let term = parse_arithmetic_term(inner.next().unwrap())?;

    let ops = {
        let mut ops = vec![];
        while let Some(op) = inner.next() {
            let op = match op.as_rule() {
                Rule::add => ArithmeticOp::Add,
                Rule::sub => ArithmeticOp::Sub,
                Rule::mult => ArithmeticOp::Mult,
                Rule::div => ArithmeticOp::Div,
                Rule::modu => ArithmeticOp::Modu,
                rule @ _ => todo!("{:?}", rule),
            };

            let term = parse_arithmetic_term(inner.next().unwrap())?;

            ops.push((op, term));
        }

        ops
    };

    Ok(Arithmetic { term, ops })
}

fn parse_arithmetic_term(pair: Pair<Rule>) -> Result<ArithmeticTerm, ParseError> {
    assert_rule!(pair, Rule::arthm_term);

    let term = pair.into_inner().next().unwrap();
    Ok(match term.as_rule() {
        Rule::ident => ArithmeticTerm::Ident(parse_ident(term)?),
        Rule::num => ArithmeticTerm::Num(parse_num(term)?),
        rule @ _ => todo!("{:?}", rule),
    })
}

fn parse_incr_decr(pair: Pair<Rule>) -> Result<IncrDecr, ParseError> {
    assert_rule!(pair, Rule::incr_decr);

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::increment => {
            let mut inner = inner.into_inner();

            let incr_type = inner.next().unwrap();
            let incr_type_rule = incr_type.as_rule();
            let target = parse_incr_decr_target(incr_type.into_inner().next().unwrap())?;

            match incr_type_rule {
                Rule::pre_incr => IncrDecr::Incr(Increment::Pre(target)),
                Rule::post_incr => IncrDecr::Incr(Increment::Post(target)),
                _ => todo!(),
            }
        }

        Rule::decrement => {
            let mut inner = inner.into_inner();

            let decr_type = inner.next().unwrap();
            let decr_type_rule = decr_type.as_rule();
            let target = parse_incr_decr_target(decr_type.into_inner().next().unwrap())?;

            match decr_type_rule {
                Rule::pre_incr => IncrDecr::Decr(Decrement::Pre(target)),
                Rule::post_incr => IncrDecr::Decr(Decrement::Post(target)),
                _ => todo!(),
            }
        }

        _ => todo!(),
    })
}

fn parse_incr_decr_target(pair: Pair<Rule>) -> Result<IncrDecrTarget, ParseError> {
    assert_rule!(pair, Rule::incr_target);

    let inner = pair.into_inner().next().unwrap();

    Ok(match inner.as_rule() {
        Rule::ident => IncrDecrTarget::Ident(parse_ident(inner)?),
        _ => todo!(),
    })
}

fn parse_num(pair: Pair<Rule>) -> Result<f32, ParseError> {
    assert_rule!(pair, Rule::num);

    Ok(pair.as_str().parse::<f32>().map_err(|e| format!("{e}"))?)
}

fn parse_str(pair: Pair<Rule>) -> Result<String, ParseError> {
    assert_rule!(pair, Rule::string);

    Ok(pair.into_inner().next().unwrap().as_str().into())
}

fn parse_ident(pair: Pair<Rule>) -> Result<String, ParseError> {
    assert_rule!(pair, Rule::ident);

    Ok(pair.as_str().into())
}

fn parse_fn_inst(pair: Pair<Rule>) -> Result<FnInst, ParseError> {
    assert_rule!(pair, Rule::fn_inst);

    let (mut name, mut params, mut return_type, mut body) = (None, None, None, vec![]);
    for inner in pair.into_inner() {
        match inner.as_rule() {
            Rule::ident => name = Some(parse_ident(inner)?),
            Rule::fn_param_list => params = Some(parse_fn_param_list(inner)?),
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
    assert_rule!(pair, Rule::stmt);

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::for_loop => {
            let mut inner = inner.into_inner();

            let init = parse_let_decl(inner.next().unwrap())?;
            let condition = parse_expr(inner.next().unwrap())?;
            let after = parse_expr(inner.next().unwrap())?;

            let mut body = vec![];
            while let Some(pair) = inner.next() {
                body.push(parse_stmt_or_expr(pair)?);
            }

            Stmt::ForLoop {
                init: Box::new(init),
                condition,
                after,
                body,
            }
        }
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
        Rule::if_stmt => Stmt::If(parse_if_stmt(inner)?),
        rule @ _ => todo!("{:?}", rule),
    })
}

fn parse_if_stmt(pair: Pair<Rule>) -> Result<IfStmt, ParseError> {
    assert_rule!(pair, Rule::if_stmt);

    let mut inner = pair.into_inner();

    let condition = parse_expr(inner.next().unwrap())?;
    let body = {
        let mut body = vec![];

        while let Some(pair) = inner.peek() {
            match pair.as_rule() {
                Rule::stmt_or_expr => body.push(parse_stmt_or_expr(inner.next().unwrap())?),
                Rule::else_if_stmt => break,
                rule @ _ => todo!("{:?}", rule),
            }
        }

        body
    };

    let else_ifs = {
        let mut else_ifs = vec![];

        while let Some(pair) = inner.peek() {
            match pair.as_rule() {
                Rule::else_if_stmt => {
                    let mut inner = inner.next().unwrap().into_inner();

                    let condition = parse_expr(inner.next().unwrap())?;
                    let body = {
                        let mut body = vec![];

                        for pair in inner {
                            body.push(parse_stmt_or_expr(pair)?);
                        }

                        body
                    };

                    else_ifs.push(ElseIfStmt { condition, body })
                }
                Rule::else_stmt => break,
                rule @ _ => todo!("{:?}", rule),
            }
        }

        else_ifs
    };

    let els = match inner.next() {
        Some(pair) => {
            let inner = pair.into_inner();

            let mut body = vec![];
            for pair in inner {
                body.push(parse_stmt_or_expr(pair)?);
            }

            Some(ElseStmt { body })
        }
        None => None,
    };

    Ok(IfStmt {
        condition,
        body,
        else_ifs,
        els,
    })
}

fn parse_let_decl(pair: Pair<Rule>) -> Result<Stmt, ParseError> {
    assert_rule!(pair, Rule::let_decl);

    let mut inner = pair.into_inner();

    let (name, mut typ, mut assignment) = (parse_ident(inner.next().unwrap())?, None, None);
    for pair in inner {
        match pair.as_rule() {
            Rule::type_ident => typ = Some(parse_type_ident(pair)?),
            Rule::assignment => assignment = Some(parse_assignment(pair)?),
            _ => unreachable!(),
        }
    }

    Ok(Stmt::LetDecl {
        name,
        typ,
        assignment,
    })
}

fn parse_assignment(pair: Pair<Rule>) -> Result<Expr, ParseError> {
    assert_rule!(pair, Rule::assignment);

    Ok(parse_expr(pair.into_inner().next().unwrap())?)
}

fn parse_type_ident(pair: Pair<Rule>) -> Result<TypeIdent, ParseError> {
    assert_rule!(pair, Rule::type_ident);

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
    assert_rule!(pair, Rule::type_ident_type);

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::literal_type => TypeIdentType::LiteralType(Box::new(parse_literal_type(inner)?)),
        Rule::ident => TypeIdentType::Name(parse_ident(inner)?),
        _ => unreachable!(),
    })
}

fn parse_literal_type(pair: Pair<Rule>) -> Result<LiteralType, ParseError> {
    assert_rule!(pair, Rule::literal_type);

    let inner = pair.into_inner().next().unwrap();
    Ok(match inner.as_rule() {
        Rule::fn_type => {
            let mut inner = inner.into_inner();

            let params = parse_fn_param_list(inner.next().unwrap())?;
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

fn parse_fn_param_list(pair: Pair<Rule>) -> Result<Vec<FnParam>, ParseError> {
    assert_rule!(pair, Rule::fn_param_list);

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
