use crate::parser;

use crate::emitter::*;

impl Emitter {
    pub(in crate::emitter) fn emit_chained_obj_op(
        &mut self,
        chained_obj_op: parser::ChainedObjOp,
    ) -> EmitResult {
        let mut curr_acc_type = match chained_obj_op.accessable {
            parser::Accessable::Ident(ref ident) => {
                self.write(&self.mangle_ident(ident))?;

                match self.curr_scope.borrow().get_ident_type(ident) {
                    Some(typ) => typ.clone(),
                    None => return Err(format!("unknown ident '{ident}' in scope")),
                }
            }
            parser::Accessable::LiteralType(_) => todo!(),
        };

        let mut last_op = None;
        let has_assignment = chained_obj_op.assignment.is_some();
        let n = chained_obj_op.obj_ops.len();
        for (i, op) in chained_obj_op.obj_ops.iter().enumerate() {
            match op.clone() {
                parser::ObjOp::Access(prop) => {
                    // If this is the last op and there's an assignment we need to do,
                    // don't emit a `getFieldValue`; just skip it and write the set in a sec.
                    if has_assignment && i == n - 1 {
                        last_op = Some(op);

                        break;
                    }

                    self.emit_get_field_val(&prop)?;

                    if let Some(_) = (*curr_acc_type.borrow()).rest {
                        todo!("complex types")
                    }

                    let typ = (*curr_acc_type.borrow()).head.clone();
                    curr_acc_type = match typ {
                        parser::TypeIdentType::Name(ref typ_name) => self
                            .get_type(typ_name)
                            .ok_or(format!("could not find type {typ_name}"))?,
                        parser::TypeIdentType::LiteralType(typ) => match *typ {
                            parser::LiteralType::FnType { .. } => todo!(),
                            parser::LiteralType::ObjType { fields } => {
                                match fields.iter().find(|field| field.name == prop) {
                                    Some(field) => rcref(field.typ.clone()),
                                    None => {
                                        return Err(format!("unknown field '{prop}' on target"))
                                    }
                                }
                            }
                        },
                        parser::TypeIdentType::Interface(_) => todo!(),
                    };
                }
                parser::ObjOp::Invoc { args } => {
                    self.write("->invoke({")?;

                    let n = args.len();
                    for (i, arg) in args.into_iter().enumerate() {
                        self.write(&format!("TsFunctionArg(\"{}\", ", "arg_name"))?;
                        self.emit_expr(arg)?;
                        self.write(")")?;

                        if i != n - 1 {
                            self.write(", ")?;
                        }
                    }

                    self.write("})")?;
                }
                parser::ObjOp::Arithmetic(artm) => self.emit_arithmetic(artm)?,
            }

            last_op = Some(op)
        }

        if let Some(assignment) = chained_obj_op.assignment {
            let name = match last_op {
                Some(op) => match op {
                    parser::ObjOp::Access(name) => name,
                    parser::ObjOp::Invoc { .. } => todo!(),
                    parser::ObjOp::Arithmetic(_) => todo!(),
                },
                None => unreachable!("assignment without access is impossible"),
            };

            self.write(&format!("->setFieldValue(\"{name}\", "))?;
            self.emit_expr(*assignment)?;
            self.write(")")?;
        }

        Ok(())
    }

    pub(in crate::emitter) fn emit_obj_inst(&mut self, obj_inst: parser::ObjInst) -> EmitResult {
        // TODO: actually impl this thang.
        let type_id = 1;

        let obj_type = rcref(Type {
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
                        "TsObjectFieldDescriptor(TsString(\"{}\"), {type_id}), ",
                        field.name
                    ))?;
                }

                self.enter_scope();
                if let parser::Expr::FnInst(_) = &field.value {
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
