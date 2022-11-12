use std::{cell::RefCell, collections::HashMap, rc::Rc};

use maplit::hashmap;

use crate::util::rcref;

use super::*;

#[derive(Debug, PartialEq)]
pub struct Scope {
    pub parent: Option<Rc<RefCell<Scope>>>,
    pub children: Option<Vec<Rc<RefCell<Scope>>>>,

    pub ident_types: HashMap<String, Rc<RefCell<Type>>>,
    pub types: HashMap<String, Rc<RefCell<Type>>>,

    pub this: Rc<RefCell<Type>>,
    pub mod_path: String,
}

impl Scope {
    pub fn new(mod_path: String) -> Self {
        Scope {
            parent: None,
            children: None,
            ident_types: hashmap! {},
            types: hashmap! {},
            this: rcref(Type::simple(
                MOD_CORE_PATH,
                TypeIdentType::name(MOD_CORE_PATH, "Global"),
            )),
            mod_path,
        }
    }

    /// Adds an ident to the scope and returns an `Rc<RefCell<Type>>` handle to the type.
    pub fn add_ident(&mut self, name: &str, typ: Type) -> Rc<RefCell<Type>> {
        let typ = rcref(typ);
        self.ident_types.insert(name.to_owned(), typ.clone());

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
            Some(t) => Some(t.clone()),
            None => match self.parent.clone() {
                Some(parent) => parent.borrow().get_type(name),
                None => None,
            },
        }
    }

    pub fn add_iface(&mut self, iface: Interface) -> Rc<RefCell<Type>> {
        let name = iface.name.clone();
        let typ = Type {
            mod_path: self.mod_path.clone(),
            head: TypeIdentType::Interface(iface),
            rest: None,
        };

        // If we already had an interface def, make sure it was an unresolved iface.
        if let Some(existing) = self.get_type(&name) {
            if let Some(_) = existing.borrow().rest {
                panic!("expected interface");
            }

            match existing.borrow_mut().head {
                TypeIdentType::Interface(ref iface) => {
                    if iface.resolved {
                        panic!("expected unresolved")
                    }
                }
                _ => panic!("expected interface"),
            };

            existing.swap(&RefCell::new(typ));
            return self.get_type(&name).unwrap();
        }

        let typ = rcref(typ);
        self.types.insert(name, typ.clone());

        typ
    }

    fn type_ident_type_field_type(
        &self,
        typ: &TypeIdentType,
        field_name: &str,
    ) -> Result<Type, String> {
        match typ {
            TypeIdentType::Name(ref type_ref) => {
                let t = self
                    .get_type(&type_ref.name)
                    .ok_or(format!("unknown type {type_ref:?}"))?;
                let t = t.borrow().clone();

                if let Some(_) = t.rest {
                    todo!("complex types")
                }

                self.type_ident_type_field_type(&t.head, field_name)
            }
            TypeIdentType::LiteralType(typ) => match *typ.clone() {
                LiteralType::FnType { .. } => todo!(),
                LiteralType::ObjType { ref fields } => {
                    let field = fields
                        .iter()
                        .find(|field| &field.name == field_name)
                        .ok_or(format!("failed to find field '{field_name}' on {:?}", typ))?;

                    return Ok(field.typ.clone());
                }
            },
            TypeIdentType::Interface(ref iface) => {
                let field = iface
                    .fields
                    .iter()
                    .find(|field| &field.name == field_name)
                    .ok_or(format!("failed to find field '{field_name}' on {:?}", typ))?;

                return Ok(field.typ.clone());
            }
        }
    }

    fn type_field_type(&self, typ: &Type, field_name: &str) -> Result<Type, String> {
        if let Some(_) = typ.rest {
            todo!("complex types")
        }

        self.type_ident_type_field_type(&typ.head, field_name)
    }

    pub fn invoc_type(&self, typ: &Type) -> Result<Type, String> {
        if let Some(_) = &typ.rest {
            todo!("complex types");
        }

        match typ.head.clone() {
            TypeIdentType::Name(ref type_ref) => {
                let t = self
                    .get_type(&type_ref.name)
                    .ok_or(format!("unknown type {type_ref:?}"))?;
                let t = t.borrow();

                if let Some(_) = t.rest {
                    todo!("complex types");
                }

                return self.invoc_type(&t);
            }
            TypeIdentType::LiteralType(typ) => match *typ {
                LiteralType::FnType {
                    return_type: ret_typ,
                    ..
                } => return Ok(ret_typ.ok_or("failed to resolve return type".to_owned())?),
                LiteralType::ObjType { .. } => todo!(),
            },
            TypeIdentType::Interface(_) => todo!(),
        }
    }

    pub fn type_of_expr_inner(&self, expr_inner: &ExprInner) -> Result<Type, String> {
        Ok(match expr_inner {
            ExprInner::Comparison(_) => BuiltInTypes::Boolean.to_type(),
            ExprInner::IncrDecr(_) => BuiltInTypes::Number.to_type(),
            ExprInner::Num(_) => BuiltInTypes::Number.to_type(),
            ExprInner::Str(_) => BuiltInTypes::String.to_type(),
            ExprInner::IdentAssignment(ref ident_assign) => self
                .get_ident_type(&ident_assign.ident)
                .unwrap()
                .borrow()
                .clone(),
            ExprInner::FnInst(ref fn_inst) => Type {
                mod_path: self.mod_path.clone(),
                head: TypeIdentType::literal(LiteralType::FnType {
                    params: fn_inst.params.clone(),
                    return_type: fn_inst.return_type.clone(),
                }),
                rest: None,
            },
            ExprInner::ObjInst(ref obj_inst) => {
                let mut fields = vec![];
                for field in &obj_inst.fields {
                    let typ = self.type_of(&field.value)?;

                    fields.push(ObjTypeField {
                        name: field.name.clone(),
                        optional: false,
                        typ: typ,
                    })
                }

                Type::simple(
                    &self.mod_path.clone(),
                    TypeIdentType::literal(LiteralType::ObjType { fields }),
                )
            }
            ExprInner::Ident(ref ident) => self.get_ident_type(ident).unwrap().borrow().clone(),
        })
    }

    pub fn type_of(&self, expr: &Expr) -> Result<Type, String> {
        let mut typ = self.type_of_expr_inner(&expr.inner)?;

        // Walk through each obj op and update the typ to match the last op's type.
        for op in &expr.ops {
            match op {
                super::ObjOp::Access(ref access) => typ = self.type_field_type(&typ, access)?,
                super::ObjOp::Invoc { .. } => typ = self.invoc_type(&typ)?,
                super::ObjOp::Arithmetic(_) => typ = BuiltInTypes::Number.to_type(),
                super::ObjOp::ComparisonOp(_) => typ = BuiltInTypes::Boolean.to_type(),
                r @ _ => todo!("{:?}", r),
            }
        }

        Ok(typ)
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
                            .get_type(&left.name)
                            .ok_or(format!("unknown type '{left:?}'"))?;
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
                    TypeIdentType::Name { .. } => {
                        unreachable!("should have already resolved named types")
                    }
                }
            }
            _ => Err(format!("not implemented: ({:?}, {:?})", left, right)),
        }
    }
}

pub fn new_mod_scope(mod_path: String) -> Scope {
    GLOBAL_SCOPE.with(|global_scope| {
        Scope {
            mod_path,
            parent: Some(global_scope.clone()),
            children: None,
            ident_types: hashmap! {},
            types: hashmap! {},
            // Wire up the scope's `this` to have the type `Global`.
            this: rcref(Type::simple(
                MOD_CORE_PATH.clone(),
                TypeIdentType::name(MOD_CORE_PATH, "Global"),
            )),
        }
    })
}
