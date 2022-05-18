package interpreter

import (
	"fmt"

	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/kind"
)

type Field struct {
	Name  string
	Type  Type
	Value Object
}

type Object interface {
	Kind() kind.Kind
	Inspect() string
}

type Null struct{}

func (n *Null) Kind() kind.Kind { return kind.Null }
func (n *Null) Inspect() string { return "NULL" }

type Default struct{}

func (d *Default) Kind() kind.Kind { return kind.Default }
func (d *Default) Inspect() string { return "<default>" }

type I32 struct{ Value int32 }

func (i *I32) Kind() kind.Kind { return kind.I32 }
func (i *I32) Inspect() string { return fmt.Sprintf("<i32 %d>", i.Value) }

type I64 struct{ Value int64 }

func (i *I64) Kind() kind.Kind { return kind.I64 }
func (i *I64) Inspect() string { return fmt.Sprintf("<i64 %d>", i.Value) }

type F64 struct{ Value float64 }

func (i *F64) Kind() kind.Kind { return kind.F64 }
func (i *F64) Inspect() string { return fmt.Sprintf("<f64 %f>", i.Value) }

type Bool struct{ IsTrue bool }

func (b *Bool) Kind() kind.Kind { return kind.Bool }
func (b *Bool) Inspect() string { return fmt.Sprintf("<bool %t>", b.IsTrue) }

type String struct{ Value string }

func (s *String) Kind() kind.Kind { return kind.String }
func (s *String) Inspect() string { return fmt.Sprintf("\"%s\"", s.Value) }

type Function struct {
	Name  string
	Args  []Field
	Body  ast.Statement
	Frame *FunctionFrame
}

func (f *Function) Kind() kind.Kind { return kind.Function }
func (f *Function) Inspect() string { return fmt.Sprintf("<function '%s'>", f.Name) }

type BuiltinFunction struct {
	Name string
}

func (pf *BuiltinFunction) Kind() kind.Kind { return kind.BuiltinFunction }
func (pf *BuiltinFunction) Inspect() string { return fmt.Sprintf("<builtin function '%s'>", pf.Name) }

type Tuple struct {
	Fields []Field
}

func (t *Tuple) Kind() kind.Kind { return kind.Tuple }
func (t *Tuple) Inspect() string {
	s := "<tuple ("
	for i, f := range t.Fields {
		s = s + fmt.Sprintf("%s:%v", f.Name, f.Value.Inspect())
		if i != len(t.Fields) - 1{
			s = s + ", "
		}
	}
	return s + ")>"
}

type Type struct {
	Name       string
	ObjectKind kind.Kind
	Spec       []Field
}

func (t *Type) Kind() kind.Kind { return kind.Type }
func (t *Type) Inspect() string {
	if t.Spec != nil {
		// TODO
		return fmt.Sprintf("<type '%s' of kind '%s'>", t.Name, t.ObjectKind.String())
	}
	return fmt.Sprintf("<type '%s' of kind '%s'>", t.Name, t.ObjectKind.String())
}

// Factory can configure either a function, a struct, or some builtin types like arrays, slices, etc.
type Factory struct {
	Params      []Field
	ProductKind kind.Kind
}

func (f *Factory) Kind() kind.Kind { return kind.Factory }
func (f *Factory) Inspect() string { return fmt.Sprintf("<factory of '%T'>", f.ProductKind.String()) }

type Array struct {
	Objects  []Object
	ItemType *Type
}

func (a *Array) Kind() kind.Kind { return kind.Array }
func (a *Array) Inspect() string {
	s := "<array ["
	for _, obj := range a.Objects {
		s += obj.Inspect()
	}
	s += "]>"
	return s
}

type Range struct {
	Start int64
	End   int64
}

func (r *Range) Kind() kind.Kind { return kind.Range }
func (r *Range) Inspect() string { return "range" }

var (
	NULL  Object = &Null{}
	TRUE         = &Bool{IsTrue: true}
	FALSE        = &Bool{IsTrue: false}
)
