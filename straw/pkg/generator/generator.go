package generator

import (
	"fmt"

	"github.com/yjp20/turtle/straw/pkg/ast"
	"github.com/yjp20/turtle/straw/pkg/ir"
	"github.com/yjp20/turtle/straw/pkg/kind"
	"github.com/yjp20/turtle/straw/pkg/token"
)

type Generator struct {
	program ir.Program
	counter ir.Assignment
	errors  *token.ErrorList
}

func NewGenerator(errors *token.ErrorList) *Generator {
	return &Generator{
		counter: 0,
		program: ir.Program{Procedures: make([]*ir.Procedure, 0), Names: make(map[string]int)},
		errors:  errors,
	}
}

func (g *Generator) NewProcedure(name string) *ir.Procedure {
	procedure := &ir.Procedure{
		Name:   name,
		Blocks: make([]*ir.Block, 0),
		Names:  make(map[string]int),
	}
	g.program.AppendProcdeure(procedure)
	return procedure
}

func (g *Generator) NewBlock(name string, procedure *ir.Procedure) *ir.Block {
	block := &ir.Block{
		Name:         name,
		Instructions: make([]*ir.Instruction, 0),
		Map:          make(map[ir.Assignment]int),
		Symbols:      make(map[string]ir.Assignment),
	}
	procedure.AppendBlock(block)
	return block
}

func (g *Generator) Generate(n ast.Node) ir.Program {
	g.generate(n, g.NewProcedure("_init"), nil)
	return g.program
}

func (g *Generator) generate(node ast.Node, procedure *ir.Procedure, block *ir.Block) (ir.Assignment, *ir.Block) {
	if block == nil {
		block = g.NewBlock("_init", procedure)
	}
	var a ir.Assignment
	switch node := node.(type) {
	case *ast.Program:
		for _, stmt := range node.Statements {
			a, block = g.generate(stmt, procedure, block)
		}
		g.insertInstruction(block, ir.Instruction{
			Kind: ir.End,
			Left: a,
		})

	case *ast.Block:
		for _, stmt := range node.Statements {
			a, block = g.generate(stmt, procedure, block)
		}

	case *ast.AssignStatement:
		switch left := node.Left.(type) {
		case *ast.Identifier:
			a, block = g.generate(node.Right, procedure, block)
			block.Symbols[left.Name] = a
		}

	case *ast.ReturnStatement:
		a, block = g.generate(node.Expression, procedure, block)
		g.insertInstruction(block, ir.Instruction{
			Kind: ir.Ret,
			Left: a,
		})

	case *ast.IfExpression:
		var condA, trueResultA, falseResultA ir.Assignment
		condA, block = g.generate(node.Conditional, procedure, block)

		notA := g.insertInstruction(block, ir.Instruction{
			Kind: ir.Not,
			Type: ir.Type{Kind: kind.Bool},
			Left: condA,
		})
		jumpA := g.insertInstruction(block, ir.Instruction{
			Kind: ir.IfTrueGoto,
			Type: ir.Type{Kind: kind.None},
			Left: notA,
		})

		trueBlock := g.NewBlock("true", procedure)
		trueResultA, trueBlock = g.generate(node.True, procedure, trueBlock)

		if node.False == nil {
			//   %0 = cond
			//   %1 = not(%0)
			//   %2 = if_true_goto(%1, next)
			// true:
			//   %3 = ...
			// next:
			//   %4 = ...

			nextBlock := g.NewBlock("next", procedure)
			block.Get(jumpA).Literal = nextBlock.Index
			block = nextBlock
		} else {
			//   %0 = cond
			//   %1 = not(%0)
			//   %2 = if_true_goto(%1, false)
			// true:
			//   %3 = ...
			//   goto next
			// false:
			//   %4 = ...
			// next:
			//   %5 = ...

			skipA := g.insertInstruction(trueBlock, ir.Instruction{
				Kind: ir.Goto,
				Type: ir.Type{Kind: kind.None},
			})

			falseBlock := g.NewBlock("false", procedure)
			block.Get(jumpA).Literal = falseBlock.Index
			falseResultA, block = g.generate(node.False, procedure, falseBlock)

			nextBlock := g.NewBlock("next", procedure)
			trueBlock.Get(skipA).Literal = nextBlock.Index
			block = nextBlock
		}
		_ = trueResultA
		_ = falseResultA

	case *ast.ExpressionStatement:
		a, block = g.generate(node.Expression, procedure, block)

	case *ast.MatchExpression:
		var ia ir.Assignment
		ia, block = g.generate(node.Item, procedure, block)
		for i := range node.Bodies {
			if _, ok := node.Conditions[i].(*ast.DefaultLiteral); !ok {
				var ca ir.Assignment
				ca, block = g.generate(node.Conditions[i], procedure, block)
				g.insertInstruction(block, ir.Instruction{
					Kind:  ir.NotEquals,
					Type:  ir.Type{Kind: kind.Bool},
					Left:  ia,
					Right: ca,
				})
				g.insertInstruction(block, ir.Instruction{
					Kind: ir.IfTrueGoto,
					Type: ir.Type{Kind: kind.None},
				})
			}

			var ba ir.Assignment
			ba, block = g.generate(node.Bodies[i], procedure, block)
			g.insertInstruction(block, ir.Instruction{
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

		newProcedure := g.NewProcedure(name)
		newBlock := g.NewBlock("_start", newProcedure)

		a = g.insertInstruction(block, ir.Instruction{
			Kind:    ir.Function,
			Type:    ir.Type{Kind: kind.Function},
			Static:  true,
			Literal: newProcedure.Index,
		})
		if node.Identifier != nil {
			block.Symbols[name] = a
		}

		for idx := len(node.Args.Statements) - 1; idx >= 0; idx-- {
			stmt := node.Args.Statements[idx]
			switch stmt := stmt.(type) {
			case *ast.ExpressionStatement:
				asExpr := stmt.Expression.(*ast.As)
				name := asExpr.Value.(*ast.Identifier).Name
				newBlock.Symbols[name] = g.insertInstruction(newBlock, ir.Instruction{
					Kind: ir.Pop,
					Type: g.lookupType(asExpr.Type),
				})
			default:
				// TODO
			}
		}

		var returnBodyA ir.Assignment
		returnBodyA, newBlock = g.generate(node.Body, newProcedure, newBlock)
		g.insertInstruction(newBlock, ir.Instruction{
			Kind: ir.Ret,
			Left: returnBodyA,
		})

	case *ast.CallExpression:
		exprs := make([]ir.Assignment, len(node.Expressions))
		for i, stmt := range node.Expressions {
			exprs[i], block = g.generate(stmt, procedure, block)
			if i > 0 {
				g.insertInstruction(block, ir.Instruction{
					Kind: ir.Push,
					Left: exprs[i],
				})
			}
		}
		a = g.insertInstruction(block, ir.Instruction{
			Kind: ir.Call,
			Left: exprs[0],
		})

	case *ast.TrueLiteral:
		a = g.insertInstruction(block, ir.Instruction{
			Type:    ir.Type{Kind: kind.Bool},
			Static:  true,
			Kind:    ir.Bool,
			Literal: true,
		})

	case *ast.FalseLiteral:
		a = g.insertInstruction(block, ir.Instruction{
			Type:    ir.Type{Kind: kind.Bool},
			Static:  true,
			Kind:    ir.Bool,
			Literal: false,
		})

	case *ast.IntLiteral:
		a = g.insertInstruction(block, ir.Instruction{
			Type:    ir.Type{Kind: kind.IntConstant},
			Static:  true,
			Kind:    ir.I64,
			Literal: node.Value,
		})

	case *ast.Infix:
		var la, ra ir.Assignment
		la, block = g.generate(node.Left, procedure, block)
		ra, block = g.generate(node.Right, procedure, block)
		switch node.Operator {
		case token.ADD:
			a = g.insertInstruction(block, ir.Instruction{
				Kind:  ir.Add,
				Left:  la,
				Right: ra,
			})
		case token.SUB:
			a = g.insertInstruction(block, ir.Instruction{
				Kind:  ir.Sub,
				Left:  la,
				Right: ra,
			})
		case token.MUL:
			a = g.insertInstruction(block, ir.Instruction{
				Kind:  ir.Mul,
				Left:  la,
				Right: ra,
			})
		case token.QUO:
			a = g.insertInstruction(block, ir.Instruction{
				Kind:  ir.Quo,
				Left:  la,
				Right: ra,
			})
		case token.AND:
			a = g.insertInstruction(block, ir.Instruction{
				Kind:  ir.And,
				Left:  la,
				Right: ra,
			})
		case token.EQUAL:
			a = g.insertInstruction(block, ir.Instruction{
				Kind:  ir.Equals,
				Left:  la,
				Right: ra,
			})
		case token.NOT_EQUAL:
			a = g.insertInstruction(block, ir.Instruction{
				Kind:  ir.NotEquals,
				Left:  la,
				Right: ra,
			})
		}

	case *ast.Prefix:
		var exprA ir.Assignment
		exprA, block = g.generate(node.Expression, procedure, block)
		switch node.Operator {
		case token.NOT:
			a = g.insertInstruction(block, ir.Instruction{
				Kind: ir.Not,
				Left: exprA,
			})
		}
	case *ast.Identifier:
		a = g.lookupSymbol(node.Name, block)

	default:
		fmt.Printf("NOT GENERATED: %T\n", node)
	}
	return a, block
}

func (g *Generator) lookupSymbol(name string, block *ir.Block) ir.Assignment {
	if a, ok := block.Symbols[name]; ok {
		return a
	} else {
		return -1
	}
}

func (g *Generator) lookupType(node ast.Expression) ir.Type {
	switch node := node.(type) {
	case *ast.Identifier:
		switch node.Name {
		case "i64":
			return ir.Type{Kind: kind.I64}
		}
	}
	return ir.Type{Kind: kind.Unresolved}
}

func (g *Generator) unionType(a, b ir.Assignment) {

}

func (g *Generator) insertInstruction(block *ir.Block, inst ir.Instruction) ir.Assignment {
	inst.Index = g.counter
	g.counter += 1
	block.Map[inst.Index] = len(block.Instructions)
	block.Instructions = append(block.Instructions, &inst)
	return inst.Index
}
