package ir

import (
	"fmt"
	"strings"
)

//go:generate stringer -type=InstructionKind

type Program struct {
	Funcs []Procedure
	Names map[string]int
}

func (p *Program) AppendProcedure(fn Procedure) {
	p.Names[fn.Name] = len(p.Funcs)
	p.Funcs = append(p.Funcs, fn)
}

func (p Program) Get(idx int) *Procedure {
	if idx >= len(p.Funcs) {
		return nil
	}
	return &p.Funcs[idx]
}

func (p Program) Lookup(name string) *Procedure {
	if idx, ok := p.Names[name]; ok {
		return p.Get(idx)
	}
	return nil
}

func (p Program) String() string {
	sb := strings.Builder{}
	for _, procedure := range p.Funcs {
		sb.WriteString(fmt.Sprintf("%s()\n", procedure.Name))
		sb.WriteString(procedure.String())
		sb.WriteRune('\n')
	}
	return sb.String()
}

type Procedure struct {
	Blocks []Block
	Name   string
	Names  map[string]int
}

func (p *Procedure) AppendBlock(block Block) {
	p.Blocks = append(p.Blocks, block)
	p.Names[block.Name] = block.Index
}

func (p Procedure) Get(idx int) *Block {
	if idx >= len(p.Blocks) {
		return nil
	}
	return &p.Blocks[idx]
}

func (p Procedure) Lookup(name string) *Block {
	if idx, ok := p.Names[name]; ok {
		return p.Get(idx)
	}
	return nil
}

func (p Procedure) String() string {
	sb := strings.Builder{}
	for _, block := range p.Blocks {
		sb.WriteString(fmt.Sprintf("::%s[%d]\n", block.Name, block.Index))
		for _, inst := range block.Instructions {
			sb.WriteString(inst.String())
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

func (p Procedure) Next(b *Block) *Block {
	return p.Get(b.Index + 1)
}

type Block struct {
	Index        int
	Name         string
	Offset       Assignment
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
	Literal interface{}
}

func (i *Instruction) String() string {
	switch i.Kind {
	case Add, Sub, Mul, Quo,
		Equals, NotEquals, Move, IfTrueGoto,
		And, Or:
		return fmt.Sprintf("%4s = %s(%s, %s)", i.Index, i.Kind, i.Left, i.Right)

	case Not:
		return fmt.Sprintf("%4s = %s(%s)", i.Index, i.Kind, i.Left)

	case Push:
		return fmt.Sprintf("%4s = Push(%s)", i.Index, i.Left)

	case Pop:
		return fmt.Sprintf("%4s = Pop()", i.Index)

	case Bool:
		return fmt.Sprintf("%4s = Bool(%t)", i.Index, i.Literal.(bool))

	case I64:
		return fmt.Sprintf("%4s = Int(%d)", i.Index, i.Literal.(int64))

	case Function:
		return fmt.Sprintf("%4s = Function(%s)", i.Index, i.Literal.(string))

	case Call:
		sb := strings.Builder{}
		sb.WriteString(fmt.Sprintf("%4s = Call(%s)", i.Index, i.Left))
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
	Mod
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

	Call
	Push
	Pop
)
