package interpreter

import (
	"encoding/json"
	"fmt"

	"github.com/yjp20/turtle/straw/ast"
)

type Field struct {
	Name  string
	Type  Object
	Value Object
}

type Object interface {
	Get(string) Object
	Type() TypeKind
	Inspect() string
}

type Frame struct {
	Parent *Frame
	Values map[string]Object
}

func NewFrame(parent *Frame) *Frame {
	return &Frame{
		Parent: parent,
		Values: make(map[string]Object),
	}
}

func NewGlobalFrame() *Frame {
	return &Frame{Values: map[string]Object{
		"print":   &BuiltinFunction{Kind: "print"},
		"debug":   &BuiltinFunction{Kind: "debug"},
		"int32":   &Type{Name: "int32", Kind: TypeInt32},
		"int64":   &Type{Name: "int64", Kind: TypeInt64},
		"bool":    &Type{Name: "bool", Kind: TypeBool},
		"float64": &Type{Name: "float64", Kind: TypeFloat64},
		"any":     &Type{Name: "any", Kind: TypeAny},
		"array": &Factory{
			Params:     []Field{{Name: "T", Type: &Type{Kind: TypeType}}},
			Configures: &Type{Kind: TypeArray},
		},
	}}
}

func (f *Frame) Get(selector string) Object {
	if _, ok := f.Values[selector]; ok {
		return f.Values[selector]
	}
	if f.Parent != nil {
		return f.Parent.Get(selector)
	}
	return NULL
}
func (f *Frame) Set(selector string, obj Object) {
	f.Values[selector] = obj
}
func (f *Frame) Type() TypeKind  { return TypeFrame }
func (f *Frame) Inspect() string { b, _ := json.MarshalIndent(f, "", "| "); return string(b) }

type Null struct{}

func (n *Null) Get(string) Object { return NULL }
func (n *Null) Type() TypeKind    { return TypeNull }
func (n *Null) Inspect() string   { return "NULL" }

type Default struct{}

func (d *Default) Get(string) Object { return NULL }
func (d *Default) Type() TypeKind    { return TypeDefault }
func (d *Default) Inspect() string   { return "<default>" }

type Int32 struct{ Value int32 }

func (i *Int32) Get(string) Object { return NULL }
func (i *Int32) Type() TypeKind    { return TypeInt32 }
func (i *Int32) Inspect() string   { return fmt.Sprintf("<int32 %d>", i.Value) }

type Int64 struct{ Value int64 }

func (i *Int64) Get(string) Object { return NULL }
func (i *Int64) Type() TypeKind    { return TypeInt64 }
func (i *Int64) Inspect() string   { return fmt.Sprintf("<int64 %d>", i.Value) }

type Float64 struct{ Value float64 }

func (i *Float64) Get(string) Object { return NULL }
func (i *Float64) Type() TypeKind    { return TypeFloat64 }
func (i *Float64) Inspect() string   { return fmt.Sprintf("<float64 %f>", i.Value) }

type Bool struct{ Value bool }

func (b *Bool) Get(string) Object { return NULL }
func (b *Bool) Type() TypeKind    { return TypeBool }
func (b *Bool) Inspect() string   { return fmt.Sprintf("<bool %t>", b.Value) }

type String struct{ Value string }

func (s *String) Get(string) Object { return NULL }
func (s *String) Type() TypeKind    { return TypeString }
func (s *String) Inspect() string   { return fmt.Sprintf("\"%s\"", s.Value) }

type Function struct {
	Name string
	Args []Field
	Body ast.Statement
}

func (f *Function) Get(string) Object { return NULL }
func (f *Function) Type() TypeKind    { return TypeFunction }
func (f *Function) Inspect() string   { return fmt.Sprintf("<function '%s'>", f.Name) }

type BuiltinFunction struct {
	Kind string
}

func (pf *BuiltinFunction) Get(string) Object { return NULL }
func (pf *BuiltinFunction) Type() TypeKind    { return TypeBuiltinFunction }
func (pf *BuiltinFunction) Inspect() string   { return fmt.Sprintf("<builtin function '%s'>", pf.Kind) }

type Tuple struct {
	Fields []Field
}

func (t *Tuple) Get(string) Object { return NULL }
func (t *Tuple) Type() TypeKind    { return TypeTuple }
func (t *Tuple) Inspect() string {
	s := "<tuple ( "
	for _, f := range t.Fields {
		s = s + fmt.Sprintf("%s:%s ", f.Name, f.Value.Inspect())
	}
	return s + ")>"
}

type Type struct {
	Name string
	Kind TypeKind
	Spec []Field
}

func (t *Type) Get(string) Object { return NULL }
func (t *Type) Type() TypeKind    { return TypeType }
func (t *Type) Inspect() string {
	if t.Spec != nil {
		// TODO
		return fmt.Sprintf("<type '%s' of kind '%T'>", t.Name, t.Kind)
	}
	return fmt.Sprintf("<type '%s' of kind '%T'>", t.Name, t.Kind)
}

type Factory struct {
	Params     []Field
	Configures Object
}

func (f *Factory) Get(string) Object { return NULL }
func (f *Factory) Type() TypeKind    { return TypeFactory }
func (f *Factory) Inspect() string   { return fmt.Sprintf("<factory of '%s'>", f.Configures.Inspect()) }

var (
	NULL  Object = &Null{}
	TRUE         = &Bool{Value: true}
	FALSE        = &Bool{Value: false}
)
