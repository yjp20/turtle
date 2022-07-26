package generator

import (
	"fmt"

	"github.com/yjp20/turtle/straw/pkg/ast"
	"github.com/yjp20/turtle/straw/pkg/ir"
	"github.com/yjp20/turtle/straw/pkg/kind"
	"github.com/yjp20/turtle/straw/pkg/token"
)

type Generator struct {
	fns     []*Fn
	errors  *token.ErrorList
	fnIndex int
}

func NewGenerator(errors *token.ErrorList) *Generator {
	return &Generator{
		errors: errors,
	}
}

type Fn struct {
	name    string
	insts   []ir.Instruction
	marks   map[ir.Assignment]string
	symbols map[string]ir.Assignment
}

func (g *Generator) NewFn(name string) *Fn {
	fn := &Fn{
		name:    name,
		insts:   make([]ir.Instruction, 0),
		marks:   make(map[ir.Assignment]string),
		symbols: make(map[string]ir.Assignment),
	}
	g.fns = append(g.fns, fn)
	return fn
}

func (g *Generator) Generate(n ast.Node) ir.Program {
	init := g.NewFn("_init")
	g.generate(n, init)
	program := ir.Program{Funcs: make([]ir.Procedure, 0), Names: make(map[string]int)}
	for _, fn := range g.fns {
		procedure := ir.Procedure{
			Blocks: make([]ir.Block, 0),
			Names:  make(map[string]int),
			Name:   fn.name,
		}
		block := ir.Block{
			Index:        0,
			Name:         "_start",
			Offset:       0,
			Instructions: make([]ir.Instruction, 0),
		}
		for idx, inst := range fn.insts {
			a := ir.Assignment(idx)
			if name, ok := fn.marks[a]; ok {
				procedure.AppendBlock(block)
				block = ir.Block{
					Index:        block.Index + 1,
					Name:         name,
					Offset:       a,
					Instructions: make([]ir.Instruction, 0),
				}
			}
			block.Instructions = append(block.Instructions, inst)
		}
		procedure.AppendBlock(block)
		program.AppendProcedure(procedure)
	}
	return program
}

func (g *Generator) generate(node ast.Node, fn *Fn) ir.Assignment {
	var a ir.Assignment
	switch node := node.(type) {
	case *ast.Program:
		for _, stmt := range node.Statements {
			a = g.generate(stmt, fn)
		}
		fn.insertInstruction(ir.Instruction{
			Kind: ir.End,
			Left: a,
		})
	case *ast.Block:
		for _, stmt := range node.Statements {
			a = g.generate(stmt, fn)
		}

	case *ast.AssignStatement:
		switch left := node.Left.(type) {
		case *ast.Identifier:
			a = g.generate(node.Right, fn)
			fn.symbols[left.Name] = a
		}

	case *ast.IfExpression:
		ca := g.generate(node.Conditional, fn)
		na := fn.insertInstruction(ir.Instruction{
			Kind: ir.Not,
			Type: ir.Type{Kind: kind.Bool},
			Left: ca,
		})
		ja := fn.insertInstruction(ir.Instruction{
			Kind: ir.IfTrueGoto,
			Type: ir.Type{Kind: kind.None},
			Left: na,
		})

		tcond := fn.markNext(&g.fnIndex)
		tres := g.generate(node.True, fn)

		if node.False == nil {
			cont := fn.markNext(&g.fnIndex)
			fn.insts[ja].Literal = cont

			a = fn.insertInstruction(ir.Instruction{
				Kind:    ir.Phi,
				Type:    ir.Type{Kind: kind.None},
				Literal: ir.PhiLiteral{{tcond, tres}},
			})
		} else {
			ba := fn.insertInstruction(ir.Instruction{
				Kind: ir.Goto,
				Type: ir.Type{Kind: kind.None},
			})

			fcond := fn.markNext(&g.fnIndex)
			fres := g.generate(node.False, fn)

			cont := fn.markNext(&g.fnIndex)
			fn.insts[ja].Literal = fcond
			fn.insts[ba].Literal = cont

			a = fn.insertInstruction(ir.Instruction{
				Kind:    ir.Phi,
				Type:    ir.Type{Kind: kind.None},
				Literal: ir.PhiLiteral{{tcond, tres}, {fcond, fres}},
			})
		}

	case *ast.ExpressionStatement:
		a = g.generate(node.Expression, fn)

	case *ast.MatchExpression:
		// TODO: FIX
		ia := g.generate(node.Item, fn)
		for i := range node.Bodies {
			if _, ok := node.Conditions[i].(*ast.DefaultLiteral); !ok {
				ca := g.generate(node.Conditions[i], fn)
				fn.insertInstruction(ir.Instruction{
					Kind:  ir.NotEquals,
					Type:  ir.Type{Kind: kind.Bool},
					Left:  ia,
					Right: ca,
				})
				fn.insertInstruction(ir.Instruction{
					Kind: ir.IfTrueGoto,
					Type: ir.Type{Kind: kind.None},
				})
			}
			ba := g.generate(node.Bodies[i], fn)
			fn.insertInstruction(ir.Instruction{
				Kind:  ir.Move,
				Left:  a,
				Right: ba,
			})
		}

	case *ast.FunctionDefinition:
		name := "anon"
		if node.Identifier != nil {
			name = node.Identifier.Name
		}

		name = fmt.Sprintf("%s_%d", name, g.fnIndex)
		g.fnIndex += 1
		a = fn.insertInstruction(ir.Instruction{
			Kind:    ir.Function,
			Type:    ir.Type{Kind: kind.Function},
			Static:  true,
			Literal: name,
		})
		if node.Identifier != nil {
			fn.symbols[name] = a
		}
		newFn := g.NewFn(name)
		for idx := len(node.Args.Statements) - 1; idx >= 0; idx-- {
			stmt := node.Args.Statements[idx]
			switch stmt := stmt.(type) {
			case *ast.ExpressionStatement:
				asExpr := stmt.Expression.(*ast.As)
				name := asExpr.Value.(*ast.Identifier).Name
				newFn.symbols[name] = newFn.insertInstruction(ir.Instruction{
					Kind: ir.Pop,
					Type: fn.lookupType(asExpr.Type),
				})
			default:
				// TODO
			}
		}
		r := g.generate(node.Body, newFn)
		newFn.insertInstruction(ir.Instruction{
			Kind: ir.Ret,
			Left: r,
		})

	case *ast.CallExpression:
		exprs := make([]ir.Assignment, len(node.Expressions))
		for i, stmt := range node.Expressions {
			exprs[i] = g.generate(stmt, fn)
			if i > 0 {
				fn.insertInstruction(ir.Instruction{
					Kind: ir.Push,
					Left: exprs[i],
				})
			}
		}
		a = fn.insertInstruction(ir.Instruction{
			Kind: ir.Call,
			Left: exprs[0],
		})

	case *ast.TrueLiteral:
		a = fn.insertInstruction(ir.Instruction{
			Type:    ir.Type{Kind: kind.Bool},
			Static:  true,
			Kind:    ir.Bool,
			Literal: true,
		})

	case *ast.FalseLiteral:
		a = fn.insertInstruction(ir.Instruction{
			Type:    ir.Type{Kind: kind.Bool},
			Static:  true,
			Kind:    ir.Bool,
			Literal: false,
		})

	case *ast.IntLiteral:
		a = fn.insertInstruction(ir.Instruction{
			Type:    ir.Type{Kind: kind.IntConstant},
			Static:  true,
			Kind:    ir.I64,
			Literal: node.Value,
		})

	case *ast.Infix:
		la := g.generate(node.Left, fn)
		ra := g.generate(node.Right, fn)
		switch node.Operator {
		case token.ADD:
			a = fn.insertInstruction(ir.Instruction{
				Kind:  ir.Add,
				Left:  la,
				Right: ra,
			})
		case token.SUB:
			a = fn.insertInstruction(ir.Instruction{
				Kind:  ir.Sub,
				Left:  la,
				Right: ra,
			})
		case token.MUL:
			a = fn.insertInstruction(ir.Instruction{
				Kind:  ir.Mul,
				Left:  la,
				Right: ra,
			})
		case token.QUO:
			a = fn.insertInstruction(ir.Instruction{
				Kind:  ir.Quo,
				Left:  la,
				Right: ra,
			})
		case token.AND:
			a = fn.insertInstruction(ir.Instruction{
				Kind:  ir.And,
				Left:  la,
				Right: ra,
			})
		case token.EQUAL:
			a = fn.insertInstruction(ir.Instruction{
				Kind:  ir.Equals,
				Left:  la,
				Right: ra,
			})
		case token.NOT_EQUAL:
			a = fn.insertInstruction(ir.Instruction{
				Kind:  ir.NotEquals,
				Left:  la,
				Right: ra,
			})
		}

	case *ast.Prefix:
		ea := g.generate(node.Expression, fn)
		switch node.Operator {
		case token.NOT:
			a = fn.insertInstruction(ir.Instruction{
				Kind: ir.Not,
				Left: ea,
			})
		}
	case *ast.Identifier:
		a = fn.lookupSymbol(node.Name)
	default:
		fmt.Printf("NOT GENERATED: %T\n", node)
	}
	return a
}

func (fn *Fn) lookupSymbol(name string) ir.Assignment {
	if a, ok := fn.symbols[name]; ok {
		return a
	} else {
		return -1
	}
}

func (fn *Fn) lookupType(node ast.Expression) ir.Type {
	switch node := node.(type) {
	case *ast.Identifier:
		switch node.Name {
		case "i64":
			return ir.Type{Kind: kind.I64}
		}
	}
	return ir.Type{Kind: kind.Unresolved}
}

func (fn *Fn) unionType(a, b ir.Assignment) {

}

func (fn *Fn) insertInstruction(inst ir.Instruction) ir.Assignment {
	inst.Index = ir.Assignment(len(fn.insts))
	fn.insts = append(fn.insts, inst)
	return inst.Index
}

func (fn *Fn) markBlock(a ir.Assignment, name string) {
	fn.marks[a] = name
}

func (fn *Fn) markNext(index *int) string {
	*index += 1
	name := fmt.Sprintf("%d", *index)
	fn.marks[ir.Assignment(len(fn.insts))] = name
	return name
}
