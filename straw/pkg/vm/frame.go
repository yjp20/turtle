package vm

import (
	"github.com/yjp20/turtle/straw/pkg/ir"
	"github.com/yjp20/turtle/straw/pkg/kind"
)

type Frame struct {
	parent    *Frame
	registers map[ir.Assignment]Object
	variables map[string]Object
}

func NewFrame(parent *Frame) *Frame {
	return &Frame{
		parent:    parent,
		registers: make(map[ir.Assignment]Object),
		variables: make(map[string]Object),
	}
}

func (f *Frame) Kind() kind.Kind { return kind.Frame }
func (f *Frame) Inspect() string { return "<frame>" }
func (f *Frame) Get(a ir.Assignment) Object {
	if val, ok := f.registers[a]; ok {
		return val
	}
	if f.parent != nil {
		return f.parent.Get(a)
	}
	return NULL
}
func (f *Frame) Set(a ir.Assignment, obj Object) {
	f.registers[a] = obj
}
func (f *Frame) GetVar(name string) Object {
	return f.variables[name]
}
func (f *Frame) SetVar(name string, obj Object) {
	f.variables[name] = obj
}
