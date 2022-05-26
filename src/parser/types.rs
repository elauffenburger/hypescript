#[derive(Debug)]
pub enum TopLevelConstruct {
    Interface(Interface),
    StmtOrExpr(StmtOrExpr),
}

#[derive(Debug)]
pub struct Interface {
    pub name: String,
    pub fields: Vec<InterfaceField>,
    pub methods: Vec<InterfaceMethod>,
}

#[derive(Debug)]
pub struct InterfaceMethod {
    pub name: String,
    pub params: Vec<FnParam>,
}

#[derive(Debug)]
pub struct InterfaceField {
    pub name: String,
    pub optional: bool,
    pub typ: TypeIdent,
}

#[derive(Debug)]
pub struct FnParam {
    pub name: String,
    pub optional: bool,
    pub typ: Option<TypeIdent>,
}

#[derive(Debug)]
pub struct TypeIdent {
    pub head: TypeIdentType,
    pub rest: Vec<TypeIdentPart>,
}

#[derive(Debug)]
pub enum TypeIdentPart {
    Union(TypeIdentType),
    Sum(TypeIdentType),
}

#[derive(Debug)]
pub enum TypeIdentType {
    Name(String),
    LiteralType(Box<LiteralType>),
}

#[derive(Debug)]
pub enum LiteralType {
    FnType {
        params: Vec<FnParam>,
        return_type: TypeIdent,
    },
    ObjType {
        fields: Vec<ObjTypeField>,
    },
}

#[derive(Debug)]
pub struct ObjTypeField {
    pub name: String,
    pub optional: bool,
    pub typ: TypeIdent,
}

#[derive(Debug)]
pub enum StmtOrExpr {
    Stmt(Stmt),
    Expr(Expr),
}

#[derive(Debug)]
pub enum Stmt {
    LetDecl {
        name: String,
        typ: Option<TypeIdent>,
        assignment: Option<Expr>,
    },
    Expr(Expr),
    ReturnExpr(Expr),
}

#[derive(Debug)]
pub enum Expr {
    Num(f32),
    Str(String),
    IdentAssignment(Box<IdentAssignment>),
    FnInst(FnInst),
    ChainedObjOp(ChainedObjOp),
    ObjInst(ObjInst),
    Ident(String),
}

#[derive(Debug)]
pub struct FnInst {
    pub name: Option<String>,
    pub params: Vec<FnParam>,
    pub body: Vec<StmtOrExpr>,
    pub return_type: Option<TypeIdent>,
}

#[derive(Debug)]
pub struct ChainedObjOp {
    pub accessable: Accessable,
    pub obj_ops: Vec<ObjOp>,
    pub assignment: Option<Box<Expr>>,
}

#[derive(Debug)]
pub enum Accessable {
    Ident(String),
    LiteralType(LiteralType),
}

#[derive(Debug)]
pub enum ObjOp {
    Access(String),
    Invoc { args: Vec<Expr> },
}

#[derive(Debug)]
pub struct ObjInst {
    pub fields: Vec<ObjFieldInst>,
}

#[derive(Debug)]
pub struct ObjFieldInst {
    pub name: String,
    pub value: Expr,
}

#[derive(Debug)]
pub struct IdentAssignment {
    pub ident: String,
    pub assignment: Expr,
}
