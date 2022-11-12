use std::{cell::RefCell, collections::HashMap, rc::Rc};

use maplit::{hashmap, hashset};

use crate::{parser, util::rcref};

mod core;
pub use self::core::*;

mod expr;
pub use expr::*;

mod module;
pub use module::*;

mod runtime;

mod scope;
pub use scope::*;

mod stmt;
pub use stmt::*;

mod types;
pub use types::*;

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
    modules: HashMap<String, Rc<RefCell<Module>>>,

    buffer: String,
}

impl Emitter {
    pub fn new() -> Self {
        MOD_CORE.with(|mod_core| {
            Emitter {
                // Create a dummy scope.
                // HACK: this feels...wrong -- probably should be using an Option, but it feels weird because you should have a handle when we're actually running the emitter.
                curr_scope: rcref(Scope::new("_/dummy".into())),

                // Register the global module.
                modules: hashmap! {
                    String::from(MOD_CORE_PATH) => mod_core.clone(),
                },

                buffer: String::new(),
            }
        })
    }

    pub fn emit(mut self, parsed_mods: &[parser::Module]) -> Result<EmitterResult, EmitterError> {
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

            for parsed_mod in parsed_mods {
                // Parse the module and add it to the emitter.
                let module: Rc<RefCell<Module>> = rcref(parsed_mod.into());
                self.modules.insert(module.borrow().path.clone(), module.clone());

                // Apply the module scope to the current emitter scope.
                let old_scope = self.curr_scope.clone();
                self.curr_scope = module.borrow().scope.clone();

                for construct in module.borrow().top_level_constructs.iter() {
                    let construct = construct.clone();

                    match construct {
                        TopLevelConstruct::Interface(_) => {}
                        TopLevelConstruct::StmtOrExpr(stmt_or_expr) => match stmt_or_expr {
                            StmtOrExpr::Stmt(stmt) => self.emit_stmt(stmt)?,
                            StmtOrExpr::Expr(_) => todo!(),
                        },
                    }
                }

                self.curr_scope = old_scope;
            }

            self.write("}")?;
        }

        Ok(EmitterResult {
            files: vec![
                // out
                EmittedFile::Dir {
                    name: "src".into(),
                    files: vec![
                        // out/main.cpp
                        EmittedFile::File {
                            name: "main.cpp".into(),
                            content: self.buffer,
                        },
                        // out/runtime.cpp
                        EmittedFile::File {
                            name: "runtime.cpp".into(),
                            content: runtime::RUNTIME_CPP.into(),
                        },
                        // out/runtime.hpp
                        EmittedFile::File {
                            name: "runtime.hpp".into(),
                            content: runtime::RUNTIME_HPP.into(),
                        },
                    ],
                },
            ],
        })
    }

    fn emit_body(&mut self, body: Vec<StmtOrExpr>) -> Result<(), EmitterError> {
        for stmt_or_expr in body {
            match stmt_or_expr {
                StmtOrExpr::Stmt(stmt) => self.emit_stmt(stmt)?,
                StmtOrExpr::Expr(expr) => self.emit_expr(expr)?,
            }
        }

        Ok(())
    }

    fn type_of(&self, expr: &Expr) -> Result<Type, EmitterError> {
        self.curr_scope.borrow().type_of(expr)
    }

    fn type_of_expr_inner(&self, expr_inner: &ExprInner) -> Result<Type, EmitterError> {
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
        let mod_path = self.curr_scope.borrow().mod_path.clone();
        self.enter_scope_in_module(&mod_path)
    }

    fn enter_scope_in_module(&mut self, mod_path: &str) {
        // Create the new scope.
        let scope = rcref(Scope {
            mod_path: mod_path.into(),
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

    fn get_type(&self, type_ref: &TypeRef) -> Option<Rc<RefCell<Type>>> {
        let mod_path: String = type_ref.mod_path.clone().into();
        match self.modules.get(&mod_path) {
            Some(module) => module.borrow().get_type(&type_ref.name),
            None => {
                eprintln!("failed to find module {mod_path:#?}");
                None
            }
        }
    }
}
