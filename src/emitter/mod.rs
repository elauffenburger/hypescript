use std::{cell::RefCell, rc::Rc};

use maplit::hashmap;

use crate::{
    parser::{self, StmtOrExpr},
    util::rcref,
};

mod scope;
pub use scope::*;

mod stmt;
pub use stmt::*;

mod types;
pub use types::*;

mod expr;
pub use expr::*;

mod runtime;

type EmitterError = String;
type EmitResult = Result<(), EmitterError>;

#[derive(Debug)]
pub struct EmitterResult {
    pub files: Vec<EmittedFile>,
}

#[derive(Debug)]
pub enum EmittedFile {
    File {
        name: String,
        content: String,
    },
    Dir {
        name: String,
        files: Vec<EmittedFile>,
    },
}

pub struct Emitter {
    curr_scope: Rc<RefCell<Scope>>,

    buffer: String,
}

impl Emitter {
    pub fn new() -> Self {
        let global = rcref(scope::new_global_scope());

        Emitter {
            curr_scope: global.clone(),
            buffer: String::new(),
        }
    }

    pub fn emit(mut self, parsed: parser::ParserResult) -> Result<EmitterResult, EmitterError> {
        // Write includes to buffer.
        {
            let includes = [
                "<stdlib.h>",
                "<stdio.h>",
                "<string>",
                "<vector>",
                "<algorithm>",
                "<memory>",
                "\"runtime.hpp\"",
            ];

            for incl in includes {
                self.write(&format!("#include {incl}\n"))?;
            }
        }

        // Write main to buffer.
        {
            self.write("int main() {\n")?;

            for construct in parsed.top_level_constructs {
                match construct {
                    parser::TopLevelConstruct::Interface(iface) => self.reg_iface(iface)?,
                    parser::TopLevelConstruct::StmtOrExpr(stmt_or_expr) => match stmt_or_expr {
                        parser::StmtOrExpr::Stmt(stmt) => self.emit_stmt(stmt)?,
                        parser::StmtOrExpr::Expr(_) => todo!(),
                    },
                }
            }

            self.write("}")?;
        }

        Ok(EmitterResult {
            files: vec![
                // out
                EmittedFile::Dir {
                    name: String::from("src"),
                    files: vec![
                        // out/main.cpp
                        EmittedFile::File {
                            name: String::from("main.cpp"),
                            content: self.buffer,
                        },
                        // out/runtime.cpp
                        EmittedFile::File {
                            name: String::from("runtime.cpp"),
                            content: String::from(runtime::RUNTIME_CPP),
                        },
                        // out/runtime.hpp
                        EmittedFile::File {
                            name: String::from("runtime.hpp"),
                            content: String::from(runtime::RUNTIME_HPP),
                        },
                    ],
                },
            ],
        })
    }

    fn emit_body(&mut self, body: Vec<StmtOrExpr>) -> Result<(), EmitterError> {
        for stmt_or_expr in body {
            match stmt_or_expr {
                crate::parser::StmtOrExpr::Stmt(stmt) => self.emit_stmt(stmt)?,
                crate::parser::StmtOrExpr::Expr(expr) => self.emit_expr(expr)?,
            }
        }

        Ok(())
    }

    fn reg_iface(&mut self, iface: parser::Interface) -> Result<(), EmitterError> {
        self.curr_scope.borrow_mut().add_iface(iface);

        Ok(())
    }

    fn type_of(&self, expr: &parser::Expr) -> Result<Type, EmitterError> {
        self.curr_scope.borrow().type_of(expr)
    }

    fn type_of_expr_inner(&self, expr_inner: &parser::ExprInner) -> Result<Type, EmitterError> {
        self.curr_scope.borrow().type_of_expr_inner(expr_inner)
    }

    fn write(&mut self, code: &str) -> EmitResult {
        self.buffer.push_str(code);

        Ok(())
    }

    fn mangle_ident(&self, ident: &str) -> String {
        match ident {
            "console" => ident.into(),
            _ => format!("_{ident}"),
        }
    }

    fn enter_scope(&mut self) {
        // Create the new scope.
        let scope = rcref(Scope {
            parent: Some(self.curr_scope.clone()),
            children: None,
            ident_types: hashmap! {},
            types: hashmap! {},
            this: self.curr_scope.borrow().this.clone(),
        });

        // Add this scope to the list of child scopes for the parent.
        {
            let mut parent = self.curr_scope.borrow_mut();
            match &mut parent.children {
                Some(children) => children.push(scope.clone()),
                None => parent.children = Some(vec![scope.clone()]),
            };
        }

        // Update the current scope to point to this new scope.
        self.curr_scope = scope
    }

    fn leave_scope(&mut self) {
        let parent = self.curr_scope.borrow().parent.clone();

        match &parent {
            Some(parent) => self.curr_scope = parent.clone(),
            None => unreachable!("can't leave a scope that doesn't have a parent"),
        }
    }

    fn types_equal(&self, left: &Type, right: &Type) -> Result<bool, String> {
        self.curr_scope.borrow().types_equal(left, right)
    }

    fn get_type(&self, name: &str) -> Option<Rc<RefCell<Type>>> {
        self.curr_scope.borrow().get_type(name)
    }
}
