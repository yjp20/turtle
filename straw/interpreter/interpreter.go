package interpreter

import (
	"fmt"

	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/token"
)

// Basic tree walk interpreter implementation

func Eval(node ast.Node, env *Frame) Object {
	fmt.Printf("EVAL: %T\n", node)
	switch e := node.(type) {
	case *ast.Program:
		return evalStatements(e.Statements, env)
	case *ast.Block:
		return evalStatements(e.Statements, env)
	case *ast.Tuple:
		return evalTuple(e, env)

	case *ast.AssignStatement:
		evalAssignment(e, env)
	case *ast.ExpressionStatement:
		return Eval(e.Expression, env)
	case *ast.EmptyStatement:
		return NULL

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

	case *ast.IntLiteral:
		return &Int64{e.Value}
	case *ast.StringLiteral:
		return &String{e.Value}
	case *ast.FloatLiteral:
		return &Float64{e.Value}
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
		return Eval(e.Body, frame)
	case *BuiltinFunction:
		switch e.Kind {
		case "print":
			for _, o := range operands {
				print(o.(*String).Value)
			}
		case "debug":
			for _, o := range operands {
				print(o.Inspect())
			}
		default:
			return NULL
		}
	}
	return NULL
}

func evalAssignment(e *ast.AssignStatement, env *Frame) {
	switch l := e.Left.(type) {
	case *ast.Identifier:
		env.Set(l.Name, Eval(e.Right, env))
	}
}

func evalInfix(i *ast.Infix, operator token.Token, env *Frame) Object {
	l := Eval(i.Left, env)
	r := Eval(i.Right, env)
	switch l.(type) {
	case *Int64:
		ll := l.(*Int64)
		rr := r.(*Int64)
		switch operator {
		case token.ADD:
			return &Int64{Value: ll.Value + rr.Value}
		case token.MUL:
			return &Int64{Value: ll.Value * rr.Value}
		case token.SUB:
			return &Int64{Value: ll.Value - rr.Value}
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
	case *Float64:
		ll := l.(*Float64)
		rr := r.(*Float64)
		switch operator {
		case token.ADD:
			return &Float64{Value: ll.Value + rr.Value}
		case token.MUL:
			return &Float64{Value: ll.Value * rr.Value}
		case token.SUB:
			return &Float64{Value: ll.Value - rr.Value}
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

func evalTuple(tuple *ast.Tuple, env *Frame) *Tuple {
	t := Tuple{Fields: make([]Field, len(tuple.Statements))}
	for i, stmt := range tuple.Statements {
		switch s := stmt.(type) {
		case *ast.AssignStatement:
		case *ast.ExpressionStatement:
			t.Fields[i] = Field{
				Name:  fmt.Sprintf("%d", i),
				Value: Eval(s, env),
			}
		default:
			fmt.Printf("UNHANDLED TUPLE STATEMENT: %T\n", s)
		}
	}
	return &t
}

func evalSchema(tuple *ast.Tuple, env *Frame) []Field {
	fields := make([]Field, 0)
	for _, stmt := range tuple.Statements {
		switch s := stmt.(type) {
		case *ast.ExpressionStatement:
			if e, ok := s.Expression.(*ast.As); ok {
				fields = append(fields, Field{Name: e.Value.(*ast.Identifier).Name, Type: Eval(e.Type, env)})
			}
		default:
			fmt.Printf("UNHANDLED STATEMENT IN SCHEMA: %T\n", s)
		}
	}
	return fields
}

func evalIndexor(i *ast.Indexor, env *Frame) Object {
	e := Eval(i.Expression, env)
	c := Eval(i.Index, env)

	switch k := e.(type) {
	case *Array:

	case *Factory:
	}

	return NULL
}
