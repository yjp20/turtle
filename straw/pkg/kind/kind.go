package kind

//go:generate stringer -type=Kind

type Kind int8

const (
	Unresolved Kind = iota
	None
	Null
	Default
	Any
	Frame

	Bool

	IntConstant
	I8
	I16
	I32
	I64
	U8
	U16
	U32
	U64
	F32
	F64

	StringConstant
	String

	Function
	BuiltinFunction

	Array
	Slice
	Struct
	Interface
	Tuple
	Range

	Type
	Factory
)
