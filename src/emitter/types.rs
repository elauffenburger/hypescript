use std::{cell::RefCell, rc::Rc};

use crate::{
    parser::{self, Module, TypeIdentType},
    util::rcref,
};

thread_local! {
    pub static MOD_CORE: Rc<RefCell<Module>> = rcref(Module {});
}

pub type Type = parser::TypeIdent;

pub enum BuiltInTypes {
    String,
    Number,
    Boolean,
}

impl BuiltInTypes {
    pub fn type_name(&self) -> &str {
        match self {
            BuiltInTypes::String => "string",
            BuiltInTypes::Number => "number",
            BuiltInTypes::Boolean => "boolean",
        }
    }

    pub fn to_type(&self) -> Type {
        MOD_CORE.with(|m| Type::simple(m.clone(), TypeIdentType::name(self.type_name())))
    }
}
