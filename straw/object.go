package straw

import (
	"encoding/json"
	"fmt"

	"github.com/yjp20/turtle/straw/ast"
)

type ObjectType int

const (
	IntegerType ObjectType = iota
	FloatType
	FrameType
	NullType
	BoolType
	FunctionType
)

type Object interface {
	Get(string) *Object
	Type() ObjectType
	Inspect() string
}

type Frame struct {
	Parent *Frame
	Values map[string]*Object
}

func NewFrame(parent *Frame) *Frame {
	return &Frame{
		Parent: parent,
		Values: make(map[string]*Object),
	}
}

func (f *Frame) Get(selector string) *Object {
	if _, ok := f.Values[selector]; !ok {
		var k = NULL
		f.Values[selector] = &k
	}
	return f.Values[selector]
}
func (f *Frame) Type() ObjectType { return FrameType }
func (f *Frame) Inspect() string  { b, _ := json.MarshalIndent(f, "", "| "); return string(b) }

type Integer struct{ Value int64 }

func (i *Integer) Get(string) *Object { return &NULL }
func (i *Integer) Type() ObjectType   { return IntegerType }
func (i *Integer) Inspect() string    { return fmt.Sprintf("%d", i.Value) }

type Float struct{ Value float64 }

func (i *Float) Get(string) *Object { return &NULL }
func (i *Float) Type() ObjectType   { return FloatType }
func (i *Float) Inspect() string    { return fmt.Sprintf("%f", i.Value) }

type Null struct{}

func (n *Null) Get(string) *Object { return &NULL }
func (n *Null) Type() ObjectType   { return NullType }
func (n *Null) Inspect() string    { return "null" }

type Bool struct{ Value bool }

func (b *Bool) Get(string) *Object { return &NULL }
func (b *Bool) Type() ObjectType   { return BoolType }
func (b *Bool) Inspect() string    { return fmt.Sprintf("%t", b.Value) }

type Field struct {
	Name string
	Type Object
}

type Function struct {
	Args       []Field
	Block      *ast.Block
}

func (f *Function) Get(string) *Object { return &NULL }
func (f *Function) Type() ObjectType   { return FunctionType }
func (f *Function) Inspect() string    { return "func" }

var (
	NULL  Object = &Null{}
	TRUE         = &Bool{Value: true}
	FALSE        = &Bool{Value: false}
)
