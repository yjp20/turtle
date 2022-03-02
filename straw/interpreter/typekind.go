package interpreter

//go:generate stringer -type=TypeKind

type TypeKind int32

const (
	TypeNull TypeKind = iota
	TypeFrame
	TypeDefault
	TypeAny

	// Primitives
	TypeI32
	TypeI64
	TypeString
	TypeBool
	TypeF64
	TypeFunction
	TypeBuiltinFunction

	// Complex Types
	TypeArray
	TypeSlice
	TypeStruct
	TypeInterface
	TypeTuple
	TypeRange

	// Higher-order types
	TypeType
	TypeFactory
)
