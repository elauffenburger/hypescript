#[derive(Debug, PartialEq)]
pub enum TopLevelConstruct {
    Interface(Interface),
    StmtOrExpr(StmtOrExpr),
}

#[derive(Debug, PartialEq)]
pub struct Interface {
    pub name: String,
    pub fields: Vec<InterfaceField>,
    pub methods: Vec<InterfaceMethod>,
}

#[derive(Debug, PartialEq)]
pub struct InterfaceMethod {
    pub name: String,
    pub params: Vec<FnParam>,
}

#[derive(Debug, PartialEq)]
pub struct InterfaceField {
    pub name: String,
    pub optional: bool,
    pub typ: TypeIdent,
}

#[derive(Debug, PartialEq)]
pub struct FnParam {
    pub name: String,
    pub optional: bool,
    pub typ: Option<TypeIdent>,
}

#[derive(Debug, PartialEq)]
pub struct TypeIdent {
    pub head: TypeIdentType,
    pub rest: Vec<TypeIdentPart>,
}

#[derive(Debug, PartialEq)]
pub enum TypeIdentPart {
    Union(TypeIdentType),
    Sum(TypeIdentType),
}

#[derive(Debug, PartialEq)]
pub enum TypeIdentType {
    Name(String),
    LiteralType(Box<LiteralType>),
}

#[derive(Debug, PartialEq)]
pub enum LiteralType {
    FnType {
        params: Vec<FnParam>,
        return_type: TypeIdent,
    },
    ObjType {
        fields: Vec<ObjTypeField>,
    },
}

#[derive(Debug, PartialEq)]
pub struct ObjTypeField {
    pub name: String,
    pub optional: bool,
    pub typ: TypeIdent,
}

#[derive(Debug, PartialEq)]
pub enum StmtOrExpr {
    Stmt(Stmt),
    Expr(Expr),
}

#[derive(Debug, PartialEq)]
pub enum Stmt {
    LetDecl {
        name: String,
        typ: Option<TypeIdent>,
        assignment: Option<Expr>,
    },
    Expr(Expr),
    ReturnExpr(Expr),
}

#[derive(Debug, PartialEq)]
pub enum Expr {
    Num(f32),
    Str(String),
    IdentAssignment(Box<IdentAssignment>),
    FnInst(FnInst),
    ChainedObjOp(ChainedObjOp),
    ObjInst(ObjInst),
    Ident(String),
}

#[derive(Debug, PartialEq)]
pub struct FnInst {
    pub name: Option<String>,
    pub params: Vec<FnParam>,
    pub body: Vec<StmtOrExpr>,
    pub return_type: Option<TypeIdent>,
}

#[derive(Debug, PartialEq)]
pub struct ChainedObjOp {
    pub accessable: Accessable,
    pub obj_ops: Vec<ObjOp>,
    pub assignment: Option<Box<Expr>>,
}

#[derive(Debug, PartialEq)]
pub enum Accessable {
    Ident(String),
    LiteralType(LiteralType),
}

#[derive(Debug, PartialEq)]
pub enum ObjOp {
    Access(String),
    Invoc { args: Vec<Expr> },
}

#[derive(Debug, PartialEq)]
pub struct ObjInst {
    pub fields: Vec<ObjFieldInst>,
}

#[derive(Debug, PartialEq)]
pub struct ObjFieldInst {
    pub name: String,
    pub value: Expr,
}

#[derive(Debug, PartialEq)]
pub struct IdentAssignment {
    pub ident: String,
    pub assignment: Expr,
}
