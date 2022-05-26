package interpreter

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/kind"
	"github.com/yjp20/turtle/straw/parser"
	"github.com/yjp20/turtle/straw/token"
)

// Basic tree walk interpreter implementation

func (gf *GlobalFrame) Eval(node ast.Node, env *FunctionFrame) Object {
	if env.Return != nil {
		return NULL
	}
	// fmt.Printf("EVAL: %T\n", node)
	switch e := node.(type) {
	case *ast.Program:
		return gf.evalStatements(e.Statements, env)
	case *ast.Block:
		return gf.evalStatements(e.Statements, env)
	case *ast.Tuple:
		return &Tuple{Fields: gf.evalTuple(e, env)}

	case *ast.AssignStatement:
		gf.assign(e.Left, gf.Eval(e.Right, env), env)
	case *ast.ExpressionStatement:
		return gf.Eval(e.Expression, env)
	case *ast.EmptyStatement:
		return NULL
	case *ast.ForStatement:
		return gf.evalForStatement(e, env)
	case *ast.ReturnStatement:
		env.Return = gf.Eval(e.Expression, env)
	case *ast.CallExpression:
		return gf.evalCallExpression(e, env)
	case *ast.IfExpression:
		return gf.evalIf(e, env)
	case *ast.Prefix:
		return gf.evalPrefix(e, e.Operator, env)
	case *ast.Infix:
		return gf.evalInfix(e, e.Operator, env)
	case *ast.MatchExpression:
		return gf.evalMatch(e, env)
	case *ast.IndexExpression:
		return gf.evalIndexor(e, env)

	case *ast.DefaultLiteral:
		return &Default{}
	case *ast.IntLiteral:
		return &I64{e.Value}
	case *ast.TrueLiteral:
		return &Bool{true}
	case *ast.FalseLiteral:
		return &Bool{false}
	case *ast.StringLiteral:
		return &String{e.Value}
	case *ast.FloatLiteral:
		return &F64{e.Value}
	case *ast.RangeLiteral:
		return gf.evalRangeLiteral(e, env)
	case *ast.Identifier:
		obj := env.Get(e.Name)
		if obj == NULL {
			gf.appendError("Couldn't find identifier", node.Pos(), node.End())
		}
		return obj
	case *ast.FunctionDefinition:
		return gf.evalFunctionDefinition(e, env)
	case *ast.ConstructExpression:
		return gf.evalConstructExpression(e, env)
	case *ast.TypeSpec:
		// TODO
	default:
		fmt.Printf("COULDN'T EVAL: %T\n", node)
	}
	return NULL
}

func (gf *GlobalFrame) evalStatements(stmts []ast.Statement, env *FunctionFrame) Object {
	var last Object = NULL
	for _, stmt := range stmts {
		switch stmt.(type) {
		case *ast.EmptyStatement:
		default:
			last = gf.Eval(stmt, env)
		}
	}
	return last
}

func (gf *GlobalFrame) evalForStatement(fs *ast.ForStatement, env *FunctionFrame) Object {
	if len(fs.Clauses) == 1 {
		if s, ok := fs.Clauses[0].(*ast.EachStatement); ok {
			r := gf.Eval(s.Right, env)
			switch rr := r.(type) {
			case *Range:
				for i := rr.Start; i < rr.End; i++ {
					gf.assign(s.Left, &I64{Value: i}, env)
					gf.Eval(fs.Expression, env)
				}
			case *Array:
				for _, item := range rr.Objects {
					gf.assign(s.Left, item, env)
					gf.Eval(fs.Expression, env)
				}
			}
		}
	}
	return NULL
}

func (gf *GlobalFrame) evalFunctionDefinition(fd *ast.FunctionDefinition, env *FunctionFrame) Object {
	f := &Function{Body: fd.Body, Frame: env}
	if fd.Identifier != nil {
		f.Name = fd.Identifier.Name
		env.Set(f.Name, f)
	}
	f.Args = gf.evalSchema(fd.Args, env)
	return f
}

func (gf *GlobalFrame) evalIf(expr *ast.IfExpression, env *FunctionFrame) Object {
	cond := gf.Eval(expr.Conditional, env).(*Bool)
	if cond.IsTrue {
		return gf.Eval(expr.True, env)
	}
	if !cond.IsTrue && expr.False != nil {
		return gf.Eval(expr.False, env)
	}
	return NULL
}

func (gf *GlobalFrame) evalMatch(m *ast.MatchExpression, env *FunctionFrame) Object {
	o := gf.Eval(m.Item, env)
	for i := range m.Conditions {
		c := gf.Eval(m.Conditions[i], env)
		if _, ok := c.(*Default); ok || c.Inspect() == o.Inspect() {
			return gf.Eval(m.Bodies[i], env)
		}
	}
	return NULL
}

func (gf *GlobalFrame) evalCallExpression(c *ast.CallExpression, env *FunctionFrame) Object {
	objs := make([]Object, len(c.Expressions))
	for i, expr := range c.Expressions {
		objs[i] = gf.Eval(expr, env)
	}

	operator := objs[0]
	operands := objs[1:]
	switch e := operator.(type) {
	case *Function:
		frame := NewFunctionFrame(e.Frame)
		for i := range e.Args {
			if len(operands) <= i {
				frame.Set(e.Args[i].Name, e.Args[i].Value)
			} else {
				frame.Set(e.Args[i].Name, operands[i])
			}
		}
		last := gf.Eval(e.Body, frame)
		if frame.Return != nil {
			return frame.Return
		}
		return last
	case *BuiltinFunction:
		switch e.Name {
		case "import":
			mf := NewFunctionFrame(gf)
			dir := operands[0].(*String).Value
			entries, _ := os.ReadDir(dir)
			for _, entry := range entries {
				if !entry.Type().IsRegular() || strings.HasPrefix(entry.Name(), ".st") {
					continue
				}

				path := filepath.Join(dir, entry.Name())
				file, _ := os.Open(path)
				b, _ := ioutil.ReadAll(file)
				pf := token.NewFile(b)
				ps := parser.NewParser(pf, gf.Errors)
				at := ps.ParseProgram()

				gf.Eval(at, mf)
			}
			return mf
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
			switch t.ObjectKind {
			case kind.Array:
				length := operands[1].(*I64).Value
				return &Array{
					Objects:  make([]Object, length),
					ItemType: &t.Spec[0].Type,
				}
			}
		default:
			return NULL
		}
	}
	return NULL
}

func (gf *GlobalFrame) evalConstructExpression(c *ast.ConstructExpression, env *FunctionFrame) Object {
	t := gf.Eval(c.Type, env).(*Type)
	switch t.ObjectKind {
	case kind.Array:
		objs := make([]Object, len(c.Value.Statements))
		for i, stmt := range c.Value.Statements {
			objs[i] = gf.Eval(stmt.(*ast.ExpressionStatement).Expression, env)
		}
		return &Array{
			Objects:  objs,
			ItemType: &t.Spec[0].Type,
		}

	}
	return NULL
}

func (gf *GlobalFrame) assign(left ast.Node, obj Object, env *FunctionFrame) {
	switch l := left.(type) {
	case *ast.Identifier:
		env.Set(l.Name, obj)
	case *ast.IndexExpression:
		x := gf.Eval(l.Expression, env)
		t := gf.Eval(l.Index, env).(*Tuple)
		switch k := x.(type) {
		case *Array:
			k.Objects[t.Fields[0].Value.(*I64).Value] = obj
		}
	case *ast.Tuple:
		for i, s := range l.Statements {
			switch st := s.(type) {
			case *ast.ExpressionStatement:
				gf.assign(st.Expression.(*ast.Identifier), obj.(*Tuple).Fields[i].Value, env)
			default:
				// TODO: ERORR?
			}
		}
	}
}

func (gf *GlobalFrame) evalPrefix(i *ast.Prefix, operator token.Token, env *FunctionFrame) Object {
	e := gf.Eval(i.Expression, env)
	switch e := e.(type) {
	case *Bool:
		if operator == token.NOT {
			return &Bool{!e.IsTrue}
		}
	}
	return NULL
}

func (gf *GlobalFrame) evalInfix(i *ast.Infix, operator token.Token, env *FunctionFrame) Object {
	l := gf.Eval(i.Left, env)
	r := gf.Eval(i.Right, env)
	switch l.(type) {
	case *Bool:
		ll := l.(*Bool)
		rr := r.(*Bool)
		switch operator {
		case token.AND:
			return &Bool{IsTrue: ll.IsTrue && rr.IsTrue}
		case token.OR:
			return &Bool{IsTrue: ll.IsTrue || rr.IsTrue}
		case token.XOR:
			return &Bool{IsTrue: (ll.IsTrue || rr.IsTrue) && !(ll.IsTrue && rr.IsTrue)}
		case token.EQUAL:
			return &Bool{IsTrue: ll.IsTrue == rr.IsTrue}
		case token.NOT_EQUAL:
			return &Bool{IsTrue: ll.IsTrue != rr.IsTrue}
		default:
			fmt.Printf("UNHANDLED INFIX OPERATOR: %T\n", operator)
			return NULL
		}
	case *I64:
		ll := l.(*I64)
		rr := r.(*I64)
		switch operator {
		case token.ADD:
			return &I64{Value: ll.Value + rr.Value}
		case token.MUL:
			return &I64{Value: ll.Value * rr.Value}
		case token.QUO:
			return &I64{Value: ll.Value / rr.Value}
		case token.SUB:
			return &I64{Value: ll.Value - rr.Value}
		case token.LESS:
			return &Bool{IsTrue: ll.Value < rr.Value}
		case token.LESS_EQUAL:
			return &Bool{IsTrue: ll.Value <= rr.Value}
		case token.GREATER:
			return &Bool{IsTrue: ll.Value > rr.Value}
		case token.GREATER_EQUAL:
			return &Bool{IsTrue: ll.Value >= rr.Value}
		case token.EQUAL:
			return &Bool{IsTrue: ll.Value == rr.Value}
		case token.NOT_EQUAL:
			return &Bool{IsTrue: ll.Value != rr.Value}
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
		case token.QUO:
			return &F64{Value: ll.Value / rr.Value}
		case token.SUB:
			return &F64{Value: ll.Value - rr.Value}
		case token.LESS:
			return &Bool{IsTrue: ll.Value < rr.Value}
		case token.LESS_EQUAL:
			return &Bool{IsTrue: ll.Value <= rr.Value}
		case token.GREATER:
			return &Bool{IsTrue: ll.Value > rr.Value}
		case token.GREATER_EQUAL:
			return &Bool{IsTrue: ll.Value >= rr.Value}
		case token.EQUAL:
			return &Bool{IsTrue: ll.Value == rr.Value}
		case token.NOT_EQUAL:
			return &Bool{IsTrue: ll.Value != rr.Value}
		default:
			fmt.Printf("UNHANDLED INFIX OPERATOR: %T\n", operator)
			return NULL
		}
	}
	return NULL
}

func (gf *GlobalFrame) evalRangeLiteral(rl *ast.RangeLiteral, env *FunctionFrame) *Range {
	l := gf.Eval(rl.Left, env).(*I64).Value
	r := gf.Eval(rl.Right, env).(*I64).Value
	if !rl.LeftInclusive {
		l += 1
	}
	if rl.RightInclusive {
		r += 1
	}
	return &Range{Start: l, End: r}
}

func (gf *GlobalFrame) evalTuple(tuple *ast.Tuple, env *FunctionFrame) []Field {
	f := make([]Field, len(tuple.Statements))
	for i, stmt := range tuple.Statements {
		switch s := stmt.(type) {
		case *ast.AssignStatement:
		case *ast.ExpressionStatement:
			o := gf.Eval(s, env)
			f[i] = Field{
				Name:  fmt.Sprintf("%d", i),
				Type:  Type{ObjectKind: o.Kind()},
				Value: o,
			}
		default:
			fmt.Printf("UNHANDLED TUPLE STATEMENT: %T\n", s)
		}
	}
	return f
}

func (gf *GlobalFrame) evalSchema(tuple *ast.Tuple, env *FunctionFrame) []Field {
	fields := make([]Field, 0)
	for _, stmt := range tuple.Statements {
		var (
			nameExpression ast.Expression
			typeExpression ast.Expression
			defaultValue   Object
		)

		switch s := stmt.(type) {
		case *ast.AssignStatement:
			nameExpression = s.Left.(*ast.As).Value
			typeExpression = s.Left.(*ast.As).Type
			defaultValue = gf.Eval(s.Right, env)
		case *ast.ExpressionStatement:
			nameExpression = s.Expression.(*ast.As).Value
			typeExpression = s.Expression.(*ast.As).Type
		default:
			fmt.Printf("UNHANDLED STATEMENT IN SCHEMA: %T\n", s)
		}

		var resolvedType *Type
		if ot, ok := gf.Eval(typeExpression, env).(*Type); ok {
			resolvedType = ot
		} else {
			// TODO Error
		}

		switch c := nameExpression.(type) {
		case *ast.Spread:
			fields = append(fields, Field{
				Name:  c.Expression.(*ast.Identifier).Name,
				Type:  *resolvedType,
				Value: defaultValue,
			})
		case *ast.Identifier:
			fields = append(fields, Field{
				Name:  c.Name,
				Type:  *resolvedType,
				Value: defaultValue,
			})
		default:
			// TODO: Error
		}
	}
	return fields
}

func (gf *GlobalFrame) evalIndexor(i *ast.IndexExpression, env *FunctionFrame) Object {
	operand := gf.Eval(i.Expression, env)

	switch operand := operand.(type) {
	case *FunctionFrame:
		return operand.Get(i.Index.(*ast.Identifier).Name)
	case *Array:
		idx := gf.Eval(i.Index, env).(*Tuple)
		return operand.Objects[idx.Fields[0].Value.(*I64).Value]
	case *Factory:
		idx := gf.Eval(i.Index, env).(*Tuple)
		switch operand.ProductKind {
		case kind.Array:
			t := idx.Fields[0].Value.(*Type)
			return &Type{
				Name:       fmt.Sprintf("array[%s]", t.Name),
				ObjectKind: kind.Array,
				Spec:       []Field{{Name: "T", Type: Type{ObjectKind: kind.Type}, Value: t}},
			}
		}
	}

	return NULL
}

func (gf *GlobalFrame) appendError(msg string, pos token.Pos, end token.Pos) {
	*gf.Errors = append(*gf.Errors, token.NewError("[interpreter] "+msg, pos, end))
}
