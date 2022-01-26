package core

type PrimitiveType string

const (
	TsString PrimitiveType = "string"
	TsNumber PrimitiveType = "number"
	TsVoid   PrimitiveType = "void"
	TsAny    PrimitiveType = "any"
)

var PrimitiveTypes = []PrimitiveType{TsString, TsNumber, TsVoid}

type RuntimeType string

const (
	RtTsObject   RuntimeType = "TsObject"
	RtTsFunction RuntimeType = "TsFunction"
	RtTsVoid     RuntimeType = "void"
)

type TypeId int

const (
	TypeIdNone       TypeId = 0
	TypeIdTsObject   TypeId = 1
	TypeIdTsNum      TypeId = 2
	TypeIdTsString   TypeId = 3
	TypeIdTsFunction TypeId = 4
	TypeIdVoid       TypeId = 5
	TypeIdIntrinsic  TypeId = 6
)
