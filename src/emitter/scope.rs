use std::{any::Any, cell::RefCell, collections::HashMap, rc::Rc};

use maplit::hashmap;

use super::types::*;
use crate::parser::{self, Expr, ObjTypeField};

#[derive(Default, Debug)]
pub struct Scope {
    pub parent: Option<Rc<RefCell<Scope>>>,
    pub children: Option<Vec<Rc<RefCell<Scope>>>>,
    pub ident_types: HashMap<String, Rc<RefCell<Type>>>,
}

impl Scope {
    /// Adds an ident to the scope and returns an `Rc<RefCell<Type>>` handle to the type.
    pub fn add_ident(&mut self, name: String, typ: Type) -> Rc<RefCell<Type>> {
        let typ = Rc::new(RefCell::new(typ));
        self.ident_types.insert(name, typ.clone());

        typ
    }

    /// Returns Some with the given ident's type if we know about the ident, or None if we don't.
    pub fn get_ident(&self, ident: &str) -> Option<Rc<RefCell<Type>>> {
        match self.ident_types.get(ident) {
            Some(t) => Some(t.to_owned()),
            None => match self.parent.clone() {
                Some(parent) => parent.borrow().get_ident(ident),
                None => None,
            },
        }
    }

    pub fn type_of(&self, expr: &Expr) -> Result<Rc<RefCell<Type>>, String> {
        Ok(match expr {
            Expr::Num(_) => Rc::new(RefCell::new(Type {
                head: parser::TypeIdentType::Name("number".into()),
                rest: None,
            })),
            Expr::Str(_) => Rc::new(RefCell::new(Type {
                head: parser::TypeIdentType::Name("string".into()),
                rest: None,
            })),
            Expr::IdentAssignment(ref ident_assign) => self
                .ident_types
                .get(&ident_assign.ident)
                .ok_or(format!("unknown ident {}", &ident_assign.ident))?
                .clone(),
            Expr::FnInst(ref fn_inst) => Rc::new(RefCell::new(Type {
                head: parser::TypeIdentType::LiteralType(Box::new(parser::LiteralType::FnType {
                    params: fn_inst.params.clone(),
                    return_type: fn_inst.return_type.clone(),
                })),
                rest: None,
            })),
            Expr::ChainedObjOp(_) => todo!(),
            Expr::ObjInst(ref obj_inst) => {
                let mut fields = vec![];
                for field in &obj_inst.fields {
                    let typ = self.type_of(&field.value)?.borrow().clone();

                    fields.push(parser::ObjTypeField {
                        name: field.name.clone(),
                        optional: false,
                        typ: typ,
                    })
                }

                Rc::new(RefCell::new(parser::TypeIdent {
                    head: parser::TypeIdentType::LiteralType(Box::new(
                        parser::LiteralType::ObjType { fields },
                    )),
                    rest: None,
                }))
            }
            Expr::Ident(ref ident) => self
                .ident_types
                .get(ident)
                .ok_or_else(|| format!("unknown ident {}", &ident))?
                .clone(),
        })
    }
}

pub fn new_global_scope() -> Scope {
    Scope {
        parent: None,
        children: None,
        ident_types: hashmap! {
            "console".into() => Rc::new(RefCell::new(Type{
                head: parser::TypeIdentType::LiteralType(
                    Box::new(parser::LiteralType::ObjType {
                        fields: vec![
                            parser::ObjTypeField {
                                name: "log".into(),
                                optional: false,
                                typ: parser::TypeIdent {
                                    head: parser::TypeIdentType::LiteralType(
                                        Box::new(parser::LiteralType::FnType {
                                            params: vec![
                                                parser::FnParam{
                                                    name: "msg".into(),
                                                    optional: false,
                                                    typ: Some(parser::TypeIdent {
                                                        head: parser::TypeIdentType::Name(BuiltInTypes::String.type_name().into()),
                                                        rest: None,
                                                    })
                                                },
                                            ],
                                            return_type: None,
                                        })
                                    ),
                                    rest: None
                                },
                            },
                        ]
                    })
                ),
                rest: None
            })),
        },
    }
}
