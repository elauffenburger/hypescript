use crate::parser;

use crate::emitter::*;

impl Emitter {
    pub(in crate::emitter) fn emit_obj_inst(&mut self, obj_inst: parser::ObjInst) -> EmitResult {
        // TODO: actually impl this thang.
        let type_id = 1;

        let obj_type = rcref(Type {
            module: self.curr_scope.borrow().module.clone(),
            head: parser::TypeIdentType::literal(parser::LiteralType::ObjType {
                fields: {
                    let mut fields = vec![];

                    for field in obj_inst.fields.iter() {
                        fields.push(parser::ObjTypeField {
                            name: field.name.clone(),
                            optional: false,
                            typ: self.type_of(&field.value)?,
                        });
                    }

                    fields
                },
            }),
            rest: None,
        });

        // Start the obj.
        self.write(&format!("new TsObject({type_id}, "))?;

        // Write the fields.
        {
            self.write("{")?;

            let n = obj_inst.fields.len();
            for (i, field) in obj_inst.fields.into_iter().enumerate() {
                self.write("new TsObjectField(")?;

                // Write the field descriptor.
                {
                    // TODO: actually impl this thang.
                    let type_id = 0;

                    self.write(&format!(
                        "TsObjectFieldDescriptor(\"{}\", {type_id}), ",
                        field.name
                    ))?;
                }

                self.enter_scope();
                if let parser::ExprInner::FnInst(_) = &field.value.inner {
                    self.curr_scope.borrow_mut().this = obj_type.clone();
                }

                // Write the value.
                self.emit_expr(field.value)?;

                self.write(")")?;

                if i != n - 1 {
                    self.write(", ")?;
                }
            }

            self.write("}")?;
        }

        // End the obj.
        self.write(")")?;

        Ok(())
    }
}
