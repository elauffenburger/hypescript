use crate::parser::{Expr, Stmt};

use super::{EmitResult, Emitter};

impl Emitter {
    pub(in crate::emitter) fn emit_stmt(&mut self, stmt: Stmt) -> EmitResult {
        match stmt {
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

                for stmt_or_expr in body {
                    match stmt_or_expr {
                        crate::parser::StmtOrExpr::Stmt(stmt) => self.emit_stmt(stmt)?,
                        crate::parser::StmtOrExpr::Expr(expr) => self.emit_expr(expr)?,
                    }
                }

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
        }

        self.write(";\n")?;

        Ok(())
    }
}
