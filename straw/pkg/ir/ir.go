package ir

import (
	"fmt"
	"strings"
)

//go:generate stringer -type=InstructionKind

type Program struct {
	Blocks []Block
	Names  map[string]int
}

func (p *Program) AppendBlock(block Block) {
	p.Blocks = append(p.Blocks, block)
	p.Names[block.Name] = block.Index
}

func (p Program) Get(idx int) *Block {
	if idx >= len(p.Blocks) {
		return nil
	}
	return &p.Blocks[idx]
}

func (p Program) Lookup(name string) *Block {
	if idx, ok := p.Names[name]; ok {
		return p.Get(idx)
	}
	return nil
}

func (p Program) String() string {
	sb := strings.Builder{}
	for _, block := range p.Blocks {
		sb.WriteString(fmt.Sprintf("= %d %s\n", block.Index, block.Name))
		for _, inst := range block.Instructions {
			sb.WriteString(inst.String())
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

func (p Program) Next(b *Block) *Block {
	return p.Get(b.Index + 1)
}


type Block struct {
	Index        int
	Name         string
	Instructions []Instruction
}

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

func (i *Instruction) String() string {
	switch i.Kind {
	case Add, Sub, Mul, Quo,
		Equals, NotEquals, Move, IfTrueGoto,
		And, Or:
		return fmt.Sprintf("%s = %s(%s, %s)", i.Index, i.Kind, i.Left, i.Right)

	case Not:
		return fmt.Sprintf("%s = %s(%d)", i.Index, i.Kind, i.Left)

	case Arg:
		return fmt.Sprintf("%s = Arg(%d)", i.Index, i.Literal.(int))

	case Bool:
		return fmt.Sprintf("%s = Bool(%t)", i.Index, i.Literal.(bool))

	case I64:
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

type InstructionKind int8

const (
	Undefined InstructionKind = iota

	// Three address
	Add
	Sub
	Mul
	Quo
	Equals
	NotEquals
	Move
	IfTrueGoto
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
	Function
	Arg
	Call
)
