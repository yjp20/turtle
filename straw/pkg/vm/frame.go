package vm

import (
	"github.com/yjp20/turtle/straw/pkg/ir"
	"github.com/yjp20/turtle/straw/pkg/kind"
)

type Frame struct {
	parent    *Frame
	offset    ir.Assignment
	end       ir.Assignment
	registers []Object
}

func NewFrame(parent *Frame, offset ir.Assignment, count int) *Frame {
	return &Frame{
		parent:    parent,
		offset:    offset,
		end:       offset + ir.Assignment(count),
		registers: make([]Object, count),
	}
}

func (f *Frame) Kind() kind.Kind { return kind.Frame }
func (f *Frame) Inspect() string { return "<frame>" }
func (f *Frame) Get(a ir.Assignment) Object {
	if f.offset <= a && a < f.end {
		return f.registers[int(a-f.offset)]
	}
	if f.parent != nil {
		return f.parent.Get(a)
	}
	return NULL
}
func (f *Frame) Set(a ir.Assignment, obj Object) {
	if f.offset <= a && a < f.end {
		f.registers[int(a-f.offset)] = obj
	}
}
