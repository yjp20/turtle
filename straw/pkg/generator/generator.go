package generator

import (
	"fmt"

	"github.com/yjp20/turtle/straw/pkg/ast"
	"github.com/yjp20/turtle/straw/pkg/ir"
	"github.com/yjp20/turtle/straw/pkg/kind"
	"github.com/yjp20/turtle/straw/pkg/token"
)

type Generator struct {
	insts   []ir.Instruction
	marks   map[ir.Assignment]string
	symbols map[string]ir.Assignment
	errors  *token.ErrorList
}

func NewGenerator(errors *token.ErrorList) *Generator {
	return &Generator{
		insts:   make([]ir.Instruction, 0),
		marks:   make(map[ir.Assignment]string),
		symbols: make(map[string]ir.Assignment),
		errors:  errors,
	}
}

func (g *Generator) Generate(n ast.Node) ir.Program {
	g.generate(n)
	program := ir.Program{
		Blocks: make([]ir.Block, 0),
		Names:  make(map[string]int),
	}
	block := ir.Block{
		Index:        0,
		Name:         "start",
		Instructions: make([]ir.Instruction, 0),
	}
	for idx, inst := range g.insts {
		a := ir.Assignment(idx)
		if name, ok := g.marks[a]; ok {
			program.AppendBlock(block)
			block = ir.Block{
				Index:        block.Index + 1,
				Name:         name,
				Instructions: make([]ir.Instruction, 0),
			}
		}
		block.Instructions = append(block.Instructions, inst)
	}
	program.AppendBlock(block)
	return program
}

func (g *Generator) generate(node ast.Node) ir.Assignment {
	var a ir.Assignment
	switch node := node.(type) {
	case *ast.Program:
		for _, stmt := range node.Statements {
			a = g.generate(stmt)
		}
	case *ast.Block:
		for _, stmt := range node.Statements {
			a = g.generate(stmt)
		}

	case *ast.AssignStatement:
		leftIdentifier := node.Left.(*ast.Identifier)
		a = g.generate(node.Right)
		g.symbols[leftIdentifier.Name] = a

	case *ast.ExpressionStatement:
		a = g.generate(node.Expression)

	case *ast.MatchExpression:
		// TODO: FIX
		ia := g.generate(node.Item)
		for i := range node.Bodies {
			if _, ok := node.Conditions[i].(*ast.DefaultLiteral); !ok {
				ca := g.generate(node.Conditions[i])
				g.insertInstruction(ir.Instruction{
					Kind:  ir.NotEquals,
					Type:  ir.Type{Kind: kind.Bool},
					Left:  ia,
					Right: ca,
				})
				g.insertInstruction(ir.Instruction{
					Kind:  ir.IfTrueGoto,
					Type:  ir.Type{Kind: kind.None},
					Right: -1,
				})
			}
			ba := g.generate(node.Bodies[i])
			g.insertInstruction(ir.Instruction{
				Kind:  ir.Move,
				Left:  a,
				Right: ba,
			})
		}

	case *ast.FunctionDefinition:
		a = g.insertInstruction(ir.Instruction{
			Kind:    ir.Function,
			Type:    ir.Type{Kind: kind.Function},
			Static:  true,
			Literal: "func_0",
		})
		for i, stmt := range node.Args.Statements {
			asExpr := stmt.(*ast.ExpressionStatement).Expression.(*ast.As)
			name := asExpr.Value.(*ast.Identifier).Name
			g.symbols[name] = g.insertInstruction(ir.Instruction{
				Kind:    ir.Arg,
				Type:    g.lookupType(asExpr.Type.(*ast.Identifier).Name),
				Literal: i,
			})
		}
		g.generate(node.Body)

	case *ast.CallExpression:
		exprs := make([]ir.Assignment, len(node.Expressions))
		for i, stmt := range node.Expressions {
			exprs[i] = g.generate(stmt)
		}
		g.insertInstruction(ir.Instruction{
			Kind:  ir.Call,
			Left:  exprs[0],
			Extra: exprs[1:],
		})

	case *ast.TrueLiteral:
		a = g.insertInstruction(ir.Instruction{
			Type:    ir.Type{Kind: kind.Bool},
			Static:  true,
			Kind:    ir.Bool,
			Literal: true,
		})

	case *ast.FalseLiteral:
		a = g.insertInstruction(ir.Instruction{
			Type:    ir.Type{Kind: kind.Bool},
			Static:  true,
			Kind:    ir.Bool,
			Literal: false,
		})

	case *ast.IntLiteral:
		a = g.insertInstruction(ir.Instruction{
			Type:    ir.Type{Kind: kind.IntConstant},
			Static:  true,
			Kind:    ir.I64,
			Literal: node.Value,
		})

	case *ast.Infix:
		la := g.generate(node.Left)
		ra := g.generate(node.Right)
		switch node.Operator {
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
		case token.AND:
			a = g.insertInstruction(ir.Instruction{
				Kind:  ir.And,
				Left:  la,
				Right: ra,
			})
		case token.EQUAL:
			a = g.insertInstruction(ir.Instruction{
				Kind:  ir.Equals,
				Left:  la,
				Right: ra,
			})
		case token.NOT_EQUAL:
			a = g.insertInstruction(ir.Instruction{
				Kind:  ir.NotEquals,
				Left:  la,
				Right: ra,
			})
		}

	case *ast.Prefix:
		ea := g.generate(node.Expression)
		switch node.Operator {
		case token.NOT:
			a = g.insertInstruction(ir.Instruction{
				Kind: ir.Not,
				Left: ea,
			})
		}
	case *ast.Identifier:
		a = g.lookupSymbol(node.Name)
	default:
		fmt.Printf("NOT GENERATED: %T\n", node)
	}
	return a
}

func (g *Generator) lookupSymbol(name string) ir.Assignment {
	if a, ok := g.symbols[name]; ok {
		return a
	} else {
		return -1
	}
}

func (g *Generator) lookupType(name string) ir.Type {
	switch name {
	case "i64":
		return ir.Type{Kind: kind.I64}
	}
	return ir.Type{Kind: kind.Unresolved}
}

func (g *Generator) unionType(a, b ir.Assignment) {

}

func (g *Generator) insertInstruction(inst ir.Instruction) ir.Assignment {
	inst.Index = ir.Assignment(len(g.insts))
	g.insts = append(g.insts, inst)
	return inst.Index
}

func (g *Generator) markBlock(a ir.Assignment) {

}
