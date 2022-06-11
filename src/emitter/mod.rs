use std::{cell::RefCell, rc::Rc};

use crate::parser;

mod types;
pub use types::*;

mod scope;
pub use scope::*;

#[derive(Debug)]
pub struct EmitterResult {
    pub code: String,
}

type EmitterError = String;

pub struct Emitter {
    curr_scope: Rc<RefCell<Scope>>,

    buffer: String,
}

impl Emitter {
    pub fn new() -> Self {
        let global = Rc::new(RefCell::new(scope::new_global_scope()));

        Emitter {
            curr_scope: global.clone(),
            buffer: String::new(),
        }
    }

    pub fn emit(mut self, parsed: parser::ParserResult) -> Result<EmitterResult, EmitterError> {
        for incl in [
            "<stdlib.h>",
            "<stdio.h>",
            "<string>",
            "<vector>",
            "<algorithm>",
            "<memory>",
            "\"runtime.hpp\"",
        ] {
            self.write(&format!("#include {incl}\n"))?;
        }

        self.write("int main() {\n")?;

        for construct in parsed.top_level_constructs {
            match construct {
                parser::TopLevelConstruct::Interface(_) => {}
                parser::TopLevelConstruct::StmtOrExpr(stmt_or_expr) => match stmt_or_expr {
                    parser::StmtOrExpr::Stmt(stmt) => self.emit_stmt(stmt)?,
                    parser::StmtOrExpr::Expr(_) => todo!(),
                },
            }
        }

        self.write("}")?;

        Ok(EmitterResult { code: self.buffer })
    }

    fn emit_stmt(&mut self, stmt: parser::Stmt) -> Result<(), EmitterError> {
        match stmt {
            parser::Stmt::LetDecl { .. } => todo!(),
            parser::Stmt::Expr(expr) => {
                match expr {
                    parser::Expr::Num(_) => todo!(),
                    parser::Expr::Str(_) => todo!(),
                    parser::Expr::IdentAssignment(_) => todo!(),
                    parser::Expr::FnInst(ref fn_inst) => {
                        if let Some(ref name) = fn_inst.name {
                            let mangled_name = self.mangle_ident(name);
                            self.write(&format!("TsFunction* {mangled_name} = "))?;
                        }
                    }
                    parser::Expr::ChainedObjOp(_) => {}
                    parser::Expr::ObjInst(_) => todo!(),
                    parser::Expr::Ident(_) => todo!(),
                }

                self.emit_expr(expr)?;
            }
            parser::Stmt::ReturnExpr(_) => todo!(),
        }

        self.write(";\n")?;

        Ok(())
    }

    fn emit_expr(&mut self, expr: parser::Expr) -> Result<(), EmitterError> {
        match expr {
            parser::Expr::Num(_) => todo!(),
            parser::Expr::Str(str) => self.emit_str(&str),
            parser::Expr::IdentAssignment(_) => todo!(),
            parser::Expr::FnInst(fn_inst) => self.emit_fn_inst(fn_inst),
            parser::Expr::ChainedObjOp(chained_obj_op) => self.emit_chained_obj_op(chained_obj_op),
            parser::Expr::ObjInst(_) => todo!(),
            parser::Expr::Ident(_) => todo!(),
        }
    }

    fn emit_fn_inst(&mut self, fn_inst: parser::FnInst) -> Result<(), String> {
        // If the fn is named, add a reference to it in the current scope.
        let fn_name = if let Some(ref name) = &fn_inst.name {
            self.curr_scope.borrow_mut().add_ident(
                name.clone(),
                Type {
                    head: parser::TypeIdentType::LiteralType(Box::new(
                        parser::LiteralType::FnType {
                            params: fn_inst.params.clone(),
                            return_type: fn_inst.return_type.clone(),
                        },
                    )),
                    rest: None,
                },
            );

            name
        } else {
            ""
        };

        // Find all the return stmts in the body and see if we can figure out what the
        // actual return type is.
        let ret_typ = fn_inst.body.iter().fold(None, |acc, stmt_or_expr| {
            match stmt_or_expr {
                parser::StmtOrExpr::Stmt(parser::Stmt::ReturnExpr(ret_expr)) => {
                    let typ = self.type_of(ret_expr);
                    match acc {
                        None => return Some(typ),
                        Some(ref ret_type) => {
                            // If the types are the same, just return the existing acc.
                            if *ret_type == typ {
                                return acc;
                            }

                            todo!("return type consolidation");
                        }
                    }
                }
                _ => acc,
            }
        });

        match fn_inst.return_type {
            // If there's an explicit return type, make sure it lines up with the actual one.
            Some(_) => {
                todo!("confirm return types line up");
            }
            // If there's no explicit return type on the fn, set it.
            None => {
                if let Some(ret_typ) = ret_typ {
                    todo!("patch return type on fn_inst: {ret_typ:#?}");
                }
            }
        }

        // Write start of inst.
        self.write(&format!("new TsFunction(\"{fn_name}\","))?;

        // Write params.
        self.write(&"TsCoreHelpers::toVector<TsFunctionParam>({")?;
        for param in fn_inst.params.iter() {
            let name = &param.name;
            // TODO: actually use type ids.
            let type_id = 0;
            self.write(&format!("TsFunctionParam(\"{name}\", {type_id})"))?;
        }
        self.write("})")?;

        // Write start of lambda.
        self.write(", [=](TsObject* _this, std::vector<TsFunctionArg> args) -> TsObject* {\n")?;

        // Write body of lambda.
        let mut did_return = false;
        for stmt_or_expr in fn_inst.body.into_iter() {
            match stmt_or_expr {
                parser::StmtOrExpr::Stmt(stmt) => {
                    self.emit_stmt(stmt.clone())?;

                    if let parser::Stmt::ReturnExpr(_) = &stmt {
                        did_return = true;
                    }
                }
                parser::StmtOrExpr::Expr(expr) => self.emit_expr(expr)?,
            }
        }

        // If we never explicitly returned, synthesize a return stmt.
        if !did_return {
            self.write("return NULL;\n")?;
        }

        // Close up inst.
        self.write("})")?;

        Ok(())
    }

    fn type_of(&self, expr: &parser::Expr) -> Result<Rc<RefCell<Type>>, EmitterError> {
        self.curr_scope.borrow().type_of(expr)
    }

    fn emit_chained_obj_op(&mut self, chained_obj_op: parser::ChainedObjOp) -> Result<(), String> {
        let mut curr_acc_type = match chained_obj_op.accessable {
            parser::Accessable::Ident(ref ident) => {
                self.write(&self.mangle_ident(ident))?;

                match self.curr_scope.borrow().get_ident(ident) {
                    Some(typ) => typ.clone(),
                    None => return Err(format!("unknown ident '{ident}' in scope")),
                }
            }
            parser::Accessable::LiteralType(_) => todo!(),
        };

        for op in chained_obj_op.obj_ops {
            match op.clone() {
                parser::ObjOp::Access(prop) => {
                    self.write(&format!("->getFieldValue(\"{prop}\")"))?;

                    if let Some(_) = (*curr_acc_type.borrow()).rest {
                        todo!("anything other than a simple type")
                    }

                    let typ = (*curr_acc_type.borrow()).head.clone();
                    curr_acc_type = match typ {
                        parser::TypeIdentType::Name(_) => todo!(),
                        parser::TypeIdentType::LiteralType(typ) => match *typ {
                            parser::LiteralType::FnType { .. } => todo!(),
                            parser::LiteralType::ObjType { fields } => {
                                match fields.iter().find(|field| field.name == prop) {
                                    Some(field) => Rc::new(RefCell::new(field.typ.clone())),
                                    None => {
                                        return Err(format!("unknown field '{prop}' on target"))
                                    }
                                }
                            }
                        },
                    };
                }
                parser::ObjOp::Invoc { args } => {
                    self.write("->invoke(TsCoreHelpers::toVector<TsFunctionArg>({")?;

                    let n = args.len();
                    for (i, arg) in args.into_iter().enumerate() {
                        self.write(&format!("TsFunctionArg(\"{}\", ", "sdjkl"))?;
                        self.emit_expr(arg)?;
                        self.write(")")?;

                        if i != n - 1 {
                            self.write(", ")?;
                        }
                    }

                    self.write("}))")?;
                }
            }
        }

        if let Some(_) = chained_obj_op.assignment {
            todo!("assignment")
        }

        Ok(())
    }

    fn emit_str(&mut self, str: &str) -> Result<(), String> {
        self.write(&format!("new TsString(\"{str}\")"))
    }

    fn write(&mut self, code: &str) -> Result<(), String> {
        self.buffer.push_str(code);

        Ok(())
    }

    fn mangle_ident(&self, ident: &str) -> String {
        if Self::is_built_in_ident(ident) {
            return ident.into();
        }

        return format!("_{ident}");
    }

    fn is_built_in_ident(ident: &str) -> bool {
        match ident {
            "console" => true,
            _ => false,
        }
    }
}
