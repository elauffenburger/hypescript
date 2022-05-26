package core

type Function struct {
	Name               *string
	Parameters         []*FunctionParameter
	ImplicitReturnType *TypeSpec
	ExplicitReturnType *TypeSpec
	Body               []*StatementOrExpression
}

type StatementOrExpression struct {
	Statement  *Statement
	Expression *Expression

	Scope *Scope
}

type Expression struct {
	Number                 *Number
	String                 *string
	IdentAssignment        *IdentAssignment
	FunctionInstantiation  *Function
	ChainedObjectOperation *ChainedObjectOperation
	ObjectInstantiation    *ObjectInstantiation
	Ident                  *string
}

type ObjectTypeField struct {
	Name string
	Type *TypeSpec
}

type FunctionParameter struct {
	Name     string
	Optional bool
	Type     *TypeSpec
}

type Union struct {
	Types map[*TypeSpec]bool
}

type IdentAssignment struct {
	Ident      string
	Assignment Assignment
}

type ChainedObjectOperation struct {
	First *ObjectOperation
	Last  *ObjectOperation
}

type ObjectOperation struct {
	Accessee *Accessable

	Access     *ObjectAccess
	Invocation *ObjectInvocation
	Assignment *Assignment

	Next *ObjectOperation
	Prev *ObjectOperation
}

type ObjectInvocation struct {
	Accessee  *Accessable
	Arguments []*Expression
}

type ObjectAccess struct {
	AccessedIdent string
	Type          *TypeSpec
}

type Accessable struct {
	Ident *string
	Type  *TypeSpec
}

type Number struct {
	Integer *int
}

type LetDecl struct {
	Name         string
	ExplicitType *TypeSpec
	Value        *Expression
}

type Statement struct {
	FunctionInstantiation *Function
	ExpressionStmt        *Expression
	LetDecl               *LetDecl
	ReturnStmt            *Expression
}

type Assignment struct {
	Value *Expression
}

type ObjectInstantiation struct {
	Fields []*ObjectFieldInstantiation
}

type ObjectFieldInstantiation struct {
	Name  string
	Type  *TypeSpec
	Value *Expression
}

type TopLevelConstruct struct {
	InterfaceDefinition   *Interface
	StatementOrExpression *StatementOrExpression
}

type TS struct {
	TopLevelConstructs []TopLevelConstruct
}

/// ContainsAllTypeSpecs returns true if right contains all specs in left.
func ContainsAllTypeSpecs(left, right []*TypeSpec) bool {
	unseen := left[:]

	for _, t1 := range unseen {
		for _, t2 := range right {
			if t1.EqualsStrict(t2) {
				unseen = unseen[1:]
				break
			}
		}
	}

	return len(unseen) == 0
}
