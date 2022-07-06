use crate::parser::{self, TypeIdentType};

pub type Type = parser::TypeIdent;

pub enum BuiltInTypes {
    String,
    Number,
}

impl BuiltInTypes {
    pub fn type_name(&self) -> &str {
        match self {
            BuiltInTypes::String => "string",
            BuiltInTypes::Number => "number",
        }
    }

    pub fn to_type(&self) -> Type {
        Type::simple(TypeIdentType::name(self.type_name()))
    }
}
