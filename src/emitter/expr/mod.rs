use crate::util::rcref;
use super::*;

mod fn_inst;
pub use fn_inst::*;

mod obj;
pub use obj::*;

impl Emitter {
    pub(in crate::emitter) fn emit_expr(&mut self, expr: super::Expr) -> EmitResult {
        if expr.is_sub_expr {
            let mut expr = expr.clone();
            expr.is_sub_expr = false;

            return self.emit_sub_expr(expr);
        } else {
            match expr.clone().inner {
                super::ExprInner::Comparison(comp) => self.emit_comparison(comp),
                super::ExprInner::IncrDecr(incr_decr) => self.emit_incr_decr(incr_decr),
                super::ExprInner::Num(num) => self.emit_num(num),
                super::ExprInner::Str(str) => self.emit_str(&str),
                super::ExprInner::IdentAssignment(ident_assignment) => {
                    self.emit_ident_assignment(*ident_assignment)
                }
                super::ExprInner::FnInst(fn_inst) => self.emit_fn_inst(fn_inst),
                super::ExprInner::ObjInst(obj_inst) => self.emit_obj_inst(obj_inst),
                super::ExprInner::Ident(ident) => self.emit_ident(&ident),
            }?;
        }

        self.emit_expr_ops(expr)?;

        Ok(())
    }

    fn emit_expr_ops(&mut self, expr: super::Expr) -> EmitResult {
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
                super::ObjOp::Access(prop) => {
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
                        super::TypeIdentType::Name(ref type_ref) => self
                            .get_type(type_ref)
                            .ok_or(format!("could not find type {type_ref:?}"))?,
                        super::TypeIdentType::LiteralType(typ) => match *typ {
                            super::LiteralType::FnType { .. } => todo!(),
                            super::LiteralType::ObjType { fields } => {
                                match fields.iter().find(|field| field.name == prop) {
                                    Some(field) => rcref(field.typ.clone()),
                                    None => {
                                        return Err(format!("unknown field '{prop}' on target"))
                                    }
                                }
                            }
                        },
                        super::TypeIdentType::Interface(_) => todo!(),
                    };
                }
                super::ObjOp::Invoc { args } => {
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
                super::ObjOp::Arithmetic(artm) => self.emit_arithmetic(artm)?,
                super::ObjOp::ComparisonOp(cmp) => self.emit_comparison_op(cmp)?,
                super::ObjOp::Assignment(asgn) => {
                    let name = match last_op {
                        Some(op) => match op {
                            super::ObjOp::Access(name) => name,
                            super::ObjOp::Invoc { .. } => todo!(),
                            super::ObjOp::Arithmetic(_) => todo!(),
                            super::ObjOp::ComparisonOp(_) => todo!(),
                            super::ObjOp::Assignment(_) => {
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

    fn emit_sub_expr(&mut self, expr: super::Expr) -> EmitResult {
        self.write("([=] {\n")?;
        self.write("auto _result = ")?;
        self.emit_expr(expr)?;
        self.write(";\n")?;
        self.write("return _result;\n")?;
        self.write("})()\n")?;

        Ok(())
    }

    fn emit_comparison_op(&mut self, op: super::ComparisonOp) -> EmitResult {
        self.emit_get_field_val(match op {
            super::ComparisonOp::LooseEq => "==",
            super::ComparisonOp::LooseNeq => "!=",
            super::ComparisonOp::Lt => "<",
            super::ComparisonOp::Gt => ">",
            super::ComparisonOp::And => "&&",
        })
    }

    fn emit_comparison(&mut self, comp: super::Comparison) -> EmitResult {
        self.emit_comparison_term(comp.left)?;

        self.emit_comparison_op(comp.op)?;

        self.write("->invoke({")?;
        self.write("TsFunctionArg(\"other\", ")?;
        self.emit_comparison_term(comp.right)?;
        self.write(")")?;
        self.write("})")?;

        Ok(())
    }

    fn emit_comparison_term(&mut self, term: super::ComparisonTerm) -> EmitResult {
        match term {
            super::ComparisonTerm::IncrDecr(incr_decr) => self.emit_incr_decr(incr_decr),
            super::ComparisonTerm::Num(num) => self.emit_num(num),
            super::ComparisonTerm::Str(str) => self.emit_str(&str),
            super::ComparisonTerm::IdentAssignment(ident_assign) => {
                self.emit_ident_assignment(*ident_assign)
            }
            super::ComparisonTerm::Ident(ident) => self.emit_ident(&ident),
            super::ComparisonTerm::Comparison(comp) => self.emit_comparison(*comp),
            super::ComparisonTerm::Arithmetic(arthm) => self.emit_arithmetic(arthm),
        }
    }

    fn emit_get_field_val(&mut self, field: &str) -> EmitResult {
        self.write(&format!("->getFieldValue(\"{field}\")"))
    }

    fn emit_arithmetic(&mut self, arthm: super::Arithmetic) -> EmitResult {
        self.emit_arithmetic_term(arthm.term)?;

        for op in arthm.ops {
            self.emit_get_field_val(match op.0 {
                super::ArithmeticOp::Add => "+",
                super::ArithmeticOp::Sub => "-",
                super::ArithmeticOp::Mult => "*",
                super::ArithmeticOp::Div => "/",
                super::ArithmeticOp::Modu => "%",
            })?;

            self.write("->invoke({")?;
            self.write("TsFunctionArg(\"other\", ")?;
            self.emit_arithmetic_term(op.1)?;
            self.write(")")?;
            self.write("})")?;
        }

        Ok(())
    }

    fn emit_arithmetic_term(&mut self, term: super::ArithmeticTerm) -> EmitResult {
        match term {
            super::ArithmeticTerm::Ident(ident) => self.emit_ident(&ident),
            super::ArithmeticTerm::Num(num) => self.emit_num(num),
        }
    }

    fn emit_num(&mut self, num: f32) -> EmitResult {
        self.write(&format!("new TsNum({num})"))
    }

    fn emit_str(&mut self, str: &str) -> EmitResult {
        self.write(&format!("new TsString(\"{str}\")"))
    }

    fn emit_ident_assignment(&mut self, assignment: super::IdentAssignment) -> EmitResult {
        self.emit_ident(&assignment.ident)?;
        self.write(" = ")?;
        self.emit_expr(assignment.assignment)?;

        Ok(())
    }

    fn emit_incr_decr(&mut self, incr_decr: super::IncrDecr) -> EmitResult {
        let (target, fn_name) = match incr_decr {
            super::IncrDecr::Incr(incr) => match incr {
                super::Increment::Pre(tgt) => match tgt {
                    super::IncrDecrTarget::Ident(_) => {
                        todo!()
                    }
                },
                super::Increment::Post(tgt) => match tgt {
                    super::IncrDecrTarget::Ident(ident) => (self.mangle_ident(&ident), "++"),
                },
            },
            super::IncrDecr::Decr(decr) => match decr {
                super::Decrement::Pre(tgt) => match tgt {
                    super::IncrDecrTarget::Ident(_) => {
                        todo!()
                    }
                },
                super::Decrement::Post(tgt) => match tgt {
                    super::IncrDecrTarget::Ident(_) => {
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
