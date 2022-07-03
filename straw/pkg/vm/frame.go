package vm

import (
	"github.com/yjp20/turtle/straw/pkg/ir"
	"github.com/yjp20/turtle/straw/pkg/kind"
)

type Frame struct {
	parent *Frame
	offset ir.Assignment
	end    ir.Assignment
	stack  []Object
	Return Object
}

func NewFrame(parent *Frame, offset ir.Assignment, count int) *Frame {
	return &Frame{
		parent: parent,
		offset: offset,
		end:    offset + ir.Assignment(count),
		stack:  make([]Object, count),
	}
}

func (f *Frame) Kind() kind.Kind { return kind.Frame }
func (f *Frame) Inspect() string { return "<frame>" }
func (f *Frame) Get(a ir.Assignment) Object {
	if f.offset <= a && a < f.end {
		return f.stack[int(a-f.offset)]
	}
	if f.parent != nil {
		return f.parent.Get(a)
	}
	return NULL
}
func (f *Frame) Set(a ir.Assignment, obj Object) {
	f.stack[int(a-f.offset)] = obj
}
