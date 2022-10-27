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
        todo!()
    }
}

impl From<&parser::Module> for Module {
    fn from(parsed_mod: &parser::Module) -> Self {
        let mut m = Module{
            path: parsed_mod.path.clone(),
            scope: rcref(new_mod_scope(parsed_mod.path.clone())),
            top_level_constructs: vec![],
        };

        for parsed_construct in parsed_mod.top_level_constructs.iter() {
           let construct = TopLevelConstruct::from_parsed(&m, &parsed_construct); 
           m.top_level_constructs.push(construct);
        }

        m
    }
}
