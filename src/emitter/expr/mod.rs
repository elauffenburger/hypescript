use crate::{
    parser::{self, ObjOp},
    util::rcref,
};

use super::{EmitResult, Emitter};

mod fn_inst;
pub use fn_inst::*;

mod obj;
pub use obj::*;

impl Emitter {
    pub(in crate::emitter) fn emit_expr(&mut self, expr: parser::Expr) -> EmitResult {
        if expr.is_sub_expr {
            self.emit_sub_expr(expr.clone())?;
        } else {
            match expr.clone().inner {
                parser::ExprInner::Comparison(comp) => self.emit_comparison(comp),
                parser::ExprInner::IncrDecr(incr_decr) => self.emit_incr_decr(incr_decr),
                parser::ExprInner::Num(num) => self.emit_num(num),
                parser::ExprInner::Str(str) => self.emit_str(&str),
                parser::ExprInner::IdentAssignment(ident_assignment) => {
                    self.emit_ident_assignment(*ident_assignment)
                }
                parser::ExprInner::FnInst(fn_inst) => self.emit_fn_inst(fn_inst),
                parser::ExprInner::ObjInst(obj_inst) => self.emit_obj_inst(obj_inst),
                parser::ExprInner::Ident(ident) => self.emit_ident(&ident),
            }?;
        }

        self.emit_expr_ops(expr)?;

        Ok(())
    }

    fn emit_expr_ops(&mut self, expr: parser::Expr) -> EmitResult {
        let mut curr_acc_type = rcref(self.type_of_expr_inner(&expr.inner)?);

        let mut last_op = None;
        let n = expr.ops.len();
        let has_assignment = {
            if expr.ops.is_empty() {
                false
            } else {
                expr.ops.get(n - 1).map_or(false, |op| {
                    if let ObjOp::Assignment(_) = op {
                        true
                    } else {
                        false
                    }
                })
            }
        };

        for (i, op) in expr.ops.iter().enumerate() {
            match op.clone() {
                parser::ObjOp::Access(prop) => {
                    // If this is the last op and there's an assignment we need to do,
                    // don't emit a `getFieldValue`; just skip it and write the set in a sec.
                    if has_assignment && i == n - 2 {
                        last_op = Some(op);

                        continue;
                    }

                    self.emit_get_field_val(&prop)?;

                    if let Some(_) = (*curr_acc_type.borrow()).rest {
                        todo!("complex types")
                    }

                    let typ = (*curr_acc_type.borrow()).head.clone();
                    curr_acc_type = match typ {
                        parser::TypeIdentType::Name(ref typ_name) => self
                            .get_type(typ_name)
                            .ok_or(format!("could not find type {typ_name}"))?,
                        parser::TypeIdentType::LiteralType(typ) => match *typ {
                            parser::LiteralType::FnType { .. } => todo!(),
                            parser::LiteralType::ObjType { fields } => {
                                match fields.iter().find(|field| field.name == prop) {
                                    Some(field) => rcref(field.typ.clone()),
                                    None => {
                                        return Err(format!("unknown field '{prop}' on target"))
                                    }
                                }
                            }
                        },
                        parser::TypeIdentType::Interface(_) => todo!(),
                    };
                }
                parser::ObjOp::Invoc { args } => {
                    self.write("->invoke({")?;

                    let n = args.len();
                    for (i, arg) in args.into_iter().enumerate() {
                        self.write(&format!("TsFunctionArg(\"{}\", ", "arg_name"))?;
                        self.emit_expr(arg)?;
                        self.write(")")?;

                        if i != n - 1 {
                            self.write(", ")?;
                        }
                    }

                    self.write("})")?;
                }
                parser::ObjOp::Arithmetic(artm) => self.emit_arithmetic(artm)?,
                parser::ObjOp::ComparisonOp(cmp) => self.emit_comparison_op(cmp)?,
                parser::ObjOp::Assignment(asgn) => {
                    let name = match last_op {
                        Some(op) => match op {
                            parser::ObjOp::Access(name) => name,
                            parser::ObjOp::Invoc { .. } => todo!(),
                            parser::ObjOp::Arithmetic(_) => todo!(),
                            parser::ObjOp::ComparisonOp(_) => todo!(),
                            parser::ObjOp::Assignment(_) => {
                                unreachable!("can't assign to an assignment")
                            }
                        },
                        None => unreachable!("assignment without access is impossible"),
                    };

                    self.write(&format!("->setFieldValue(\"{name}\", "))?;
                    self.emit_expr(asgn)?;
                    self.write(")")?;
                }
            }

            last_op = Some(op)
        }

        Ok(())
    }

    fn emit_sub_expr(&mut self, expr: parser::Expr) -> EmitResult {
        self.write("([=] {\n")?;
        self.write("auto _result = ")?;
        self.emit_expr(expr)?;
        self.write(";\n")?;
        self.write("return _result;\n")?;
        self.write("})()\n")?;

        Ok(())
    }

    fn emit_comparison_op(&mut self, op: parser::ComparisonOp) -> EmitResult {
        self.emit_get_field_val(match op {
            parser::ComparisonOp::LooseEq => "==",
            parser::ComparisonOp::LooseNeq => "!=",
            parser::ComparisonOp::Lt => "<",
            parser::ComparisonOp::Gt => ">",
            parser::ComparisonOp::And => "&&",
        })
    }

    fn emit_comparison(&mut self, comp: parser::Comparison) -> EmitResult {
        self.emit_comparison_term(comp.left)?;

        self.emit_comparison_op(comp.op)?;

        self.write("->invoke({")?;
        self.write("TsFunctionArg(\"other\", ")?;
        self.emit_comparison_term(comp.right)?;
        self.write(")")?;
        self.write("})")?;

        Ok(())
    }

    fn emit_comparison_term(&mut self, term: parser::ComparisonTerm) -> EmitResult {
        match term {
            parser::ComparisonTerm::IncrDecr(incr_decr) => self.emit_incr_decr(incr_decr),
            parser::ComparisonTerm::Num(num) => self.emit_num(num),
            parser::ComparisonTerm::Str(str) => self.emit_str(&str),
            parser::ComparisonTerm::IdentAssignment(ident_assign) => {
                self.emit_ident_assignment(*ident_assign)
            }
            parser::ComparisonTerm::Ident(ident) => self.emit_ident(&ident),
            parser::ComparisonTerm::Comparison(comp) => self.emit_comparison(*comp),
            parser::ComparisonTerm::Arithmetic(arthm) => self.emit_arithmetic(arthm),
        }
    }

    fn emit_get_field_val(&mut self, field: &str) -> EmitResult {
        self.write(&format!("->getFieldValue(\"{field}\")"))
    }

    fn emit_arithmetic(&mut self, arthm: parser::Arithmetic) -> EmitResult {
        for op in arthm.ops {
            self.emit_get_field_val(match op.0 {
                parser::ArithmeticOp::Add => "+",
                parser::ArithmeticOp::Sub => "-",
                parser::ArithmeticOp::Mult => "*",
                parser::ArithmeticOp::Div => "/",
                parser::ArithmeticOp::Modu => "%",
            })?;

            self.write("->invoke({")?;
            self.write("TsFunctionArg(\"other\", ")?;
            self.emit_arithmetic_term(op.1)?;
            self.write(")")?;
            self.write("})")?;
        }

        Ok(())
    }

    fn emit_arithmetic_term(&mut self, term: parser::ArithmeticTerm) -> EmitResult {
        match term {
            parser::ArithmeticTerm::Ident(ident) => self.emit_ident(&ident),
            parser::ArithmeticTerm::Num(num) => self.emit_num(num),
        }
    }

    fn emit_num(&mut self, num: f32) -> EmitResult {
        self.write(&format!("new TsNum({num})"))
    }

    fn emit_str(&mut self, str: &str) -> EmitResult {
        self.write(&format!("new TsString(\"{str}\")"))
    }

    fn emit_ident_assignment(&mut self, assignment: parser::IdentAssignment) -> EmitResult {
        self.emit_ident(&assignment.ident)?;
        self.write(" = ")?;
        self.emit_expr(assignment.assignment)?;

        Ok(())
    }

    fn emit_incr_decr(&mut self, incr_decr: parser::IncrDecr) -> EmitResult {
        let (target, fn_name) = match incr_decr {
            parser::IncrDecr::Incr(incr) => match incr {
                parser::Increment::Pre(tgt) => match tgt {
                    parser::IncrDecrTarget::Ident(_) => {
                        todo!()
                    }
                },
                parser::Increment::Post(tgt) => match tgt {
                    parser::IncrDecrTarget::Ident(ident) => (self.mangle_ident(&ident), "++"),
                },
            },
            parser::IncrDecr::Decr(decr) => match decr {
                parser::Decrement::Pre(tgt) => match tgt {
                    parser::IncrDecrTarget::Ident(_) => {
                        todo!()
                    }
                },
                parser::Decrement::Post(tgt) => match tgt {
                    parser::IncrDecrTarget::Ident(_) => {
                        todo!()
                    }
                },
            },
        };

        self.write(&target)?;
        self.emit_get_field_val(fn_name)?;

        self.write("->invoke({})")?;

        Ok(())
    }

    fn emit_ident(&mut self, ident: &str) -> EmitResult {
        self.write(&self.mangle_ident(ident))
    }
}
