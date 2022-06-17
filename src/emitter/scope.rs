use std::{cell::RefCell, collections::HashMap, rc::Rc};

use maplit::hashmap;

use super::types::*;
use crate::parser::{
    Expr, FnParam, Interface, LiteralType, ObjTypeField, TypeIdent, TypeIdentType,
};

#[derive(Debug)]
pub struct Scope {
    pub parent: Option<Rc<RefCell<Scope>>>,
    pub children: Option<Vec<Rc<RefCell<Scope>>>>,

    pub ident_types: HashMap<String, Rc<RefCell<Type>>>,
    pub types: HashMap<String, Rc<RefCell<Type>>>,

    pub this: Rc<RefCell<Type>>,
}

impl Scope {
    /// Adds an ident to the scope and returns an `Rc<RefCell<Type>>` handle to the type.
    pub fn add_ident(&mut self, name: String, typ: Type) -> Rc<RefCell<Type>> {
        let typ = Rc::new(RefCell::new(typ));
        self.ident_types.insert(name, typ.clone());

        typ
    }

    /// Returns Some with the given ident's type if we know about the ident, or None if we don't.
    pub fn get_ident_type(&self, ident: &str) -> Option<Rc<RefCell<Type>>> {
        if ident == "this" {
            return Some(self.this.clone());
        }

        match self.ident_types.get(ident) {
            Some(t) => Some(t.to_owned()),
            None => match self.parent.clone() {
                Some(parent) => parent.borrow().get_ident_type(ident),
                None => None,
            },
        }
    }

    pub fn get_type(&self, name: &str) -> Option<Rc<RefCell<Type>>> {
        match self.types.get(name) {
            Some(t) => Some(t.to_owned()),
            None => match self.parent.clone() {
                Some(parent) => parent.borrow().get_type(name),
                None => None,
            },
        }
    }

    pub fn add_iface(&mut self, iface: Interface) -> Rc<RefCell<Type>> {
        let name = iface.name.clone();

        let typ = Rc::new(RefCell::new(Type {
            head: TypeIdentType::Interface(iface),
            rest: None,
        }));

        self.types.insert(name, typ.clone());

        typ
    }

    pub fn type_of(&self, expr: &Expr) -> Result<Rc<RefCell<Type>>, String> {
        Ok(match expr {
            Expr::Num(_) => Rc::new(RefCell::new(Type {
                head: TypeIdentType::Name("number".into()),
                rest: None,
            })),
            Expr::Str(_) => Rc::new(RefCell::new(Type {
                head: TypeIdentType::Name("string".into()),
                rest: None,
            })),
            Expr::IdentAssignment(ref ident_assign) => self
                .ident_types
                .get(&ident_assign.ident)
                .ok_or(format!("unknown ident {}", &ident_assign.ident))?
                .clone(),
            Expr::FnInst(ref fn_inst) => Rc::new(RefCell::new(Type {
                head: TypeIdentType::LiteralType(Box::new(LiteralType::FnType {
                    params: fn_inst.params.clone(),
                    return_type: fn_inst.return_type.clone(),
                })),
                rest: None,
            })),
            Expr::ChainedObjOp(ref chained_op) => {
                let typ = match chained_op.accessable {
                    crate::parser::Accessable::Ident(ref ident) => self
                        .get_ident_type(ident)
                        .ok_or(format!("could not find ident {ident}"))?,
                    crate::parser::Accessable::LiteralType(ref typ) => {
                        Rc::new(RefCell::new(Type {
                            head: TypeIdentType::LiteralType(Box::new(typ.clone())),
                            rest: None,
                        }))
                    }
                };

                todo!();
                for op in &chained_op.obj_ops {
                    match op {
                        crate::parser::ObjOp::Access(access) => {
                        },
                        crate::parser::ObjOp::Invoc { .. } => todo!(),
                    }
                }

                todo!()
            }
            Expr::ObjInst(ref obj_inst) => {
                let mut fields = vec![];
                for field in &obj_inst.fields {
                    let typ = self.type_of(&field.value)?.borrow().clone();

                    fields.push(ObjTypeField {
                        name: field.name.clone(),
                        optional: false,
                        typ: typ,
                    })
                }

                Rc::new(RefCell::new(TypeIdent {
                    head: TypeIdentType::LiteralType(Box::new(LiteralType::ObjType { fields })),
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

    pub fn types_equal(&self, left: &Type, right: &Type) -> Result<bool, String> {
        match (left, right) {
            (Type { rest: Some(_), .. }, _) | (_, Type { rest: Some(_), .. }) => {
                todo!("complex types")
            }
            (Type { head: left, .. }, Type { head: right, .. }) => {
                self.type_ident_types_equal(left, right)
            }
        }
    }

    fn type_ident_types_equal(
        &self,
        left: &TypeIdentType,
        right: &TypeIdentType,
    ) -> Result<bool, String> {
        // If left just matches right, return true.
        if left == right {
            return Ok(true);
        }

        match (left, right) {
            // If either type is a named type, we need to resolve it and then try again:
            (TypeIdentType::Name(_), _) | (_, TypeIdentType::Name(_)) => {
                let (named_typ, other_typ) = match (left, right) {
                    (TypeIdentType::Name(name), right @ _) => (name, right),
                    (left @ _, TypeIdentType::Name(name)) => (name, left),
                    _ => unreachable!("at least one type should be named here"),
                };

                match (named_typ, other_typ) {
                    (left, TypeIdentType::Name(right)) => Ok(left == right),
                    (left, right @ _) => {
                        let left = self
                            .get_type(left)
                            .ok_or(format!("unknown type '{left}'"))?;
                        let left = left.borrow();

                        if let Some(_) = left.rest {
                            todo!("complex types");
                        }

                        let left = &left.head;

                        self.type_ident_types_equal(left, right)
                    }
                }
            }
            // If either is an interface:
            (TypeIdentType::Interface(_), _) | (_, TypeIdentType::Interface(_)) => {
                let (iface, other_typ) = match (left, right) {
                    (TypeIdentType::Interface(left), _) => (left, right),
                    (_, TypeIdentType::Interface(right)) => (right, left),
                    _ => unreachable!("at least one type should be an interface here"),
                };

                match other_typ.clone() {
                    TypeIdentType::LiteralType(other) => match *other {
                        LiteralType::FnType { .. } => todo!(),
                        LiteralType::ObjType { fields } => {
                            let obj_fields_by_name: HashMap<&str, &ObjTypeField> = fields
                                .iter()
                                .map(|field| (field.name.as_str(), field))
                                .collect();

                            // Make sure that the obj can satisfy each field in the iface.
                            for field in &iface.fields {
                                match obj_fields_by_name.get(field.name.as_str()) {
                                    Some(obj_field) => {
                                        if !self.types_equal(&field.typ, &obj_field.typ)? {
                                            return Err(format!("field {} on obj had wrong type (expected {:?}, received {:?})", &field.name, &field.typ, &obj_field.name));
                                        }
                                    }
                                    None => {
                                        // If the field is optional, we can skip it.
                                        if field.optional {
                                            continue;
                                        } else {
                                            // Otherwise, report an error
                                            return Err(format!(
                                                "obj was missing field {}",
                                                field.name
                                            ));
                                        }
                                    }
                                }
                            }

                            return Ok(true);
                        }
                    },
                    TypeIdentType::Interface(ref other) => Ok(iface == other),
                    TypeIdentType::Name(_) => {
                        unreachable!("should have already resolved named types")
                    }
                }
            }
            _ => Err(format!("not implemented: ({:?}, {:?})", left, right)),
        }
    }
}

pub fn new_global_scope() -> Scope {
    Scope {
        parent: None,
        children: None,
        ident_types: hashmap! {
            "console".into() => Rc::new(RefCell::new(Type{
                head: TypeIdentType::Name("Console".into()),
                rest: None
            })),
        },
        types: hashmap! {
            "Console".into() => Rc::new(RefCell::new(Type{
                head: TypeIdentType::LiteralType(
                    Box::new(LiteralType::ObjType {
                        fields: vec![
                            ObjTypeField {
                                name: "log".into(),
                                optional: false,
                                typ: TypeIdent {
                                    head: TypeIdentType::LiteralType(
                                        Box::new(LiteralType::FnType {
                                            params: vec![
                                                FnParam{
                                                    name: "msg".into(),
                                                    optional: false,
                                                    typ: Some(TypeIdent {
                                                        head: TypeIdentType::Name(BuiltInTypes::String.type_name().into()),
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
            "Global".into() => Rc::new(RefCell::new(Type{
                head: TypeIdentType::LiteralType(Box::new(LiteralType::ObjType { fields: vec![]})),
                rest: None,
            }))
        },
        this: Rc::new(RefCell::new(Type {
            head: TypeIdentType::Name("Global".into()),
            rest: None,
        })),
    }
}
