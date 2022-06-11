use crate::parser;

pub type Type = parser::TypeIdent;

pub enum BuiltInTypes {
    String,
}

impl BuiltInTypes {
    pub fn type_name(&self) -> &str {
        match self {
            BuiltInTypes::String => "string",
        }
    }
}