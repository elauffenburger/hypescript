use std::{cell::RefCell, rc::Rc};

use super::*;

#[derive(Debug, PartialEq, Clone)]
pub struct Module {
    pub name: String,
    pub path: String,
    pub scope: Rc<RefCell<Scope>>,
}

impl Module {
    pub fn new(name: &str, path: &str, scope: Rc<RefCell<Scope>>) -> Self {
        Module{
            name: name.into(),
            path: path.into(),
            scope,
        }
    }

    pub fn get_type(&self, name: &str) -> Option<Rc<RefCell<Type>>> {
        self.scope.borrow().get_type(name)
    }
}
