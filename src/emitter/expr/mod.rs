use crate::parser;

use super::{EmitResult, Emitter};

mod fn_inst;
pub use fn_inst::*;

mod obj;
pub use obj::*;

impl Emitter {
    pub(in crate::emitter) fn emit_expr(&mut self, expr: parser::Expr) -> EmitResult {
        match expr {
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

    fn emit_ident(&mut self, ident: &str) -> EmitResult {
        self.write(&self.mangle_ident(ident))
    }
}
