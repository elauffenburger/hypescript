#[derive(Debug, PartialEq)]
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
}

#[derive(Debug, PartialEq, Clone)]
pub struct InterfaceField {
    pub name: String,
    pub optional: bool,
    pub typ: TypeIdent,
}

#[derive(Debug, PartialEq, Clone)]
pub struct FnParam {
    pub name: String,
    pub optional: bool,
    pub typ: Option<TypeIdent>,
}

#[derive(Debug, PartialEq, Clone)]
pub struct TypeIdent {
    pub head: TypeIdentType,
    pub rest: Option<Vec<TypeIdentPart>>,
}

#[derive(Debug, PartialEq, Clone)]
pub enum TypeIdentPart {
    Union(TypeIdentType),
    Sum(TypeIdentType),
}

#[derive(Debug, PartialEq, Clone)]
pub enum TypeIdentType {
    Name(String),
    LiteralType(Box<LiteralType>),
    Interface(Interface),
}

#[derive(Debug, PartialEq, Clone)]
pub enum LiteralType {
    FnType {
        params: Vec<FnParam>,
        return_type: Option<TypeIdent>,
    },
    ObjType {
        fields: Vec<ObjTypeField>,
    },
}

#[derive(Debug, PartialEq, Clone)]
pub struct ObjTypeField {
    pub name: String,
    pub optional: bool,
    pub typ: TypeIdent,
}

#[derive(Debug, PartialEq, Clone)]
pub enum StmtOrExpr {
    Stmt(Stmt),
    Expr(Expr),
}

#[derive(Debug, PartialEq, Clone)]
pub enum Stmt {
    LetDecl {
        name: String,
        typ: Option<TypeIdent>,
        assignment: Option<Expr>,
    },
    Expr(Expr),
    ReturnExpr(Expr),
}

#[derive(Debug, PartialEq, Clone)]
pub enum Expr {
    Num(f32),
    Str(String),
    IdentAssignment(Box<IdentAssignment>),
    FnInst(FnInst),
    ChainedObjOp(ChainedObjOp),
    ObjInst(ObjInst),
    Ident(String),
}

#[derive(Debug, PartialEq, Clone)]
pub struct FnInst {
    pub name: Option<String>,
    pub params: Vec<FnParam>,
    pub body: Vec<StmtOrExpr>,
    pub return_type: Option<TypeIdent>,
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
    LiteralType(LiteralType),
}

#[derive(Debug, PartialEq, Clone)]
pub enum ObjOp {
    Access(String),
    Invoc { args: Vec<Expr> },
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
