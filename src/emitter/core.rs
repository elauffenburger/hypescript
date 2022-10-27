use std::{cell::RefCell, rc::Rc};

use strum::IntoEnumIterator;
use strum_macros::EnumIter;

use crate::util::rcref;

use super::*;

pub const MOD_CORE_PATH: &str = "_/core";

thread_local! {
    pub static GLOBAL_SCOPE: Rc<RefCell<Scope>> = {
        let mut scope = Scope::new(MOD_CORE_PATH.into());

        // Add `Console` interface.
        scope.add_iface(Interface {
            name: "Console".into(),
            methods: vec![],
            fields: vec![InterfaceField {
                name: "log".into(),
                optional: false,
                typ: Type::simple(MOD_CORE_PATH, TypeIdentType::literal(
                    LiteralType::FnType {
                        params: vec![FnParam {
                            name: "msg".into(),
                            optional: false,
                            typ: Some(Type {
                                mod_path: MOD_CORE_PATH.into(),
                                head: TypeIdentType::name(MOD_CORE_PATH, BuiltInTypes::String.type_name()),
                                rest: None,
                            }),
                        }],
                        return_type: None,
                    }),
                ),
            }],
        });

        // Add `console` ident.
        scope.add_ident("console", Type::simple(MOD_CORE_PATH, TypeIdentType::name(MOD_CORE_PATH, "Console")));

        // Add `global` interface.
        scope.add_iface(Interface{name: "Global".into(), fields: vec![], methods: vec![]});

        // Add built-in types.
        for typ in BuiltInTypes::iter() {
            match typ {
                BuiltInTypes::String => scope.add_iface(Interface{name: "string".into(), fields: vec![], methods: vec![]}),
                BuiltInTypes::Number => scope.add_iface(Interface{name: "number".into(), fields: vec![], methods: vec![]}),
                BuiltInTypes::Boolean => scope.add_iface(Interface{name: "boolean".into(), fields: vec![], methods: vec![]}),
                BuiltInTypes::Void => scope.add_iface(Interface{name: "void".into(), fields: vec![], methods: vec![]}),
            };
        }

        rcref(scope)
    };

    pub static MOD_CORE: Rc<RefCell<Module>> = {
        GLOBAL_SCOPE.with(|scope| rcref(Module::new(MOD_CORE_PATH, scope.clone())))
    };
}

#[derive(Debug, EnumIter)]
pub enum BuiltInTypes {
    String,
    Number,
    Boolean,
    Void,
}

impl BuiltInTypes {
    pub fn type_name(&self) -> &str {
        match self {
            BuiltInTypes::String => "string",
            BuiltInTypes::Number => "number",
            BuiltInTypes::Boolean => "boolean",
            BuiltInTypes::Void => "void",
        }
    }

    pub fn to_type(&self) -> Type {
        Type::simple(
            MOD_CORE_PATH,
            TypeIdentType::name(MOD_CORE_PATH, self.type_name()),
        )
    }
}
