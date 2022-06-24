use crate::parser;

use crate::emitter::{EmitResult, Emitter, Type};

impl Emitter {
    pub(in crate::emitter) fn emit_fn_inst(&mut self, mut fn_inst: parser::FnInst) -> EmitResult {
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
        {
            self.enter_scope();

            let mut did_return = false;
            for stmt_or_expr in fn_inst.body.clone().into_iter() {
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
        }

        // Find all the return stmts in the body and see if we can figure out what the
        // actual return type is.
        {
            let ret_type = {
                let mut acc = None;
                for stmt_or_expr in fn_inst.body.iter() {
                    match stmt_or_expr {
                        parser::StmtOrExpr::Stmt(parser::Stmt::ReturnExpr(ret_expr)) => {
                            let typ = self.type_of(ret_expr).map_err(|e| {
                                format!("couldn't determine type of return expr: {e}")
                            })?;

                            match acc {
                                None => acc = Some(typ),
                                Some(ref t) => {
                                    // If the types are the same, just return the existing acc.
                                    if t == &typ {
                                        continue;
                                    }

                                    todo!("return type consolidation");
                                }
                            }
                        }
                        _ => {}
                    }
                }

                acc
            };

            match fn_inst.return_type {
                // If there's an explicit return type, make sure it lines up with the actual one.
                Some(expl_ret_type) => {
                    match ret_type {
                        Some(ret_type) => {
                            if expl_ret_type != ret_type {
                                return Err(format!(
                                    "fn said it returned {:#?} but it actually returns {:#?}",
                                    expl_ret_type, ret_type
                                ));
                            }
                        }
                        None => match expl_ret_type {
                            parser::TypeIdent {
                                head: parser::TypeIdentType::Name(type_name),
                                rest: None,
                            } => {
                                if &type_name != "void" {
                                    return Err("fn has explicit return type but actual return type does not match".into());
                                }
                            }
                            _ => return Err(
                                "fn has explicit return type but actual return type does not match"
                                    .into(),
                            ),
                        },
                    }
                }
                // If there's no explicit return type on the fn, set it.
                None => {
                    if let Some(ret_typ) = ret_type {
                        fn_inst.return_type = Some(ret_typ);

                        // Patch the fn_inst in the scope.
                        if let Some(name) = fn_inst.name {
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
                        }
                    }
                }
            }
        }

        self.leave_scope();

        // Close up inst.
        self.write("})")?;

        Ok(())
    }
}
