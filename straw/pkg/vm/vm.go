package vm

import (
	"fmt"

	"github.com/yjp20/turtle/straw/pkg/ir"
	"github.com/yjp20/turtle/straw/pkg/kind"
	"github.com/yjp20/turtle/straw/pkg/token"
)

func Eval(program ir.Program, errors *token.ErrorList, env *Frame) Object {
	s := state{
		program: program,
		errors: errors,
	}
	return s.eval(program, program.Get(0), env)
}

type state struct {
	errors  *token.ErrorList
	program ir.Program
}

func (vm *state) get(selector string) Object {
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

func (state *state) eval(program ir.Program, block *ir.Block, parent *Frame) Object {
	if block == nil {
		return NULL
	}

	res := Object(nil)
	env := NewFrame(parent, block.Instructions[0].Index, len(block.Instructions))

	for idx, inst := range block.Instructions {
		idx := ir.Assignment(idx)
		l := env.Get(inst.Left)
		r := env.Get(inst.Right)
		switch inst.Kind {
		case ir.I64:
			res = &I64{inst.Literal.(int64)}
		case ir.Bool:
			res = &Bool{inst.Literal.(bool)}
		case ir.Equals:
			res = &Bool{l.String() == r.String()}
		case ir.NotEquals:
			res = &Bool{l.String() != r.String()}
		case ir.Not:
			l := l.(*Bool)
			res = &Bool{!l.IsTrue}
		case ir.And:
			l := l.(*Bool)
			r := r.(*Bool)
			res = &Bool{l.IsTrue && r.IsTrue}
		default:
			state.appendError(fmt.Sprintf("COULDN'T EVAL: %s\n", inst.String()), 0, 0)
			res = NULL
		}
		// fmt.Printf("%d: %s\n", idx, res.String())
		env.Set(idx, res)
	}

	nxt := program.Next(block)
	if nxt != nil {
		return state.eval(program, nxt, env)
	}

	return res
}

func (vm *state) appendError(msg string, pos token.Pos, end token.Pos) {
	*vm.errors = append(*vm.errors, token.NewError("[vm] "+msg, pos, end))
}
