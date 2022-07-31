package ir

import (
	"fmt"
	"strings"
)

type Program struct {
	Procedures []*Procedure
	Names      map[string]int
}

func (p *Program) AppendProcdeure(procedure *Procedure) {
	procedure.Index = len(p.Procedures)
	p.Names[procedure.Name] = procedure.Index
	p.Procedures = append(p.Procedures, procedure)
}

func (p Program) Lookup(name string) *Procedure {
	if idx, ok := p.Names[name]; ok {
		return p.Procedures[idx]
	}
	return nil
}

func (p Program) String() string {
	sb := strings.Builder{}
	for _, p := range p.Procedures {
		sb.WriteString(p.String())
	}
	return sb.String()
}

type Procedure struct {
	Index  int
	Name   string
	Blocks []*Block
	Names  map[string]int
}

func (p *Procedure) AppendBlock(block *Block) {
	block.Index = len(p.Blocks)
	p.Names[block.Name] = block.Index
	p.Blocks = append(p.Blocks, block)
}

func (p Procedure) Next(block *Block) *Block {
	if block.Index+1 == len(p.Blocks) {
		return nil
	}
	return p.Blocks[block.Index+1]
}

func (p Procedure) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s()] %d\n", p.Name, p.Index))
	for _, block := range p.Blocks {
		sb.WriteString(block.String())
	}
	sb.WriteRune('\n')
	return sb.String()
}

type Block struct {
	Index        int
	Name         string
	Instructions []*Instruction
	Map          map[Assignment]int
	Symbols      map[string]Assignment
}

func (b Block) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("(%s) %d\n", b.Name, b.Index))
	for _, inst := range b.Instructions {
		sb.WriteString(inst.String())
		sb.WriteRune('\n')
	}
	return sb.String()
}

func (b Block) Get(a Assignment) *Instruction {
	return b.Instructions[b.Map[a]]
}
