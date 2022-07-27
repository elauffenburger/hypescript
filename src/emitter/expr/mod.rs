use crate::parser;

use super::{EmitResult, Emitter};

mod fn_inst;
pub use fn_inst::*;

mod obj;
pub use obj::*;

impl Emitter {
    pub(in crate::emitter) fn emit_expr(&mut self, expr: parser::Expr) -> EmitResult {
        match expr {
            parser::Expr::Comparison(comp) => self.emit_comparison(comp),
            parser::Expr::IncrDecr(incr_decr) => self.emit_incr_decr(incr_decr),
            parser::Expr::Num(num) => self.emit_num(num),
            parser::Expr::Str(str) => self.emit_str(&str),
            parser::Expr::IdentAssignment(ident_assignment) => {
                self.emit_ident_assignment(*ident_assignment)
            }
            parser::Expr::FnInst(fn_inst) => self.emit_fn_inst(fn_inst),
            parser::Expr::ChainedObjOp(chained_obj_op) => self.emit_chained_obj_op(chained_obj_op),
            parser::Expr::ObjInst(obj_inst) => self.emit_obj_inst(obj_inst),
            parser::Expr::Ident(ident) => self.emit_ident(&ident),
        }
    }

    fn emit_comparison(&mut self, comp: parser::Comparison) -> EmitResult {
        self.emit_comparison_term(comp.left)?;

        self.emit_get_field_val(match comp.op {
            parser::ComparisonOp::LooseEq => "==",
            parser::ComparisonOp::LooseNeq => "!=",
            parser::ComparisonOp::Lt => "<",
            parser::ComparisonOp::Gt => ">",
        })?;

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
            parser::ComparisonTerm::ChainedObjOp(chained_obj_op) => {
                self.emit_chained_obj_op(chained_obj_op)
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
        self.emit_arithmetic_term(arthm.term)?;

        for op in arthm.ops {
            self.emit_get_field_val(match op.0 {
                parser::ArithmeticOp::Add => "+",
                parser::ArithmeticOp::Sub => "-",
                parser::ArithmeticOp::Mult => "*",
                parser::ArithmeticOp::Div => "/",
                parser::ArithmeticOp::Modu => "%",
            })?;

            self.write(&format!("->invoke({{"))?;
            self.emit_arithmetic_term(op.1)?;
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
                    parser::IncrDecrTarget::Ident(ident) => (self.mangle_ident(&ident), "_++"),
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
