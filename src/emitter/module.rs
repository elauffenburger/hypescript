use std::{cell::RefCell, rc::Rc};

use super::*;

#[derive(Debug, PartialEq, Clone)]
pub struct Module {
    pub path: String,
    pub scope: Rc<RefCell<Scope>>,

    pub top_level_constructs: Vec<TopLevelConstruct>,
}

impl Module {
    pub fn new(path: &str, scope: Rc<RefCell<Scope>>) -> Self {
        Module {
            path: path.into(),
            scope,
            top_level_constructs: vec![],
        }
    }

    pub fn get_type(&self, name: &str) -> Option<Rc<RefCell<Type>>> {
        self.scope.borrow().get_type(name)
    }

    pub fn path_for_type(&self, typ: &str) -> Option<String> {
        self.scope
            .borrow()
            .get_type(typ)
            .map(|t| t.borrow().mod_path.clone())
    }
}

impl From<&parser::Module> for Module {
    fn from(parsed_mod: &parser::Module) -> Self {
        let mut m = Module {
            path: parsed_mod.path.clone(),
            scope: rcref(new_mod_scope(parsed_mod.path.clone())),
            top_level_constructs: vec![],
        };

        m.register_types(parsed_mod).unwrap_or_else(|err| {
            eprintln!("failed to register types: {err}");
            panic!("failed to register types")
        });

        for parsed_construct in parsed_mod.top_level_constructs.iter() {
            let construct = TopLevelConstruct::from_parsed(&m, &parsed_construct);
            m.top_level_constructs.push(construct);
        }

        m
    }
}

impl Module {
    fn register_types(&mut self, m: &parser::Module) -> Result<(), String> {
        let mut frontier = vec![];
        frontier.extend(m.top_level_constructs.iter());

        let mut seen = hashset! {};
        while !frontier.is_empty() {
            let construct = frontier.remove(0);
            match construct {
                parser::TopLevelConstruct::Interface(iface) => {
                    let iface_name = &iface.name;

                    // Find types the iface references and if any of them are unknown, add this iface to the frontier.
                    let fields = iface.fields.clone();
                    let mut resolved_all_fields = true;
                    for field in fields.iter() {
                        if field.typ.rest.is_some() {
                            return Err("todo: complex types".into());
                        }

                        let has_type = match &field.typ.head {
                            parser::TypeIdentType::Name(name) => self.get_type(name).is_some(),
                            parser::TypeIdentType::LiteralType(_) => todo!(),
                            parser::TypeIdentType::Interface(_) => todo!(),
                        };

                        if !has_type {
                            resolved_all_fields = false;
                            break;
                        }
                    }

                    if !resolved_all_fields {
                        if !seen.contains(&iface_name) {
                            frontier.push(construct);
                            seen.insert(iface_name);
                            continue;
                        }
                    }

                    // Put a placeholder type in for the iface.
                    self.scope.borrow_mut().types.insert(
                        iface_name.clone(),
                        rcref(Type {
                            mod_path: self.path.clone(),
                            head: TypeIdentType::Interface(Interface {
                                name: iface_name.clone(),
                                fields: vec![],
                                methods: vec![],
                                resolved: false,
                            }),
                            rest: None,
                        }),
                    );

                    let iface = Interface::from_parsed(self, iface);
                    self.scope.borrow_mut().add_iface(iface);

                    // Remove iface from frontier.
                    //
                    // TODO: make this not awful.
                    frontier = frontier
                        .into_iter()
                        .filter(|c| match c {
                            parser::TopLevelConstruct::Interface(iface) => {
                                iface.name != *iface_name
                            }
                            _ => true,
                        })
                        .collect();
                }
                parser::TopLevelConstruct::StmtOrExpr(_) => {}
            };
        }

        Ok(())
    }
}
