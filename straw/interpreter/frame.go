package interpreter

import (
	"github.com/yjp20/turtle/straw/kind"
	"github.com/yjp20/turtle/straw/token"
)

type Frame interface {
	Get(selector string) Object
	Set(selector string, value Object)
}

type FunctionFrame struct {
	Parent Frame
	Values map[string]Object
	Return Object
}

func NewFunctionFrame(parent Frame) *FunctionFrame {
	return &FunctionFrame{
		Parent: parent,
		Values: make(map[string]Object),
	}
}

func (f *FunctionFrame) Kind() kind.Kind { return kind.Frame }
func (f *FunctionFrame) Inspect() string { return "<frame>" }
func (f *FunctionFrame) Get(selector string) Object {
	if _, ok := f.Values[selector]; ok {
		return f.Values[selector]
	}
	if f.Parent != nil {
		return f.Parent.Get(selector)
	}
	return NULL
}
func (f *FunctionFrame) Set(selector string, obj Object) {
	f.Values[selector] = obj
}

func NewGlobalFrame(errors *token.ErrorList) *GlobalFrame {
	return &GlobalFrame{Imports: make([]string, 0), Errors: errors}
}

type GlobalFrame struct {
	Imports []string
	Errors  *token.ErrorList
}

func (f *GlobalFrame) Get(selector string) Object {
	switch selector {
	case "print":
		return &BuiltinFunction{Name: "print"}
	case "debug":
		return &BuiltinFunction{Name: "debug"}
	case "make":
		return &BuiltinFunction{Name: "make"}
	case "import":
		return &BuiltinFunction{Name: "import"}
	case "i32":
		return &Type{ObjectKind: kind.I32}
	case "i64":
		return &Type{ObjectKind: kind.I64}
	case "bool":
		return &Type{ObjectKind: kind.Bool}
	case "f64":
		return &Type{ObjectKind: kind.F64}
	case "any":
		return &Type{ObjectKind: kind.Any}
	case "array":
		return &Factory{
			Params:      []Field{{Name: "T", Type: Type{ObjectKind: kind.Type}}},
			ProductKind: kind.Array,
		}
	case "slice":
		return &Factory{
			Params:      []Field{{Name: "T", Type: Type{ObjectKind: kind.Type}}},
			ProductKind: kind.Slice,
		}
	}
	return NULL
}
func (f *GlobalFrame) Set(selector string, obj Object) {}
func (f *GlobalFrame) Import(name string)              {}
