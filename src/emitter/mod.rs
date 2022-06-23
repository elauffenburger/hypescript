use std::{cell::RefCell, rc::Rc};

use crate::parser::{
    self, Expr, FnInst, IdentAssignment, LiteralType, ObjInst, ObjTypeField, TypeIdentType,
};

mod types;
use maplit::hashmap;
pub use types::*;

mod scope;
pub use scope::*;

#[derive(Debug)]
pub struct EmitterResult {
    pub code: String,
}

type EmitterError = String;
type EmitResult = Result<(), EmitterError>;

pub struct Emitter {
    curr_scope: Rc<RefCell<Scope>>,

    buffer: String,
}

impl Emitter {
    pub fn new() -> Self {
        let global = Rc::new(RefCell::new(scope::new_global_scope()));

        Emitter {
            curr_scope: global.clone(),
            buffer: String::new(),
        }
    }

    pub fn emit(mut self, parsed: parser::ParserResult) -> Result<EmitterResult, EmitterError> {
        for incl in [
            "<stdlib.h>",
            "<stdio.h>",
            "<string>",
            "<vector>",
            "<algorithm>",
            "<memory>",
            "\"runtime.hpp\"",
        ] {
            self.write(&format!("#include {incl}\n"))?;
        }

        self.write("int main() {\n")?;

        for construct in parsed.top_level_constructs {
            match construct {
                parser::TopLevelConstruct::Interface(iface) => self.reg_iface(iface)?,
                parser::TopLevelConstruct::StmtOrExpr(stmt_or_expr) => match stmt_or_expr {
                    parser::StmtOrExpr::Stmt(stmt) => self.emit_stmt(stmt)?,
                    parser::StmtOrExpr::Expr(_) => todo!(),
                },
            }
        }

        self.write("}")?;

        Ok(EmitterResult { code: self.buffer })
    }

    fn reg_iface(&mut self, iface: parser::Interface) -> Result<(), EmitterError> {
        self.curr_scope.borrow_mut().add_iface(iface);

        Ok(())
    }

    fn emit_stmt(&mut self, stmt: parser::Stmt) -> EmitResult {
        match stmt {
            parser::Stmt::LetDecl {
                name,
                typ: expl_typ,
                assignment,
            } => {
                let mangled_name = self.mangle_ident(&name);

                match assignment.clone() {
                    Some(assignment) => {
                        self.write(&format!("auto {mangled_name} = "))?;

                        self.emit_expr(assignment)?;
                    }
                    None => todo!("let decls without assignments"),
                }

                // Register the ident in the current scope.
                {
                    let impl_type = match assignment {
                        Some(ref assignment) => Some(self.type_of(assignment)?),
                        None => todo!("let decls without assignments"),
                    };

                    let typ = match expl_typ {
                        Some(expl_typ) => match impl_type {
                            Some(impl_typ) => {
                                if !self.types_equal(&expl_typ, &impl_typ)? {
                                    return Err(format!("explicit type of {name} marked as {:?}, but resolved implicit type as {:?}", &expl_typ, &impl_typ));
                                }

                                Some(impl_typ)
                            }
                            None => None,
                        },
                        None => impl_type,
                    };

                    match typ {
                        Some(typ) => self.curr_scope.borrow_mut().add_ident(name, typ),
                        None => return Err(format!("unable to determine the type of {name}")),
                    };
                };
            }
            parser::Stmt::Expr(expr) => {
                if let parser::Expr::FnInst(ref fn_inst) = expr {
                    if let Some(ref name) = fn_inst.name {
                        let mangled_name = self.mangle_ident(name);
                        self.write(&format!("TsFunction* {mangled_name} = "))?;
                    }
                }

                self.emit_expr(expr)?;
            }
            parser::Stmt::ReturnExpr(expr) => {
                self.write("return ")?;
                self.emit_expr(expr)?;
                self.write("\n")?;
            }
        }

        self.write(";\n")?;

        Ok(())
    }

    fn emit_expr(&mut self, expr: parser::Expr) -> EmitResult {
        match expr {
            parser::Expr::Num(num) => self.emit_num(num),
            parser::Expr::Str(str) => self.emit_str(&str),
            parser::Expr::IdentAssignment(ident_assignment) => {
                self.emit_ident_assignment(*ident_assignment)
            }
            parser::Expr::FnInst(fn_inst) => self.emit_fn_inst(fn_inst),
            parser::Expr::ChainedObjOp(chained_obj_op) => self.emit_chained_obj_op(chained_obj_op),
            parser::Expr::ObjInst(obj_inst) => self.emit_obj_inst(obj_inst),
            parser::Expr::Ident(ident) => self.emit_ident(&ident),
        }
    }

    fn emit_fn_inst(&mut self, mut fn_inst: parser::FnInst) -> EmitResult {
        // If the fn is named, add a reference to it in the current scope.
        let fn_name = if let Some(ref name) = &fn_inst.name {
            self.curr_scope.borrow_mut().add_ident(
                name.clone(),
                Type {
                    head: parser::TypeIdentType::LiteralType(Box::new(
                        parser::LiteralType::FnType {
                            params: fn_inst.params.clone(),
                            return_type: fn_inst.return_type.clone(),
                        },
                    )),
                    rest: None,
                },
            );

            name
        } else {
            ""
        };

        // Write start of inst.
        self.write(&format!("new TsFunction(\"{fn_name}\","))?;

        // Write params.
        self.write(&"TsCoreHelpers::toVector<TsFunctionParam>({")?;
        for param in fn_inst.params.iter() {
            let name = &param.name;
            // TODO: actually use type ids.
            let type_id = 0;
            self.write(&format!("TsFunctionParam(\"{name}\", {type_id})"))?;
        }
        self.write("})")?;

        // Write start of lambda.
        self.write(", [=](TsObject* _this, std::vector<TsFunctionArg> args) -> TsObject* {\n")?;

        // Write body of lambda.
        {
            self.enter_scope();

            let mut did_return = false;
            for stmt_or_expr in fn_inst.body.clone().into_iter() {
                match stmt_or_expr {
                    parser::StmtOrExpr::Stmt(stmt) => {
                        self.emit_stmt(stmt.clone())?;

                        if let parser::Stmt::ReturnExpr(_) = &stmt {
                            did_return = true;
                        }
                    }
                    parser::StmtOrExpr::Expr(expr) => self.emit_expr(expr)?,
                }
            }

            // If we never explicitly returned, synthesize a return stmt.
            if !did_return {
                self.write("return NULL;\n")?;
            }
        }

        // Find all the return stmts in the body and see if we can figure out what the
        // actual return type is.
        {
            let ret_type = {
                let mut acc = None;
                for stmt_or_expr in fn_inst.body.iter() {
                    match stmt_or_expr {
                        parser::StmtOrExpr::Stmt(parser::Stmt::ReturnExpr(ret_expr)) => {
                            let typ = self.type_of(ret_expr).map_err(|e| {
                                format!("couldn't determine type of return expr: {e}")
                            })?;

                            match acc {
                                None => acc = Some(typ),
                                Some(ref t) => {
                                    // If the types are the same, just return the existing acc.
                                    if t == &typ {
                                        continue;
                                    }

                                    todo!("return type consolidation");
                                }
                            }
                        }
                        _ => {}
                    }
                }

                acc
            };

            match fn_inst.return_type {
                // If there's an explicit return type, make sure it lines up with the actual one.
                Some(expl_ret_type) => {
                    match ret_type {
                        Some(ret_type) => {
                            if expl_ret_type != ret_type {
                                return Err(format!(
                                    "fn said it returned {:#?} but it actually returns {:#?}",
                                    expl_ret_type, ret_type
                                ));
                            }
                        }
                        None => match expl_ret_type {
                            parser::TypeIdent {
                                head: TypeIdentType::Name(type_name),
                                rest: None,
                            } => {
                                if &type_name != "void" {
                                    return Err("fn has explicit return type but actual return type does not match".into());
                                }
                            }
                            _ => return Err(
                                "fn has explicit return type but actual return type does not match"
                                    .into(),
                            ),
                        },
                    }
                }
                // If there's no explicit return type on the fn, set it.
                None => {
                    if let Some(ret_typ) = ret_type {
                        fn_inst.return_type = Some(ret_typ);

                        // Patch the fn_inst in the scope.
                        if let Some(name) = fn_inst.name {
                            self.curr_scope.borrow_mut().add_ident(
                                name.clone(),
                                Type {
                                    head: parser::TypeIdentType::LiteralType(Box::new(
                                        parser::LiteralType::FnType {
                                            params: fn_inst.params.clone(),
                                            return_type: fn_inst.return_type.clone(),
                                        },
                                    )),
                                    rest: None,
                                },
                            );
                        }
                    }
                }
            }
        }

        self.leave_scope();

        // Close up inst.
        self.write("})")?;

        Ok(())
    }

    fn type_of(&self, expr: &parser::Expr) -> Result<Type, EmitterError> {
        self.curr_scope.borrow().type_of(expr)
    }

    fn emit_chained_obj_op(&mut self, chained_obj_op: parser::ChainedObjOp) -> EmitResult {
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
        for op in chained_obj_op.obj_ops {
            match op.clone() {
                parser::ObjOp::Access(prop) => {
                    self.write(&format!("->getFieldValue(\"{prop}\")"))?;

                    if let Some(_) = (*curr_acc_type.borrow()).rest {
                        todo!("anything other than a simple type")
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
                                    Some(field) => Rc::new(RefCell::new(field.typ.clone())),
                                    None => {
                                        return Err(format!("unknown field '{prop}' on target"))
                                    }
                                }
                            }
                        },
                        TypeIdentType::Interface(_) => todo!(),
                    };
                }
                parser::ObjOp::Invoc { args } => {
                    self.write("->invoke(TsCoreHelpers::toVector<TsFunctionArg>({")?;

                    let n = args.len();
                    for (i, arg) in args.into_iter().enumerate() {
                        self.write(&format!("TsFunctionArg(\"{}\", ", "sdjkl"))?;
                        self.emit_expr(arg)?;
                        self.write(")")?;

                        if i != n - 1 {
                            self.write(", ")?;
                        }
                    }

                    self.write("}))")?;
                }
            }

            last_op = Some(op)
        }

        if let Some(assignment) = chained_obj_op.assignment {
            let name = match last_op {
                Some(op) => match op {
                    parser::ObjOp::Access(name) => name,
                    parser::ObjOp::Invoc { .. } => todo!(),
                },
                None => unreachable!("assignment without access is impossible"),
            };

            self.write(&format!("->setFieldValue(\"{name}\", "))?;
            self.emit_expr(*assignment)?;
            self.write(")")?;
        }

        Ok(())
    }

    fn emit_str(&mut self, str: &str) -> EmitResult {
        self.write(&format!("new TsString(\"{str}\")"))
    }

    fn emit_num(&mut self, num: f32) -> EmitResult {
        self.write(&format!("new TsNum({num})"))
    }

    fn emit_ident(&mut self, ident: &str) -> EmitResult {
        self.write(&self.mangle_ident(ident))
    }

    fn emit_ident_assignment(&mut self, assignment: IdentAssignment) -> EmitResult {
        self.emit_ident(&assignment.ident)?;
        self.write(" = ")?;
        self.emit_expr(assignment.assignment)?;

        Ok(())
    }

    fn emit_obj_inst(&mut self, obj_inst: ObjInst) -> EmitResult {
        // TODO: actually impl this thang.
        let type_id = 1;

        let obj_type = Rc::new(RefCell::new(Type {
            head: TypeIdentType::LiteralType(Box::new(LiteralType::ObjType {
                fields: {
                    let mut fields = vec![];

                    for field in obj_inst.fields.iter() {
                        fields.push(ObjTypeField {
                            name: field.name.clone(),
                            optional: false,
                            typ: self.type_of(&field.value)?,
                        });
                    }

                    fields
                },
            })),
            rest: None,
        }));

        // Start the obj.
        self.write(&format!("new TsObject({type_id}, "))?;

        // Write the fields.
        {
            self.write("TsCoreHelpers::toVector<ToObjectField*>({")?;

            let n = obj_inst.fields.len();
            for (i, field) in obj_inst.fields.into_iter().enumerate() {
                self.write("new TsObjectField(")?;

                // Write the field descriptor.
                {
                    // TODO: actually impl this thang.
                    let type_id = 0;

                    self.write(&format!(
                        "TsObjectFieldDescriptor(TsString(\"{}\"), {type_id})",
                        field.name
                    ))?;
                }

                self.enter_scope();
                if let Expr::FnInst(_) = &field.value {
                    self.curr_scope.borrow_mut().this = obj_type.clone();
                }

                // Write the value.
                self.emit_expr(field.value)?;

                self.write(")")?;

                if i != n - 1 {
                    self.write(", ")?;
                }
            }

            self.write("})")?;
        }

        // End the obj.
        self.write(")")?;

        Ok(())
    }

    fn write(&mut self, code: &str) -> EmitResult {
        self.buffer.push_str(code);

        Ok(())
    }

    fn mangle_ident(&self, ident: &str) -> String {
        match ident {
            "console" => ident.into(),
            _ => format!("_{ident}"),
        }
    }

    fn enter_scope(&mut self) {
        // Create the new scope.
        let scope = Rc::new(RefCell::new(Scope {
            parent: Some(self.curr_scope.clone()),
            children: None,
            ident_types: hashmap! {},
            types: hashmap! {},
            this: self.curr_scope.borrow().this.clone(),
        }));

        // Add this scope to the list of child scopes for the parent.
        {
            let mut parent = self.curr_scope.borrow_mut();
            match &mut parent.children {
                Some(children) => children.push(scope.clone()),
                None => parent.children = Some(vec![scope.clone()]),
            };
        }

        // Update the current scope to point to this new scope.
        self.curr_scope = scope
    }

    fn leave_scope(&mut self) {
        let parent = self.curr_scope.borrow().parent.clone();

        match &parent {
            Some(parent) => self.curr_scope = parent.clone(),
            None => unreachable!("can't leave a scope that doesn't have a parent"),
        }
    }

    fn types_equal(&self, left: &Type, right: &Type) -> Result<bool, String> {
        self.curr_scope.borrow().types_equal(left, right)
    }

    fn get_type(&self, name: &str) -> Option<Rc<RefCell<Type>>> {
        self.curr_scope.borrow().get_type(name)
    }
}
