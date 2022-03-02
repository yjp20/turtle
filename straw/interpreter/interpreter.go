package interpreter

import (
	"fmt"

	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/token"
)

// Basic tree walk interpreter implementation

func Eval(node ast.Node, env *Frame) Object {
	if env.Return != nil {
		return NULL
	}
	// fmt.Printf("EVAL: %T\n", node)
	switch e := node.(type) {
	case *ast.Program:
		return evalStatements(e.Statements, env)
	case *ast.Block:
		return evalStatements(e.Statements, env)
	case *ast.Tuple:
		return &Tuple{Fields: evalTuple(e, env)}

	case *ast.AssignStatement:
		assign(e.Left, Eval(e.Right, env), env)
	case *ast.ExpressionStatement:
		return Eval(e.Expression, env)
	case *ast.EmptyStatement:
		return NULL
	case *ast.ForStatement:
		return evalForStatement(e, env)
	case *ast.ReturnStatement:
		env.Return = Eval(e.Expression, env)
	case *ast.CallExpression:
		return evalCallExpression(e, env)
	case *ast.If:
		return evalIf(e, env)
	case *ast.Infix:
		return evalInfix(e, e.Operator, env)
	case *ast.DefaultLiteral:
		return &Default{}
	case *ast.Match:
		return evalMatch(e, env)
	case *ast.Indexor:
		return evalIndexor(e, env)

	case *ast.IntLiteral:
		return &I64{e.Value}
	case *ast.StringLiteral:
		return &String{e.Value}
	case *ast.FloatLiteral:
		return &F64{e.Value}
	case *ast.RangeLiteral:
		return evalRangeLiteral(e, env)
	case *ast.Identifier:
		return env.Get(e.Name)
	case *ast.FunctionDefinition:
		return evalFunctionDefinition(e, env)
	case *ast.TypeSpec:
		// TODO
	default:
		fmt.Printf("COULDN'T EVAL: %T\n", node)
	}
	return NULL
}

func evalStatements(stmts []ast.Statement, env *Frame) Object {
	var last Object = NULL
	for _, stmt := range stmts {
		switch stmt.(type) {
		case *ast.EmptyStatement:
		default:
			last = Eval(stmt, env)
		}
	}
	return last
}

func evalForStatement(fs *ast.ForStatement, env *Frame) Object {
	if len(fs.Clauses) == 1 {
		if s, ok := fs.Clauses[0].(*ast.EachStatement); ok {
			r := Eval(s.Right, env)
			switch rr := r.(type) {
			case *Range:
				for i := rr.Start; i < rr.End; i++ {
					assign(s.Left, &I64{Value: i}, env)
					Eval(fs.Expression, env)
				}
			}
		}
	}
	return NULL
}

func evalFunctionDefinition(fd *ast.FunctionDefinition, env *Frame) Object {
	f := &Function{Body: fd.Body}
	if fd.Identifier != nil {
		f.Name = fd.Identifier.Name
		env.Set(f.Name, f)
	}
	f.Args = evalSchema(fd.Args, env)
	return f
}

func evalIf(i *ast.If, env *Frame) Object {
	c := Eval(i.Conditional, env).(*Bool)
	if c.Value {
		return Eval(i.True, env)
	}
	if !c.Value && i.False != nil {
		return Eval(i.False, env)
	}
	return NULL
}

func evalMatch(m *ast.Match, env *Frame) Object {
	o := Eval(m.Item, env)
	for i := range m.Conditions {
		c := Eval(m.Conditions[i], env)
		if _, ok := c.(*Default); ok || c.Inspect() == o.Inspect() {
			return Eval(m.Bodies[i], env)
		}
	}
	return NULL
}

func evalCallExpression(c *ast.CallExpression, env *Frame) Object {
	objs := make([]Object, len(c.Expressions))
	for i, expr := range c.Expressions {
		objs[i] = Eval(expr, env)
	}

	frame := NewFrame(env)
	operator := objs[0]
	operands := objs[1:]
	switch e := operator.(type) {
	case *Function:
		for i, o := range operands {
			frame.Set(e.Args[i].Name, o)
		}
		last := Eval(e.Body, frame)
		if frame.Return != nil {
			return frame.Return
		}
		return last
	case *BuiltinFunction:
		switch e.Kind {
		case "print":
			for _, o := range operands {
				println(o.(*String).Value)
			}
		case "debug":
			for _, o := range operands {
				println(o.Inspect())
			}
		case "make":
			t := operands[0].(*Type)
			switch t.Kind {
			case TypeArray:
				return &Array{
					Objects:  make([]Object, operands[1].(*I64).Value),
					ItemType: t.Spec[0].Type,
				}
			}
		default:
			return NULL
		}
	}
	return NULL
}

func assign(left ast.Node, obj Object, env *Frame) {
	switch l := left.(type) {
	case *ast.Identifier:
		env.Set(l.Name, obj)
	case *ast.Indexor:
		x := Eval(l.Expression, env)
		t := Eval(l.Index, env).(*Tuple)
		switch k := x.(type) {
		case *Array:
			k.Objects[t.Fields[0].Value.(*I64).Value] = obj
		}
	}
}

func evalInfix(i *ast.Infix, operator token.Token, env *Frame) Object {
	l := Eval(i.Left, env)
	r := Eval(i.Right, env)
	switch l.(type) {
	case *I64:
		ll := l.(*I64)
		rr := r.(*I64)
		switch operator {
		case token.ADD:
			return &I64{Value: ll.Value + rr.Value}
		case token.MUL:
			return &I64{Value: ll.Value * rr.Value}
		case token.SUB:
			return &I64{Value: ll.Value - rr.Value}
		case token.LESS:
			return &Bool{Value: ll.Value < rr.Value}
		case token.LESS_EQUAL:
			return &Bool{Value: ll.Value <= rr.Value}
		case token.GREATER:
			return &Bool{Value: ll.Value > rr.Value}
		case token.GREATER_EQUAL:
			return &Bool{Value: ll.Value >= rr.Value}
		case token.EQUAL:
			return &Bool{Value: ll.Value == rr.Value}
		case token.NOT_EQUAL:
			return &Bool{Value: ll.Value != rr.Value}
		default:
			fmt.Printf("UNHANDLED INFIX OPERATOR: %T\n", operator)
			return NULL
		}
	case *F64:
		ll := l.(*F64)
		rr := r.(*F64)
		switch operator {
		case token.ADD:
			return &F64{Value: ll.Value + rr.Value}
		case token.MUL:
			return &F64{Value: ll.Value * rr.Value}
		case token.SUB:
			return &F64{Value: ll.Value - rr.Value}
		case token.LESS:
			return &Bool{Value: ll.Value < rr.Value}
		case token.LESS_EQUAL:
			return &Bool{Value: ll.Value <= rr.Value}
		case token.GREATER:
			return &Bool{Value: ll.Value > rr.Value}
		case token.GREATER_EQUAL:
			return &Bool{Value: ll.Value >= rr.Value}
		case token.EQUAL:
			return &Bool{Value: ll.Value == rr.Value}
		case token.NOT_EQUAL:
			return &Bool{Value: ll.Value != rr.Value}
		default:
			fmt.Printf("UNHANDLED INFIX OPERATOR: %T\n", operator)
			return NULL
		}
	}
	return NULL
}

func evalRangeLiteral(rl *ast.RangeLiteral, env *Frame) *Range {
	l := Eval(rl.Left, env).(*I64).Value
	r := Eval(rl.Right, env).(*I64).Value
	if !rl.LeftInclusive {
		l += 1
	}
	if rl.RightInclusive {
		r += 1
	}
	return &Range{Start: l, End: r}
}

func evalTuple(tuple *ast.Tuple, env *Frame) []Field {
	f := make([]Field, len(tuple.Statements))
	for i, stmt := range tuple.Statements {
		switch s := stmt.(type) {
		case *ast.AssignStatement:
		case *ast.ExpressionStatement:
			f[i] = Field{
				Name:  fmt.Sprintf("%d", i),
				Value: Eval(s, env),
			}
		default:
			fmt.Printf("UNHANDLED TUPLE STATEMENT: %T\n", s)
		}
	}
	return f
}

func evalSchema(tuple *ast.Tuple, env *Frame) []Field {
	fields := make([]Field, 0)
	for _, stmt := range tuple.Statements {
		switch s := stmt.(type) {
		case *ast.ExpressionStatement:
			if e, ok := s.Expression.(*ast.As); ok {
				fields = append(fields, Field{
					Name: e.Value.(*ast.Identifier).Name,
					Type: Eval(e.Type, env).(*Type),
				})
			}
		default:
			fmt.Printf("UNHANDLED STATEMENT IN SCHEMA: %T\n", s)
		}
	}
	return fields
}

func evalIndexor(i *ast.Indexor, env *Frame) Object {
	e := Eval(i.Expression, env)
	t := Eval(i.Index, env).(*Tuple)

	switch k := e.(type) {
	case *Array:
		return k.Objects[t.Fields[0].Value.(*I64).Value]
	case *Factory:
		switch k.Kind {
		case TypeArray:
			t := t.Fields[0].Value.(*Type)
			return &Type{
				Name: fmt.Sprintf("array[%s]", t.Name),
				Kind: TypeArray,
				Spec: []Field{{Name: "T", Type: &Type{Kind: TypeType}, Value: t}},
			}
		}
	}

	return NULL
}
