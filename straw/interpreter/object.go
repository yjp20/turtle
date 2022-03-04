package interpreter

import (
	"fmt"

	"github.com/yjp20/turtle/straw/ast"
)

type Field struct {
	Name  string
	Type  *Type
	Value Object
}

type Object interface {
	Type() TypeKind
	Inspect() string
}

type Null struct{}

func (n *Null) Type() TypeKind  { return TypeNull }
func (n *Null) Inspect() string { return "NULL" }

type Default struct{}

func (d *Default) Type() TypeKind  { return TypeDefault }
func (d *Default) Inspect() string { return "<default>" }

type I32 struct{ Value int32 }

func (i *I32) Type() TypeKind  { return TypeI32 }
func (i *I32) Inspect() string { return fmt.Sprintf("<int32 %d>", i.Value) }

type I64 struct{ Value int64 }

func (i *I64) Type() TypeKind  { return TypeI64 }
func (i *I64) Inspect() string { return fmt.Sprintf("<int64 %d>", i.Value) }

type F64 struct{ Value float64 }

func (i *F64) Type() TypeKind  { return TypeF64 }
func (i *F64) Inspect() string { return fmt.Sprintf("<float64 %f>", i.Value) }

type Bool struct{ Value bool }

func (b *Bool) Type() TypeKind  { return TypeBool }
func (b *Bool) Inspect() string { return fmt.Sprintf("<bool %t>", b.Value) }

type String struct{ Value string }

func (s *String) Type() TypeKind  { return TypeString }
func (s *String) Inspect() string { return fmt.Sprintf("\"%s\"", s.Value) }

type Function struct {
	Name string
	Args []Field
	Body ast.Statement
}

func (f *Function) Type() TypeKind  { return TypeFunction }
func (f *Function) Inspect() string { return fmt.Sprintf("<function '%s'>", f.Name) }

type BuiltinFunction struct {
	Kind string
}

func (pf *BuiltinFunction) Type() TypeKind  { return TypeBuiltinFunction }
func (pf *BuiltinFunction) Inspect() string { return fmt.Sprintf("<builtin function '%s'>", pf.Kind) }

type Tuple struct {
	Fields []Field
}

func (t *Tuple) Type() TypeKind { return TypeTuple }
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

func (t *Type) Type() TypeKind { return TypeType }
func (t *Type) Inspect() string {
	if t.Spec != nil {
		// TODO
		return fmt.Sprintf("<type '%s' of kind '%s'>", t.Name, t.Kind.String())
	}
	return fmt.Sprintf("<type '%s' of kind '%s'>", t.Name, t.Kind.String())
}

// Factory can configure either a function, a struct, or some builtin types like arrays, slices, etc.
type Factory struct {
	Params []Field
	Kind   TypeKind
}

func (f *Factory) Type() TypeKind  { return TypeFactory }
func (f *Factory) Inspect() string { return fmt.Sprintf("<factory of '%T'>", f.Kind.String()) }

type Array struct {
	Objects  []Object
	ItemType *Type
}

func (a *Array) Type() TypeKind  { return TypeArray }
func (a *Array) Inspect() string { return "<array[]>" }

type Range struct {
	Start int64
	End   int64
}

func (r *Range) Type() TypeKind  { return TypeRange }
func (r *Range) Inspect() string { return "range" }

var (
	NULL  Object = &Null{}
	TRUE         = &Bool{Value: true}
	FALSE        = &Bool{Value: false}
)
