use crate::emitter::FromParsed;

#[derive(Debug, PartialEq, Clone)]
pub enum TopLevelConstruct {
    Interface(Interface),
    StmtOrExpr(StmtOrExpr),
}

#[derive(Debug, PartialEq, Clone)]
pub struct Interface {
    pub name: String,
    pub fields: Vec<InterfaceField>,
    pub methods: Vec<InterfaceMethod>,
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

#[derive(Debug, PartialEq, Clone)]
pub struct Type {
    pub head: TypeIdentType,
    pub rest: Option<Vec<TypeIdentPart>>,
}

impl Type {
    pub fn simple(t: TypeIdentType) -> Self {
        Type {
            head: t,
            rest: None,
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
    Name(String),
    LiteralType(Box<LiteralType>),
    Interface(Interface),
}

impl TypeIdentType {
    pub fn name(name: &str) -> Self {
        TypeIdentType::Name(name.into())
    }

    pub fn literal(t: LiteralType) -> Self {
        TypeIdentType::LiteralType(Box::new(t))
    }
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

#[derive(Debug, PartialEq, Clone)]
pub enum IncrDecr {
    Incr(Increment),
    Decr(Decrement),
}

#[derive(Debug, PartialEq, Clone)]
pub struct Comparison {
    pub left: ComparisonTerm,
    pub right: ComparisonTerm,
    pub op: ComparisonOp,
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

#[derive(Debug, PartialEq, Clone)]
pub struct Arithmetic {
    pub term: ArithmeticTerm,
    pub ops: Vec<(ArithmeticOp, ArithmeticTerm)>,
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

#[derive(Debug, PartialEq, Clone)]
pub enum ComparisonOp {
    LooseEq,
    LooseNeq,
    Lt,
    Gt,
    And,
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

#[derive(Debug, PartialEq, Clone)]
pub struct ObjInst {
    pub fields: Vec<ObjFieldInst>,
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
