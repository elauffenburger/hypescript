use crate::parser::{Expr, Stmt};

use super::{EmitResult, Emitter};

impl Emitter {
    pub(in crate::emitter) fn emit_stmt(&mut self, stmt: Stmt) -> EmitResult {
        match stmt {
            Stmt::If(if_stmt) => {
                self.write("if(")?;
                self.emit_expr(if_stmt.condition)?;
                self.write(")")?;

                self.write("{")?;
                self.emit_body(if_stmt.body)?;
                self.write("}")?;

                for else_if in if_stmt.else_ifs {
                    self.write("else if(")?;
                    self.emit_expr(else_if.condition)?;
                    self.write(")")?;

                    self.write("{")?;
                    self.emit_body(else_if.body)?;
                    self.write("}")?;
                }

                if let Some(els) = if_stmt.els {
                    self.write("else {")?;
                    self.emit_body(els.body)?;
                    self.write("}")?;
                }
            }
            Stmt::ForLoop {
                init,
                condition,
                after,
                body,
            } => {
                self.emit_stmt(*init)?;

                self.write("for(; ")?;
                self.emit_expr(condition)?;
                self.write("->truthy()")?;
                self.write("; ")?;
                self.emit_expr(after)?;
                self.write(") {\n")?;
                self.emit_body(body)?;
                self.write("}\n")?;
            }
            Stmt::LetDecl {
                name,
                typ: expl_typ,
                assignment,
            } => {
                let mangled_name = self.mangle_ident(&name);

                match assignment.clone() {
                    Some(assignment) => {
                        self.write(&format!("auto {mangled_name} = "))?;

                        self.emit_expr(assignment)?;
                    }
                    None => todo!("let decls without assignments"),
                }

                // Register the ident in the current scope.
                {
                    let impl_type = match assignment {
                        Some(ref assignment) => Some(self.type_of(assignment)?),
                        None => todo!("let decls without assignments"),
                    };

                    let typ = match expl_typ {
                        Some(expl_typ) => match impl_type {
                            Some(impl_typ) => {
                                if !self.types_equal(&expl_typ, &impl_typ)? {
                                    return Err(format!("explicit type of {name} marked as {:?}, but resolved implicit type as {:?}", &expl_typ, &impl_typ));
                                }

                                Some(impl_typ)
                            }
                            None => None,
                        },
                        None => impl_type,
                    };

                    match typ {
                        Some(typ) => self.curr_scope.borrow_mut().add_ident(&name, typ),
                        None => return Err(format!("unable to determine the type of {name}")),
                    };
                };
            }
            Stmt::Expr(expr) => {
                if let Expr::FnInst(ref fn_inst) = expr {
                    if let Some(ref name) = fn_inst.name {
                        let mangled_name = self.mangle_ident(name);
                        self.write(&format!("TsFunction* {mangled_name} = "))?;
                    }
                }

                self.emit_expr(expr)?;
            }
            Stmt::ReturnExpr(expr) => {
                self.write("return ")?;
                self.emit_expr(expr)?;
                self.write("\n")?;
            }
            stmt @ _ => todo!("{:?}", stmt),
        }

        self.write(";\n")?;

        Ok(())
    }
}
