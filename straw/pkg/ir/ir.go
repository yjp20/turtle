package ir

import (
	"fmt"
	"strings"
)

type Program struct {
	Procedures []*Proc
	Names      map[string]int
}

func (p *Program) AppendProcdeure(procedure *Proc) {
	procedure.Index = len(p.Procedures)
	p.Names[procedure.Name] = procedure.Index
	p.Procedures = append(p.Procedures, procedure)
}

func (p Program) Lookup(name string) *Proc {
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

type Proc struct {
	Index  int
	Name   string
	Blocks []*Block
	Names  map[string]int
}

func (p *Proc) AppendBlock(block *Block) {
	block.Index = len(p.Blocks)
	p.Names[block.Name] = block.Index
	p.Blocks = append(p.Blocks, block)
}

func (p Proc) Next(block *Block) *Block {
	if block.Index+1 == len(p.Blocks) {
		return nil
	}
	return p.Blocks[block.Index+1]
}

func (p Proc) String() string {
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
	Instructions []*Inst
	Map          map[Assignment]int
	Symbols      map[string]Assignment
	Predecesors  []*Block

	Sealed         bool
	IncompletePhis []Assignment
}

func (b Block) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %d :: ", b.Name, b.Index))
	for _, pred := range b.Predecesors {
		sb.WriteString(fmt.Sprintf("%d ", pred.Index))
	}
	if b.Sealed {
		sb.WriteString("[sealed]")
	}
	sb.WriteRune('\n')
	for _, inst := range b.Instructions {
		sb.WriteString(inst.String())
		sb.WriteRune('\n')
	}
	return sb.String()
}

func (b Block) Get(a Assignment) *Inst {
	return b.Instructions[b.Map[a]]
}

func (b *Block) AddPredecesor(block *Block) {
	b.Predecesors = append(b.Predecesors, block)
}
