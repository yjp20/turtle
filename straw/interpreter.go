package straw

import (
	"fmt"
	"reflect"

	"github.com/yjp20/straw/ast"
	"github.com/yjp20/straw/token"
)

// Basic tree walk interpreter implementation

func Eval(node ast.Node, env *Frame) Object {
	switch e := node.(type) {
	case *ast.Program:
		var last Object = NULL
		for _, stmt := range e.Statements {
			switch stmt.(type) {
			case *ast.EmptyStatement:
			default:
				last = Eval(stmt, env)
			}
		}
		return last
	case *ast.Block:
		var last Object = NULL
		for _, stmt := range e.Statements {
			switch stmt.(type) {
			case *ast.EmptyStatement:
			default:
				last = Eval(stmt, env)
			}
		}
		return last

	case *ast.AtomicExpressionList:
		objs := make([]Object, len(e.Expressions))
		for i, expr := range e.Expressions {
			objs[i] = Eval(expr, env)
		}
		if len(objs) == 0 {
			return NULL
		}
		if objs[0].Type() == FunctionType {
			return call(objs[0].(*Function), objs[1:], env)
		}

	case *ast.AssignStatement:
		evalAssignment(e, env)
	case *ast.ExpressionStatement:
		switch i := e.Expression.(type) {
		case *ast.FunctionDefinition:
			if i.Identifier != nil {
				*env.Get(i.Identifier.Name) = Eval(e.Expression, env)
				return NULL
			}
		}
		return Eval(e.Expression, env)

	case *ast.If:
		c := Eval(e.Conditional, env).(*Bool)
		if c.Value {
			return Eval(e.True, env)
		}
		if !c.Value && e.False != nil {
			return Eval(e.False, env)
		}
	case *ast.Infix:
		l := Eval(e.Left, env)
		r := Eval(e.Right, env)
		return evalInfix(l, r, e.Operator)

	case *ast.IntLiteral:
		return &Integer{e.Value}
	case *ast.FloatLiteral:
		return &Float{e.Value}
	case *ast.Identifier:
		return *env.Get(e.Name)
	case *ast.FunctionDefinition:
		return &Function{Block: e.Block}
	case *ast.TypeSpec:
		// TODO
	default:
		fmt.Println(reflect.TypeOf(e))
	}
	return NULL
}

func EvalLeft(node ast.Node, obj Object) *Object {
	switch e := node.(type) {
	case *ast.Identifier:
		return obj.Get(e.Name)
	}
	return &NULL
}

func call(f *Function, args []Object, env *Frame) Object {
	frame := NewFrame(env)
	return Eval(f.Block, frame)
}

func evalAssignment(e *ast.AssignStatement, env *Frame) {
	left := EvalLeft(e.Left, env)
	right := Eval(e.Right, env)
	*left = right
}

func evalInfix(l, r Object, operator token.Token) Object {
	switch l.(type) {
	case *Integer:
		ll := l.(*Integer)
		rr := r.(*Integer)
		switch operator {
		case token.ADD:
			return &Integer{Value: ll.Value + rr.Value}
		case token.MUL:
			return &Integer{Value: ll.Value * rr.Value}
		case token.SUB:
			return &Integer{Value: ll.Value - rr.Value}
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
			fmt.Println(operator.String())
			return NULL
		}
	case *Float:
		ll := l.(*Float)
		rr := r.(*Float)
		switch operator {
		case token.ADD:
			return &Float{Value: ll.Value + rr.Value}
		case token.MUL:
			return &Float{Value: ll.Value * rr.Value}
		case token.SUB:
			return &Float{Value: ll.Value - rr.Value}
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
			fmt.Println(operator.String())
			return NULL
		}
	}
	return NULL
}
