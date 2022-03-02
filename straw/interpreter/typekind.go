package interpreter

//go:generate stringer -type=TypeKind

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
