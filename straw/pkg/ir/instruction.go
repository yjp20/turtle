package ir

//go:generate stringer -type=InstructionKind

import (
	"fmt"
	"strings"
)

type Instruction struct {
	Kind InstructionKind

	Type   Type
	Symbol string
	Index  Assignment
	Left   Assignment
	Right  Assignment

	Static  bool
	Literal interface{}
}

func (i *Instruction) String() string {
	switch i.Kind {
	case Add, Sub, Mul, Quo,
		Equals, NotEquals, Move,
		And, Or, Phi:
		return fmt.Sprintf("%4s = %s(%s, %s)", i.Index, i.Kind, i.Left, i.Right)

	case Not, Push, Ret, End:
		return fmt.Sprintf("%4s = %s(%s)", i.Index, i.Kind, i.Left)

	case Pop:
		return fmt.Sprintf("%4s = %s()", i.Index, i.Kind)

	case Bool:
		return fmt.Sprintf("%4s = Bool(%t)", i.Index, i.Literal.(bool))

	case I64:
		return fmt.Sprintf("%4s = Int(%d)", i.Index, i.Literal.(int64))

	case Function:
		return fmt.Sprintf("%4s = Function(ref: %s)", i.Index, i.Literal.(int))

	case Call:
		sb := strings.Builder{}
		sb.WriteString(fmt.Sprintf("%4s = Call(%s)", i.Index, i.Left))
		return sb.String()

	case IfTrueGoto:
		return fmt.Sprintf("%4s = %s(%s, %s)", i.Index, i.Kind, i.Left, i.Literal)

	case Goto:
		return fmt.Sprintf("%4s = %s(%s)", i.Index, i.Kind, i.Literal)

	default:
		return "WTF"
	}
}

type InstructionKind int8

const (
	Undefined InstructionKind = iota

	// Three address
	Add
	Sub
	Mul
	Quo
	Mod
	Equals
	NotEquals
	Move
	And
	Or

	Not

	// Literals
	Bool
	I8
	I16
	I32
	I64
	F32
	F64

	// Extra
	Phi
	Ret
	End
	Function

	IfTrueGoto
	Goto
	Call
	Push
	Pop
)

type PhiLiteral []struct {
	Block      string
	Assignment Assignment
}
