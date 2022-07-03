package ir

import (
	"fmt"
	"strings"
)

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
