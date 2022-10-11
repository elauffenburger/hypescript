use std::{cell::RefCell, rc::Rc};

use super::Scope;

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
}
