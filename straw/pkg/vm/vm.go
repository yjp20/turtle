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
	return s.eval(program, proc, env)
}

type state struct {
	stack      []Object
	stackIndex int
	errors     *token.ErrorList
	program    ir.Program
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

func (state *state) eval(program ir.Program, proc *ir.Proc, env *Frame) Object {
	var (
		lastBlock = 0
		block     = proc.Blocks[0]
		res       = NULL
	)

	for block != nil {
		for _, inst := range block.Instructions {
			l := env.Get(inst.Left)
			r := env.Get(inst.Right)
			switch inst.Kind {
			case ir.I64:
				res = &I64{inst.Literal.(int64)}
			case ir.Bool:
				res = &Bool{inst.Literal.(bool)}
			case ir.Default:
				res = &Default{}

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
			case ir.Sub:
				switch l := l.(type) {
				case *I64:
					res = &I64{l.Value - r.(*I64).Value}
				}
			case ir.Equals:
				if l.Kind() == kind.Default || r.Kind() == kind.Default {
					res = &Bool{true}
				} else {
					res = &Bool{l.String() == r.String()}
				}
			case ir.NotEquals:
				res = &Bool{l.String() != r.String()}
			case ir.Less:
				if l, ok := l.(*I64); ok {
					if r, ok := r.(*I64); ok {
						res = &Bool{l.Value < r.Value}
					}
				} else {
					res = &Bool{l.String() < r.String()}
				}
			case ir.And:
				l := l.(*Bool)
				r := r.(*Bool)
				res = &Bool{l.IsTrue && r.IsTrue}
			case ir.ConstructTuple:
				args := make([]Field, 0)
				for _, arg := range inst.Literal.([]ir.Field) {
					args = append(args, Field{
						Name:  arg.Name,
						Value: env.Get(arg.Value),
					})
				}
				res = &Tuple{args}

			case ir.ProcedureType:
				// l := l.(*Tuple)
				r := r.(*Tuple)
				_ = env.Get(inst.Literal.(ir.Assignment))
				res = &Procedure{Name: "fibo", Args: r.Fields}
			case ir.LoadEnv:
				proctype := env.Get(inst.Literal.(ir.Assignment)).(*Procedure)
				for i := len(proctype.Args) - 1; i >= 0; i-- {
					env.SetVar(proctype.Args[i].Name, state.pop())
				}
				env.SetVar(proctype.Name, proctype)
			case ir.Env:
				res = env.GetVar(inst.Literal.(string))
			case ir.Push:
				state.push(l)
			case ir.Pop:
				res = state.pop()

			case ir.Ret, ir.End:
				res = l
				return res

			case ir.Phi:
				for _, k := range inst.Literal.([]ir.PhiLiteral) {
					if lastBlock == k.BlockIndex {
						res = env.Get(k.Assignment)
						break
					}
				}
				if res == NULL {
					// TODO ERROR
				}

			case ir.ProcedureDefinition:
				res = &Procedure{Index: inst.Literal.(int), Frame: env}

			case ir.Call:
				if l, ok := l.(*Procedure); ok {
					proc := program.Procedures[l.Index]
					newEnv := NewFrame(env)
					for _, field := range l.Args {
						newEnv.SetVar(field.Name, field.Value)
					}
					res = state.eval(program, proc, newEnv)
				}

			case ir.GotoIf:
				if good, ok := l.(*Bool); ok && good.IsTrue {
					lastBlock = block.Index
					block = proc.Blocks[inst.Literal.(int)]
					goto block_loop
				}

			case ir.Goto:
				lastBlock = block.Index
				block = proc.Blocks[inst.Literal.(int)]
				goto block_loop

			default:
				state.appendError(fmt.Sprintf("COULDN'T EVAL: %s\n", inst.String()), 0, 0)
				res = NULL
			}
			if res != nil {
				fmt.Printf("%s = %s\n", inst.Index, res.String())
			}
			env.Set(inst.Index, res)
		}
		lastBlock = block.Index
		block = proc.Next(block)
	block_loop:
	}
	return NULL
}

func (state *state) appendError(msg string, pos token.Pos, end token.Pos) {
	*state.errors = append(*state.errors, token.NewError("[vm] "+msg, pos, end))
}
