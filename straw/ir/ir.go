package ir

import (
	"fmt"
	"strings"
)

//go:generate stringer -type=InstructionType

type InstructionKind int8

const (
	Undefined InstructionKind = iota

	// Three address
	Add
	Sub
	Mul
	Equals
	NotEquals
	Move
	IfTrueGoto

	// Literals
	Int
	Label

	// Extra
	Decl
	Function
	Arg
	Call
)

type Instruction struct {
	Kind InstructionKind

	Type   Type
	Symbol string
	Index  Assignment
	Left   Assignment
	Right  Assignment

	Static  bool
	Extra   []Assignment
	Literal interface{}
}

func (i *Instruction) Print() string {
	switch i.Kind {
	case Add, Sub, Equals, NotEquals, Move, IfTrueGoto, Mul:
		return fmt.Sprintf("%s = %s(%s, %s)", i.Index, i.Kind, i.Left, i.Right)

	case Arg:
		return fmt.Sprintf("%s = Arg(%d)", i.Index, i.Literal.(int))

	case Int:
		return fmt.Sprintf("%s = Int(%d)", i.Index, i.Literal.(int64))

	case Function:
		return fmt.Sprintf("%s = Function(%s)", i.Index, i.Literal.(string))

	case Call:
		sb := strings.Builder{}
		sb.WriteString(fmt.Sprintf("%s = Call(%s", i.Index, i.Left))
		for _, arg := range i.Extra {
			sb.WriteString(", ")
			sb.WriteString(fmt.Sprintf("%s", arg))
		}
		sb.WriteString(")")
		return sb.String()

	default:
		return "WTF"
	}
}
