package ir

//go:generate stringer -type=InstructionKind

import (
	"fmt"
	"strings"
)

type Inst struct {
	Kind InstKind

	Type   Type
	Symbol string
	Index  Assignment
	Left   Assignment
	Right  Assignment

	Static  bool
	Literal interface{}
}

func (i *Inst) String() string {
	switch i.Kind {
	case Add, Sub, Mul, Quo,
		Equals, NotEquals, Less, Move,
		And, Or:
		return fmt.Sprintf("%4s = %s(%s, %s)", i.Index, i.Kind, i.Left, i.Right)

	case Not, Push, Ret, End:
		return fmt.Sprintf("%4s = %s(%s)", i.Index, i.Kind, i.Left)

	case Pop:
		return fmt.Sprintf("%4s = %s()", i.Index, i.Kind)

	case LoadEnv:
		return fmt.Sprintf("%4s = %s(%s)", i.Index, i.Kind, i.Literal.(Assignment))

	case Env:
		return fmt.Sprintf("%4s = %s(\"%s\")", i.Index, i.Kind, i.Literal)

	case Bool:
		return fmt.Sprintf("%4s = Bool(%t)", i.Index, i.Literal.(bool))

	case I64:
		return fmt.Sprintf("%4s = Int(%d)", i.Index, i.Literal.(int64))

	case ProcedureType:
		return fmt.Sprintf("%4s = ProcedureType(params: %s, args: %s, return: %s)", i.Index, i.Left, i.Right, i.Literal.(Assignment))

	case ProcedureDefinition:
		return fmt.Sprintf("%4s = ProcedureDefinition(func: %d)", i.Index, i.Literal.(int))

	case Call:
		sb := strings.Builder{}
		sb.WriteString(fmt.Sprintf("%4s = Call(%s)", i.Index, i.Left))
		return sb.String()

	case Phi:
		sb := strings.Builder{}
		sb.WriteString(fmt.Sprintf("%4s = Phi(%s", i.Index, i.Symbol))
		for _, phi := range i.Literal.([]PhiLiteral) {
			sb.WriteString(fmt.Sprintf(", %d:%s", phi.BlockIndex, phi.Assignment))
		}
		sb.WriteString(")")
		return sb.String()

	case GotoIf:
		return fmt.Sprintf("%4s = %s(%s, block:%d)", i.Index, i.Kind, i.Left, i.Literal)

	case Goto:
		return fmt.Sprintf("%4s = %s(block:%d)", i.Index, i.Kind, i.Literal)

	default:
		return fmt.Sprintf("WTF %s", i.Kind)
	}
}

type InstKind int8

const (
	Undefined InstKind = iota

	// Three address
	Add
	Sub
	Mul
	Quo
	Mod
	Less
	Greater
	Equals
	NotEquals
	Move
	And
	Or

	Not

	// Literals
	Default
	Bool
	I8
	I16
	I32
	I64
	F32
	F64
	ProcedureType
	ProcedureDefinition
	ConstructTuple

	// Extra
	LoadEnv
	Env
	Phi
	Ret
	End

	GotoIf
	Goto
	Call
	Push
	Pop
)

type PhiLiteral struct {
	BlockIndex int
	Assignment Assignment
}
