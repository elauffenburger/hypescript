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

        self.write("->getFieldValue(\"")?;
        self.write(match comp.op {
            parser::ComparisonOp::LooseEq => "==",
            parser::ComparisonOp::LooseNeq => "!=",
            parser::ComparisonOp::Lt => "<",
            parser::ComparisonOp::Gt => ">",
        })?;
        self.write("\")->invoke({")?;

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
                    parser::IncrDecrTarget::Ident(ident) => {
                        todo!()
                    }
                },
                parser::Increment::Post(tgt) => match tgt {
                    parser::IncrDecrTarget::Ident(ident) => {
                        (self.mangle_ident(&ident), "_++")
                    }
                },
            },
            parser::IncrDecr::Decr(decr) => match decr {
                parser::Decrement::Pre(tgt) => match tgt {
                    parser::IncrDecrTarget::Ident(ident) => {
                        todo!()
                    }
                },
                parser::Decrement::Post(tgt) => match tgt {
                    parser::IncrDecrTarget::Ident(ident) => {
                        todo!()
                    }
                },
            },
        };

        self.write(&format!("{target}->getFieldValue(\"{fn_name}\")->invoke({{}})"))?;

        Ok(())
    }

    fn emit_ident(&mut self, ident: &str) -> EmitResult {
        self.write(&self.mangle_ident(ident))
    }
}
