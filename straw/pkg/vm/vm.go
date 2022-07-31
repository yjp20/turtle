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
	proc := program.Lookup("_init")
	return s.eval(program, proc, proc.Blocks[0], env)
}

type state struct {
	stack      []Object
	stackIndex int
	errors     *token.ErrorList
	program    ir.Program
	lastBlock  string
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

func (state *state) eval(program ir.Program, proc *ir.Procedure, block *ir.Block, parent *Frame) Object {
	if block == nil {
		return NULL
	}

	res := NULL
	env := NewFrame(parent, len(block.Instructions))
	nxt := proc.Next(block)

	for _, inst := range block.Instructions {
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

		case ir.Ret, ir.End:
			res = l
			env.ret = res
			break

		case ir.Phi:
			for _, k := range inst.Literal.(ir.PhiLiteral) {
				if state.lastBlock == k.Block {
					res = env.Get(k.Assignment)
					break
				}
			}
			if res == NULL {
				// TODO ERROR
			}

		case ir.Function:
			res = &Function{Index: inst.Literal.(int), Frame: env}

		case ir.Call:
			if l, ok := l.(*Function); ok {
				proc := program.Procedures[l.Index]
				res = state.eval(program, proc, proc.Blocks[0], l.Frame)
				l.Frame.ret = nil
			}

		case ir.IfTrueGoto:
			if good, ok := l.(*Bool); ok && good.IsTrue {
				nxt = proc.Blocks[inst.Literal.(int)]
			}

		case ir.Goto:
			nxt = proc.Blocks[inst.Literal.(int)]

		default:
			state.appendError(fmt.Sprintf("COULDN'T EVAL: %s\n", inst.String()), 0, 0)
			res = NULL
		}
		env.Set(inst.Index, res)
	}

	state.lastBlock = block.Name

	if env.ret != nil && parent != nil {
		res = env.ret
		parent.ret = env.ret
		return res
	}

	if nxt != nil {
		return state.eval(program, proc, nxt, env)
	}

	return res
}

func (state *state) appendError(msg string, pos token.Pos, end token.Pos) {
	*state.errors = append(*state.errors, token.NewError("[vm] "+msg, pos, end))
}
