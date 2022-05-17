package parser

import (
	"fmt"

	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/ir"
	"github.com/yjp20/turtle/straw/kind"
	"github.com/yjp20/turtle/straw/token"
)

type Generator struct {
	Instructions []ir.Instruction
	symbols      map[string]ir.Assignment
	counter      ir.Assignment
}

func NewGenerator() *Generator {
	return &Generator{
		Instructions: make([]ir.Instruction, 0),
		symbols:      make(map[string]ir.Assignment),
		counter:      1,
	}
}

func (g *Generator) Generate(n ast.Node) ir.Assignment {
	var a ir.Assignment
	switch n := n.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			a = g.Generate(stmt)
		}
	case *ast.Block:
		for _, stmt := range n.Statements {
			a = g.Generate(stmt)
		}

	case *ast.AssignStatement:
		li := n.Left.(*ast.Identifier)
		a = g.Generate(n.Right)
		g.symbols[li.Name] = a
	case *ast.ExpressionStatement:
		a = g.Generate(n.Expression)
	case *ast.MatchExpression:
		// TODO: FIX
		ia := g.Generate(n.Item)
		a = g.counter
		for i := range n.Bodies {
			var gt ir.Assignment

			if _, ok := n.Conditions[i].(*ast.DefaultLiteral); !ok {
				ca := g.Generate(n.Conditions[i])
				neq := g.counter

				g.insertInstruction(ir.Instruction{
					Kind:       ir.NotEquals,
					Type: ir.Type{Kind: kind.Bool},
					Left:       ia,
					Right:      ca,
				})

				g.insertInstruction(ir.Instruction{
					Kind:       ir.IfTrueGoto,
					Type: ir.Type{Kind: kind.None},
					Left:       neq,
					Right:      -1,
				})
			}

			ba := g.Generate(n.Bodies[i])

			g.insertInstruction(ir.Instruction{
				Kind:  ir.Move,
				Left:  a,
				Right: ba,
			})

			g.insertInstruction(ir.Instruction{
				Kind:    ir.Label,
				Literal: "match_next",
			})

			g.Instructions[gt].Right = g.counter
			g.counter++
		}
	case *ast.FunctionDefinition:
		a = g.insertInstruction(ir.Instruction{
			Kind:       ir.Function,
			Type: ir.Type{Kind: kind.Function},
			Static:     true,
			Literal:    "func_0",
		})

		for i, stmt := range n.Args.Statements {
			expr := stmt.(*ast.ExpressionStatement).Expression.(*ast.As)
			name := expr.Value.(*ast.Identifier).Name
			g.symbols[name] = g.insertInstruction(ir.Instruction{
				Kind:       ir.Arg,
				Type: g.LookupType(expr.Type.(*ast.Identifier).Name),
				Literal:    i,
			})
		}

		g.Generate(n.Body)

	case *ast.CallExpression:
		exprs := make([]ir.Assignment, len(n.Expressions))
		for i, stmt := range n.Expressions {
			exprs[i] = g.Generate(stmt)
		}
		g.insertInstruction(ir.Instruction{
			Kind:  ir.Call,
			Left:  exprs[0],
			Extra: exprs[1:],
		})
	case *ast.IntLiteral:
		a = g.insertInstruction(ir.Instruction{
			Type: ir.Type{Kind: kind.IntConstant},
			Static:     true,
			Kind:       ir.Int,
			Literal:    n.Value,
		})
	case *ast.Infix:
		la := g.Generate(n.Left)
		ra := g.Generate(n.Right)
		switch n.Operator {
		case token.ADD:
			a = g.insertInstruction(ir.Instruction{
				Kind:  ir.Add,
				Left:  la,
				Right: ra,
			})
		case token.SUB:
			a = g.insertInstruction(ir.Instruction{
				Kind:  ir.Sub,
				Left:  la,
				Right: ra,
			})
		case token.MUL:
			a = g.insertInstruction(ir.Instruction{
				Kind:  ir.Mul,
				Left:  la,
				Right: ra,
			})
		}
	case *ast.Identifier:
		a = g.LookupSymbol(n.Name)
	default:
		fmt.Printf("NOT GENERATED: %T\n", n)
	}
	return a
}

func (g *Generator) Print() {
	for _, inst := range g.Instructions {
		println(inst.Print())
	}
}

func (g *Generator) LookupSymbol(name string) ir.Assignment {
	if a, ok := g.symbols[name]; ok {
		return a
	} else {
		return -1
	}
}

func (g *Generator) LookupType(name string) ir.Type {
	switch name {
	case "i64":
		return ir.Type{Kind: kind.I64}
	}
	return ir.Type{Kind: kind.Unresolved}
}

func (g *Generator) Union(a, b ir.Assignment) {

}

func (g *Generator) insertInstruction(inst ir.Instruction) ir.Assignment {
	a := g.counter
	inst.Index = a
	g.Instructions = append(g.Instructions, inst)
	g.counter += 1
	return a
}
