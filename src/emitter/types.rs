use std::borrow::Borrow;

use crate::parser;

use super::Module;

pub trait FromParsed<T> {
    fn from_parsed(m: &Module, parsed: T) -> Self;
}

#[derive(Debug, PartialEq, Clone)]
pub enum TopLevelConstruct {
    Interface(Interface),
    StmtOrExpr(StmtOrExpr),
}

impl FromParsed<&parser::TopLevelConstruct> for TopLevelConstruct {
    fn from_parsed(m: &Module, c: &parser::TopLevelConstruct) -> Self {
        match c {
            parser::TopLevelConstruct::Interface(iface) => {
                TopLevelConstruct::Interface(Interface::from_parsed(m, iface))
            }
            parser::TopLevelConstruct::StmtOrExpr(stmt_or_expr) => {
                TopLevelConstruct::StmtOrExpr(StmtOrExpr::from_parsed(m, stmt_or_expr))
            }
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct Interface {
    pub name: String,
    pub fields: Vec<InterfaceField>,
    pub methods: Vec<InterfaceMethod>,

    pub resolved: bool,
}

impl FromParsed<&parser::Interface> for Interface {
    fn from_parsed(m: &Module, iface: &parser::Interface) -> Self {
        Interface {
            name: iface.name.clone(),
            fields: iface
                .fields
                .iter()
                .map(|f| InterfaceField {
                    name: f.name.clone(),
                    optional: f.optional,
                    typ: Type::from_parsed(m, &f.typ),
                })
                .collect(),
            methods: iface
                .methods
                .iter()
                .map(|mthd| InterfaceMethod {
                    name: mthd.name.clone(),
                    params: mthd
                        .params
                        .iter()
                        .map(|p| FnParam::from_parsed(m, p))
                        .collect(),
                    typ: mthd.typ.as_ref().map(|typ| Type::from_parsed(m, &typ)),
                })
                .collect(),
            resolved: true,
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct InterfaceMethod {
    pub name: String,
    pub params: Vec<FnParam>,
    pub typ: Option<Type>,
}

#[derive(Debug, PartialEq, Clone)]
pub struct InterfaceField {
    pub name: String,
    pub optional: bool,
    pub typ: Type,
}

#[derive(Debug, PartialEq, Clone)]
pub struct FnParam {
    pub name: String,
    pub optional: bool,
    pub typ: Option<Type>,
}

impl FromParsed<&parser::FnParam> for FnParam {
    fn from_parsed(m: &Module, fn_param: &parser::FnParam) -> Self {
        FnParam {
            name: fn_param.name.clone(),
            optional: fn_param.optional,
            typ: fn_param.typ.as_ref().map(|t| Type::from_parsed(m, &t)),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct Type {
    pub mod_path: String,

    pub head: TypeIdentType,
    pub rest: Option<Vec<TypeIdentPart>>,
}

impl Type {
    pub fn simple(mod_path: &str, t: TypeIdentType) -> Self {
        Type {
            mod_path: mod_path.into(),
            head: t,
            rest: None,
        }
    }

    fn path_for_parser_type(m: &Module, typ: &parser::Type) -> Option<String> {
        if typ.rest.is_some() {
            todo!("complex types")
        }

        match &typ.head {
            parser::TypeIdentType::Name(name) => m
                .scope
                .as_ref()
                .borrow()
                .get_type(name)
                .map(|typ| typ.as_ref().borrow().mod_path.clone()),
            parser::TypeIdentType::LiteralType(_) => Some(m.path.clone()),
            parser::TypeIdentType::Interface(_) => Some(m.path.clone()),
        }
    }
}

impl FromParsed<&parser::Type> for Type {
    fn from_parsed(m: &Module, typ: &parser::Type) -> Self {
        Type {
            mod_path: {
                let path = Self::path_for_parser_type(m, typ);
                if path.is_none() {
                    panic!("failed to find mod_path for type");
                }

                path.unwrap()
            },
            head: TypeIdentType::from_parsed(m, &typ.head),
            rest: match typ.rest.as_ref() {
                Some(rest) => Some(
                    rest.iter()
                        .map(|t| match t {
                            parser::TypeIdentPart::Union(u) => {
                                TypeIdentPart::Union(Type::from_parsed(m, u))
                            }
                            parser::TypeIdentPart::Sum(s) => {
                                TypeIdentPart::Sum(Type::from_parsed(m, s))
                            }
                        })
                        .collect(),
                ),
                None => None,
            },
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum TypeIdentPart {
    Union(Type),
    Sum(Type),
}

#[derive(Debug, PartialEq, Clone)]
pub enum TypeIdentType {
    Name(TypeRef),
    LiteralType(Box<LiteralType>),
    Interface(Interface),
}

impl TypeIdentType {
    pub fn name(mod_path: &str, name: &str) -> Self {
        TypeIdentType::Name(TypeRef {
            mod_path: mod_path.into(),
            name: name.into(),
        })
    }

    pub fn literal(t: LiteralType) -> Self {
        TypeIdentType::LiteralType(Box::new(t))
    }
}

impl FromParsed<&parser::TypeIdentType> for TypeIdentType {
    fn from_parsed(m: &Module, typ_ident_typ: &parser::TypeIdentType) -> Self {
        match typ_ident_typ {
            parser::TypeIdentType::Name(name) => TypeIdentType::Name(TypeRef {
                name: name.clone(),
                mod_path: m
                    .get_type(&name)
                    .map(|typ| {
                        let typ_refcell: &std::cell::RefCell<Type> = typ.borrow();
                        let typ: std::cell::Ref<Type> = typ_refcell.borrow();

                        typ.mod_path.clone()
                    })
                    .or_else(|| panic!("failed to find type"))
                    .unwrap(),
            }),
            parser::TypeIdentType::LiteralType(lit) => {
                TypeIdentType::literal(match lit.clone().borrow() {
                    parser::LiteralType::FnType {
                        params,
                        return_type,
                    } => LiteralType::FnType {
                        params: params.iter().map(|p| FnParam::from_parsed(m, p)).collect(),
                        return_type: return_type
                            .as_ref()
                            .map(|ret_typ| Type::from_parsed(m, &ret_typ)),
                    },
                    parser::LiteralType::ObjType { .. } => todo!(),
                })
            }
            parser::TypeIdentType::Interface(iface) => {
                TypeIdentType::Interface(Interface::from_parsed(m, iface))
            }
        }
    }
}

#[derive(Debug, PartialEq, Eq, Clone, Hash)]
pub struct TypeRef {
    pub name: String,
    pub mod_path: String,
}

#[derive(Debug, PartialEq, Clone)]
pub enum LiteralType {
    FnType {
        params: Vec<FnParam>,
        return_type: Option<Type>,
    },
    ObjType {
        fields: Vec<ObjTypeField>,
    },
}

#[derive(Debug, PartialEq, Clone)]
pub struct ObjTypeField {
    pub name: String,
    pub optional: bool,
    pub typ: Type,
}

#[derive(Debug, PartialEq, Clone)]
pub enum StmtOrExpr {
    Stmt(Stmt),
    Expr(Expr),
}

impl FromParsed<&parser::StmtOrExpr> for StmtOrExpr {
    fn from_parsed(m: &Module, stmt_or_expr: &parser::StmtOrExpr) -> Self {
        match stmt_or_expr {
            parser::StmtOrExpr::Stmt(stmt) => StmtOrExpr::Stmt(Stmt::from_parsed(m, stmt)),
            parser::StmtOrExpr::Expr(expr) => StmtOrExpr::Expr(Expr::from_parsed(m, expr)),
        }
    }
}

impl StmtOrExpr {
    pub fn from_parsed_vec(m: &Module, body: &Vec<parser::StmtOrExpr>) -> Vec<StmtOrExpr> {
        body.iter()
            .map(|stmt_or_expr| StmtOrExpr::from_parsed(m, stmt_or_expr))
            .collect()
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum Stmt {
    ForLoop {
        init: Box<Stmt>,
        condition: Expr,
        after: Expr,
        body: Vec<StmtOrExpr>,
    },
    LetDecl {
        name: String,
        typ: Option<Type>,
        assignment: Option<Expr>,
    },
    Expr(Expr),
    ReturnExpr(Expr),
    If(IfStmt),
}

impl FromParsed<&parser::Stmt> for Stmt {
    fn from_parsed(m: &Module, stmt: &parser::Stmt) -> Self {
        match stmt {
            parser::Stmt::ForLoop {
                init,
                condition,
                after,
                body,
            } => Stmt::ForLoop {
                init: { Box::new(Stmt::from_parsed(m, init.borrow())) },
                condition: Expr::from_parsed(m, condition),
                after: Expr::from_parsed(m, &after),
                body: StmtOrExpr::from_parsed_vec(m, body),
            },
            parser::Stmt::LetDecl {
                name,
                typ,
                assignment,
            } => Stmt::LetDecl {
                name: name.clone(),
                typ: typ.as_ref().map(|typ| Type::from_parsed(m, &typ)),
                assignment: assignment.as_ref().map(|expr| Expr::from_parsed(m, &expr)),
            },
            parser::Stmt::Expr(expr) => Stmt::Expr(Expr::from_parsed(m, expr)),
            parser::Stmt::ReturnExpr(expr) => Stmt::ReturnExpr(Expr::from_parsed(m, expr)),
            parser::Stmt::If(stmt) => Stmt::If(IfStmt {
                condition: Expr::from_parsed(m, &stmt.condition),
                body: StmtOrExpr::from_parsed_vec(m, &stmt.body),
                else_ifs: stmt
                    .else_ifs
                    .iter()
                    .map(|expr| ElseIfStmt {
                        condition: Expr::from_parsed(m, &expr.condition),
                        body: StmtOrExpr::from_parsed_vec(m, &expr.body),
                    })
                    .collect(),
                els: stmt.els.as_ref().map(|stmt| ElseStmt {
                    body: StmtOrExpr::from_parsed_vec(m, &stmt.body),
                }),
            }),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct IfStmt {
    pub condition: Expr,
    pub body: Vec<StmtOrExpr>,

    pub else_ifs: Vec<ElseIfStmt>,
    pub els: Option<ElseStmt>,
}

#[derive(Debug, PartialEq, Clone)]
pub struct ElseIfStmt {
    pub condition: Expr,
    pub body: Vec<StmtOrExpr>,
}

#[derive(Debug, PartialEq, Clone)]
pub struct ElseStmt {
    pub body: Vec<StmtOrExpr>,
}

#[derive(Debug, PartialEq, Clone)]
pub struct Expr {
    pub inner: ExprInner,
    pub is_sub_expr: bool,
    pub ops: Vec<ObjOp>,
}

impl FromParsed<&parser::Expr> for Expr {
    fn from_parsed(m: &Module, expr: &parser::Expr) -> Self {
        Expr {
            inner: ExprInner::from_parsed(m, &expr.inner),
            is_sub_expr: expr.is_sub_expr,
            ops: expr
                .ops
                .iter()
                .map(|op| ObjOp::from_parsed(m, op))
                .collect(),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum ExprInner {
    Comparison(Comparison),
    IncrDecr(IncrDecr),
    Num(f32),
    Str(String),
    IdentAssignment(Box<IdentAssignment>),
    FnInst(FnInst),
    ObjInst(ObjInst),
    Ident(String),
}

impl FromParsed<&parser::ExprInner> for ExprInner {
    fn from_parsed(m: &Module, inner: &parser::ExprInner) -> ExprInner {
        match inner {
            parser::ExprInner::Comparison(comp) => {
                ExprInner::Comparison(Comparison::from_parsed(m, comp))
            }
            parser::ExprInner::IncrDecr(incr_decr) => {
                ExprInner::IncrDecr(IncrDecr::from_parsed(m, incr_decr))
            }
            parser::ExprInner::Num(num) => ExprInner::Num(*num),
            parser::ExprInner::Str(str) => ExprInner::Str(str.clone()),
            parser::ExprInner::IdentAssignment(asgn) => {
                ExprInner::IdentAssignment(Box::new(IdentAssignment::from_parsed(m, asgn)))
            }
            parser::ExprInner::FnInst(fn_inst) => {
                ExprInner::FnInst(FnInst::from_parsed(m, fn_inst))
            }
            parser::ExprInner::ObjInst(obj_inst) => {
                ExprInner::ObjInst(ObjInst::from_parsed(m, obj_inst))
            }
            parser::ExprInner::Ident(ident) => ExprInner::Ident(ident.clone()),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum IncrDecr {
    Incr(Increment),
    Decr(Decrement),
}

impl FromParsed<&parser::IncrDecr> for IncrDecr {
    fn from_parsed(_: &Module, incr_decr: &parser::IncrDecr) -> Self {
        match incr_decr {
            parser::IncrDecr::Incr(incr) => IncrDecr::Incr(match incr {
                parser::Increment::Pre(pre) => Increment::Pre(match pre {
                    parser::IncrDecrTarget::Ident(ident) => IncrDecrTarget::Ident(ident.clone()),
                }),
                parser::Increment::Post(post) => Increment::Post(match post {
                    parser::IncrDecrTarget::Ident(ident) => IncrDecrTarget::Ident(ident.clone()),
                }),
            }),
            parser::IncrDecr::Decr(decr) => IncrDecr::Decr(match decr {
                parser::Decrement::Pre(pre) => Decrement::Pre(match pre {
                    parser::IncrDecrTarget::Ident(ident) => IncrDecrTarget::Ident(ident.clone()),
                }),
                parser::Decrement::Post(post) => Decrement::Post(match post {
                    parser::IncrDecrTarget::Ident(ident) => IncrDecrTarget::Ident(ident.clone()),
                }),
            }),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct Comparison {
    pub left: ComparisonTerm,
    pub right: ComparisonTerm,
    pub op: ComparisonOp,
}

impl FromParsed<&parser::Comparison> for Comparison {
    fn from_parsed(m: &Module, comp: &parser::Comparison) -> Self {
        Comparison {
            left: ComparisonTerm::from_parsed(m, &comp.left),
            right: ComparisonTerm::from_parsed(m, &comp.right),
            op: ComparisonOp::from_parsed(m, &comp.op),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum ComparisonTerm {
    IncrDecr(IncrDecr),
    Num(f32),
    Str(String),
    IdentAssignment(Box<IdentAssignment>),
    Ident(String),
    Comparison(Box<Comparison>),
    Arithmetic(Arithmetic),
}

impl FromParsed<&parser::ComparisonTerm> for ComparisonTerm {
    fn from_parsed(m: &Module, term: &parser::ComparisonTerm) -> Self {
        match term {
            parser::ComparisonTerm::IncrDecr(incr_decr) => {
                ComparisonTerm::IncrDecr(IncrDecr::from_parsed(m, incr_decr))
            }
            parser::ComparisonTerm::Num(num) => ComparisonTerm::Num(*num),
            parser::ComparisonTerm::Str(str) => ComparisonTerm::Str(str.clone()),
            parser::ComparisonTerm::IdentAssignment(asgn) => ComparisonTerm::IdentAssignment(
                Box::new(IdentAssignment::from_parsed(m, asgn.as_ref())),
            ),
            parser::ComparisonTerm::Ident(ident) => ComparisonTerm::Ident(ident.clone()),
            parser::ComparisonTerm::Comparison(comp) => {
                ComparisonTerm::Comparison(Box::new(Comparison::from_parsed(m, comp)))
            }
            parser::ComparisonTerm::Arithmetic(arth) => {
                ComparisonTerm::Arithmetic(Arithmetic::from_parsed(m, arth))
            }
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct Arithmetic {
    pub term: ArithmeticTerm,
    pub ops: Vec<(ArithmeticOp, ArithmeticTerm)>,
}

impl FromParsed<&parser::Arithmetic> for Arithmetic {
    fn from_parsed(m: &Module, arthm: &parser::Arithmetic) -> Self {
        Arithmetic {
            term: ArithmeticTerm::from_parsed(m, &arthm.term),
            ops: arthm
                .ops
                .iter()
                .map(|op_pair| {
                    let op = match op_pair.0 {
                        parser::ArithmeticOp::Add => ArithmeticOp::Add,
                        parser::ArithmeticOp::Sub => ArithmeticOp::Sub,
                        parser::ArithmeticOp::Mult => ArithmeticOp::Mult,
                        parser::ArithmeticOp::Div => ArithmeticOp::Div,
                        parser::ArithmeticOp::Modu => ArithmeticOp::Modu,
                    };

                    let term = ArithmeticTerm::from_parsed(m, &op_pair.1);

                    (op, term)
                })
                .collect(),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum ArithmeticOp {
    Add,
    Sub,
    Mult,
    Div,
    Modu,
}

#[derive(Debug, PartialEq, Clone)]
pub enum ArithmeticTerm {
    Ident(String),
    Num(f32),
}

impl FromParsed<&parser::ArithmeticTerm> for ArithmeticTerm {
    fn from_parsed(_: &Module, term: &parser::ArithmeticTerm) -> Self {
        match term {
            parser::ArithmeticTerm::Ident(ident) => ArithmeticTerm::Ident(ident.clone()),
            parser::ArithmeticTerm::Num(num) => ArithmeticTerm::Num(*num),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum ComparisonOp {
    LooseEq,
    LooseNeq,
    Lt,
    Gt,
    And,
}

impl FromParsed<&parser::ComparisonOp> for ComparisonOp {
    fn from_parsed(_: &Module, op: &parser::ComparisonOp) -> Self {
        match op {
            parser::ComparisonOp::LooseEq => ComparisonOp::LooseEq,
            parser::ComparisonOp::LooseNeq => ComparisonOp::LooseNeq,
            parser::ComparisonOp::Lt => ComparisonOp::Lt,
            parser::ComparisonOp::Gt => ComparisonOp::Gt,
            parser::ComparisonOp::And => ComparisonOp::And,
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum Increment {
    Pre(IncrDecrTarget),
    Post(IncrDecrTarget),
}

#[derive(Debug, PartialEq, Clone)]
pub enum Decrement {
    Pre(IncrDecrTarget),
    Post(IncrDecrTarget),
}

#[derive(Debug, PartialEq, Clone)]
pub enum IncrDecrTarget {
    Ident(String),
}

#[derive(Debug, PartialEq, Clone)]
pub struct FnInst {
    pub name: Option<String>,
    pub params: Vec<FnParam>,
    pub body: Vec<StmtOrExpr>,
    pub return_type: Option<Type>,
}

impl FromParsed<&parser::FnInst> for FnInst {
    fn from_parsed(m: &Module, fn_inst: &parser::FnInst) -> Self {
        FnInst {
            name: fn_inst.name.as_ref().map(|name| name.clone()),
            params: fn_inst
                .params
                .iter()
                .map(|param| FnParam::from_parsed(m, &param))
                .collect(),
            body: StmtOrExpr::from_parsed_vec(m, &fn_inst.body),
            return_type: fn_inst
                .return_type
                .as_ref()
                .map(|typ| Type::from_parsed(m, &typ)),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct ChainedObjOp {
    pub accessable: Accessable,
    pub obj_ops: Vec<ObjOp>,
    pub assignment: Option<Box<Expr>>,
}

#[derive(Debug, PartialEq, Clone)]
pub enum Accessable {
    Ident(String),
    FnInst(FnInst),
}

#[derive(Debug, PartialEq, Clone)]
pub enum ObjOp {
    Access(String),
    Invoc { args: Vec<Expr> },
    Arithmetic(Arithmetic),
    ComparisonOp(ComparisonOp),
    Assignment(Expr),
}

impl FromParsed<&parser::ObjOp> for ObjOp {
    fn from_parsed(m: &Module, op: &parser::ObjOp) -> Self {
        match op {
            parser::ObjOp::Access(name) => ObjOp::Access(name.clone()),
            parser::ObjOp::Invoc { args } => ObjOp::Invoc {
                args: args.iter().map(|arg| Expr::from_parsed(m, &arg)).collect(),
            },
            parser::ObjOp::Arithmetic(arthm) => {
                ObjOp::Arithmetic(Arithmetic::from_parsed(m, arthm))
            }
            parser::ObjOp::ComparisonOp(op) => {
                ObjOp::ComparisonOp(ComparisonOp::from_parsed(m, op))
            }
            parser::ObjOp::Assignment(asgn) => ObjOp::Assignment(Expr::from_parsed(m, asgn)),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct ObjInst {
    pub fields: Vec<ObjFieldInst>,
}

impl FromParsed<&parser::ObjInst> for ObjInst {
    fn from_parsed(m: &Module, obj_inst: &parser::ObjInst) -> Self {
        ObjInst {
            fields: obj_inst
                .fields
                .iter()
                .map(|field| ObjFieldInst {
                    name: field.name.clone(),
                    value: Expr::from_parsed(m, &field.value),
                })
                .collect(),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct ObjFieldInst {
    pub name: String,
    pub value: Expr,
}

#[derive(Debug, PartialEq, Clone)]
pub struct IdentAssignment {
    pub ident: String,
    pub assignment: Expr,
}

impl FromParsed<&parser::IdentAssignment> for IdentAssignment {
    fn from_parsed(m: &Module, asign: &parser::IdentAssignment) -> IdentAssignment {
        IdentAssignment {
            ident: asign.ident.clone(),
            assignment: Expr::from_parsed(m, &asign.assignment),
        }
    }
}
