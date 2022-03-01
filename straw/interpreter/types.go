package interpreter

type TypeKind int32

const (
	TypeNull TypeKind = iota
	TypeFrame
	TypeDefault
	TypeAny

	// Primitives
	TypeInt32
	TypeInt64
	TypeString
	TypeBool
	TypeFloat64
	TypeFunction
	TypeBuiltinFunction

	// Configurable Types
	TypeArray
	TypeSlice
	TypeStruct
	TypeInterface
	TypeTuple

	// Higher-order types
	TypeType
	TypeFactory
)
