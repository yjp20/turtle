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
		errors:  errors,
	}
	fn := program.Lookup("_init")
	return s.eval(program, fn, fn.Get(0), env)
}

type state struct {
	stack   []Object
	stackIndex int
	errors  *token.ErrorList
	program ir.Program
}

func (state *state) push(obj Object) {
	if state.stackIndex == len(state.stack) {
		state.stack = append(state.stack, obj)
	} else {
		state.stack[state.stackIndex] = obj
	}
	state.stackIndex += 1
}

func (state *state) pop() Object {
	state.stackIndex -= 1
	return state.stack[state.stackIndex]
}

func (state *state) get(selector string) Object {
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

func (state *state) eval(program ir.Program, fn *ir.Procedure, block *ir.Block, parent *Frame) Object {
	if block == nil {
		return NULL
	}

	res := NULL
	env := NewFrame(parent, block.Offset, len(block.Instructions))

	for idx, inst := range block.Instructions {
		idx := ir.Assignment(idx)
		l := env.Get(inst.Left)
		r := env.Get(inst.Right)

		switch inst.Kind {
		case ir.I64:
			res = &I64{inst.Literal.(int64)}
		case ir.Bool:
			res = &Bool{inst.Literal.(bool)}

		case ir.Not:
			l := l.(*Bool)
			res = &Bool{!l.IsTrue}

		case ir.Quo:
			switch l := l.(type) {
			case *I64:
				res = &I64{l.Value / r.(*I64).Value}
			}
		case ir.Mod:
			switch l := l.(type) {
			case *I64:
				res = &I64{l.Value % r.(*I64).Value}
			}
		case ir.Mul:
			switch l := l.(type) {
			case *I64:
				res = &I64{l.Value * r.(*I64).Value}
			}
		case ir.Add:
			switch l := l.(type) {
			case *I64:
				res = &I64{l.Value + r.(*I64).Value}
			}
		case ir.Equals:
			res = &Bool{l.String() == r.String()}
		case ir.NotEquals:
			res = &Bool{l.String() != r.String()}
		case ir.And:
			l := l.(*Bool)
			r := r.(*Bool)
			res = &Bool{l.IsTrue && r.IsTrue}

		case ir.Push:
			state.push(l)

		case ir.Pop:
			res = state.pop()

		case ir.Function:
			res = &Function{Name: inst.Literal.(string), Frame: env}

		case ir.Call:
			l := l.(*Function)
			fn := program.Lookup(l.Name)
			res = state.eval(program, fn, fn.Get(0), l.Frame)

		default:
			state.appendError(fmt.Sprintf("COULDN'T EVAL: %s\n", inst.String()), 0, 0)
			res = NULL
		}
		// fmt.Printf("%d: %s\n", idx, res.String())
		env.Set(idx, res)
	}

	nxt := fn.Next(block)
	if nxt != nil {
		return state.eval(program, fn, nxt, env)
	}

	return res
}

func (state *state) appendError(msg string, pos token.Pos, end token.Pos) {
	*state.errors = append(*state.errors, token.NewError("[vm] "+msg, pos, end))
}
