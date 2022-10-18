package irgen

import (
	"fmt"
	"sort"

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
		counter: 1,
		program: ir.Program{Procedures: make([]*ir.Proc, 0), Names: make(map[string]int)},
		errors:  errors,
	}
}

func (g *Generator) NewProcedure(name string) *ir.Proc {
	procedure := &ir.Proc{
		Name:   name,
		Blocks: make([]*ir.Block, 0),
		Names:  make(map[string]int),
	}
	g.program.AppendProcdeure(procedure)
	return procedure
}

func (g *Generator) NewBlock(name string, procedure *ir.Proc, predecesors []*ir.Block, sealed bool) *ir.Block {
	block := &ir.Block{
		Name:         name,
		Instructions: make([]*ir.Inst, 0),
		Map:          make(map[ir.Assignment]int),
		Symbols:      make(map[string]ir.Assignment),
		Predecesors:  predecesors,
		Sealed:       sealed,
	}
	procedure.AppendBlock(block)
	return block
}

func (g *Generator) GenerateBlock(name string, procedure *ir.Proc, predecesors []*ir.Block, sealed bool, n ast.Node) (ir.Assignment, *ir.Block, *ir.Block) {
	newBlock := g.NewBlock(name, procedure, predecesors, sealed)
	resultA, newBlockEnd := g.generate(n, procedure, newBlock)
	return resultA, newBlock, newBlockEnd
}

func (g *Generator) Generate(n ast.Node) ir.Program {
	g.generate(n, g.NewProcedure("_init"), nil)

	// Post-process outputted ir programs
	indexMap := map[ir.Assignment]ir.Assignment{}
	ct := ir.Assignment(1)
	for _, proc := range g.program.Procedures {
		for _, block := range proc.Blocks {
			// First, move phi nodes to the start of each block
			sort.SliceStable(block.Instructions, func(a, b int) bool {
				return block.Instructions[a].Kind == ir.Phi
			})
			// Second, relabel each node to linearize instructions
			for _, inst := range block.Instructions {
				indexMap[inst.Index] = ct
				ct += 1
			}
		}
	}

	// Apply relabel of nodes
	for _, proc := range g.program.Procedures {
		for _, block := range proc.Blocks {
			for _, inst := range block.Instructions {
				inst.Index = indexMap[inst.Index]
				if inst.Left != 0 {
					inst.Left = indexMap[inst.Left]
				}
				if inst.Right != 0 {
					inst.Right = indexMap[inst.Right]
				}
				if inst.Kind == ir.Phi {
					for i := range inst.Literal.([]ir.PhiLiteral) {
						inst.Literal.([]ir.PhiLiteral)[i].Assignment = indexMap[inst.Literal.([]ir.PhiLiteral)[i].Assignment]
					}
				}
			}
		}
	}

	return g.program
}

func (g *Generator) generate(node ast.Node, procedure *ir.Proc, block *ir.Block) (ir.Assignment, *ir.Block) {
	if block == nil {
		block = g.NewBlock("_init", procedure, []*ir.Block{}, true)
	}
	var a ir.Assignment
	switch node := node.(type) {
	case *ast.Program:
		for _, stmt := range node.Nodes {
			a, block = g.generate(stmt, procedure, block)
		}
		g.insertInstruction(block, ir.Inst{
			Kind: ir.End,
			Left: a,
		})

	case *ast.Block:
		for _, n := range node.Nodes {
			a, block = g.generate(n, procedure, block)
		}

	case *ast.Assign:
		switch left := node.Left.(type) {
		case *ast.Identifier:
			a, block = g.generate(node.Right, procedure, block)
			block.Symbols[left.Value] = a
		}

	case *ast.Return:
		a, block = g.generate(node.Body, procedure, block)
		g.insertInstruction(block, ir.Inst{
			Kind: ir.Ret,
			Left: a,
		})

	case *ast.Tuple:
		fields := make([]ir.Field, 0)
		for idx, n := range node.Nodes {
			switch n := n.(type) {
			case *ast.Assign:
				var ra ir.Assignment
				ra, block = g.generate(n.Right, procedure, block)
				fields = append(fields, ir.Field{
					Name:  n.Left.(*ast.Identifier).Value,
					Value: ra,
				})
			case *ast.As:
				fields = append(fields, ir.Field{
					Name:  n.Node.(*ast.Identifier).Value,
					Value: -1,
				})
			default:
				var ra ir.Assignment
				ra, block = g.generate(n, procedure, block)
				fields = append(fields, ir.Field{
					Name:  fmt.Sprintf("%d", idx),
					Value: ra,
				})
			}
		}
		a = g.insertInstruction(block, ir.Inst{
			Kind:    ir.ConstructTuple,
			Literal: fields,
		})

	case *ast.If:
		// The following assignments map to the following block structure
		// if (_cond_) {
		//   _true_
		// } else {
		//   _false_
		// }
		var condA, trueA, falseA ir.Assignment
		condA, block = g.generate(node.Condition, procedure, block)

		// The following instructions correspond to the header
		// block(inherited):
		//   ...
		//   %0 = cond
		//   %1 = not(%0)
		//   %2 = goto_if(%1, false)
		notA := g.insertInstruction(block, ir.Inst{
			Kind: ir.Not,
			Type: ir.Type{Kind: kind.Bool},
			Left: condA,
		})
		gotoIfA := g.insertInstruction(block, ir.Inst{
			Kind: ir.GotoIf,
			Type: ir.Type{Kind: kind.None},
			Left: notA,
		})

		// Generate the true, false, and next blocks which look like
		// true:
		//   %3 = ...
		//   goto next
		// false:
		//   %4 = ...
		// next:
		//
		trueA, trueBlock, trueBlockEnd := g.GenerateBlock("true", procedure, []*ir.Block{block}, true, node.TrueBody)
		gotoA := g.insertInstruction(trueBlockEnd, ir.Inst{
			Kind: ir.Goto,
			Type: ir.Type{Kind: kind.None},
		})
		goto_ := trueBlockEnd.Get(gotoA)

		falseA, falseBlock, falseBlockEnd := g.GenerateBlock("false", procedure, []*ir.Block{block}, true, node.FalseBody)
		block.Get(gotoIfA).Literal = falseBlock.Index

		nextBlock := g.NewBlock("next", procedure, []*ir.Block{trueBlockEnd, falseBlockEnd}, true)
		a = g.insertInstruction(nextBlock, ir.Inst{
			Kind:    ir.Phi,
			Type:    ir.Type{Kind: kind.None},
			Literal: []ir.PhiLiteral{{trueBlock.Index, trueA}, {falseBlock.Index, falseA}},
		})
		goto_.Literal = nextBlock.Index

		block = nextBlock

	case *ast.For:
		if clause, ok := node.Clause.(*ast.Each); ok {
			// FIXME: Currently only works for clauses which work over a range. It
			// is probably a better fix in the long term to have a transformer that
			// maps this to a while loop in terms of an ast node, either in some
			// in-between phase if there is a need for many such tranformations
			// or just inlined in this function, and then call generate again on the
			// reconstructed node.

			l := clause.Left.(*ast.Identifier)
			r := clause.Right.(*ast.RangeLiteral)

			var la, ra ir.Assignment
			la, block = g.generate(r.Left, procedure, block)
			ra, block = g.generate(r.Right, procedure, block)
			onea := g.insertInstruction(block, ir.Inst{
				Kind:    ir.I64,
				Literal: int64(1),
			})
			if !r.LeftInclusive {
				la = g.insertInstruction(block, ir.Inst{
					Kind:  ir.Add,
					Left:  la,
					Right: onea,
				})
			}
			if r.RightInclusive {
				ra = g.insertInstruction(block, ir.Inst{
					Kind:  ir.Add,
					Left:  ra,
					Right: onea,
				})
			}

			block.Symbols[l.Value] = la

			headBlock := g.NewBlock("loop", procedure, []*ir.Block{block}, false)
			iterA := g.insertInstruction(headBlock, ir.Inst{
				Kind:   ir.Phi,
				Symbol: l.Value,
				Type:   ir.Type{Kind: kind.I64},
			})
			headBlock.Symbols[l.Value] = iterA
			iterInst := headBlock.Get(iterA)

			lessA := g.insertInstruction(headBlock, ir.Inst{
				Kind:  ir.Less,
				Type:  ir.Type{Kind: kind.Bool},
				Left:  iterA,
				Right: ra,
			})
			notA := g.insertInstruction(headBlock, ir.Inst{
				Kind: ir.Not,
				Type: ir.Type{Kind: kind.Bool},
				Left: lessA,
			})
			jumpA := g.insertInstruction(headBlock, ir.Inst{
				Kind:    ir.GotoIf,
				Type:    ir.Type{Kind: kind.I64},
				Left:    notA,
				Literal: headBlock.Index,
			})
			jumpInst := headBlock.Get(jumpA)

			bodyBlock := g.NewBlock("loop_body", procedure, []*ir.Block{headBlock}, false)
			_, bodyBlock = g.generate(node.Body, procedure, bodyBlock)

			endBlock := g.NewBlock("loop_end", procedure, []*ir.Block{bodyBlock}, true)
			oneA := g.insertInstruction(endBlock, ir.Inst{
				Kind:    ir.I64,
				Type:    ir.Type{Kind: kind.I64},
				Literal: int64(1),
			})
			endA := g.insertInstruction(endBlock, ir.Inst{
				Kind:  ir.Add,
				Type:  ir.Type{Kind: kind.I64},
				Left:  oneA,
				Right: iterA,
			})
			g.insertInstruction(endBlock, ir.Inst{
				Kind:    ir.Goto,
				Literal: headBlock.Index,
			})
			headBlock.AddPredecesor(endBlock)
			nextBlock := g.NewBlock("next", procedure, []*ir.Block{headBlock}, true)
			iterInst.Literal = []ir.PhiLiteral{{endBlock.Index, endA}, {block.Index, la}}
			jumpInst.Literal = nextBlock.Index
			g.sealBlock(block)
			g.sealBlock(headBlock)
			g.sealBlock(bodyBlock)
			g.sealBlock(endBlock)
			g.sealBlock(nextBlock)
			block = nextBlock
		}

	case *ast.Match:
		var na ir.Assignment
		na, block = g.generate(node.Node, procedure, block)

		blocks := make([]*ir.Block, len(node.Tuple.Nodes))
		phi := make([]ir.PhiLiteral, 0)
		for idx := range node.Tuple.Nodes {
			blocks[idx] = g.NewBlock(fmt.Sprintf("match_%d", idx), procedure, []*ir.Block{block}, true)
		}
		blockNext := g.NewBlock("match_next", procedure, []*ir.Block{block}, true)

		for idx, n := range node.Tuple.Nodes {
			if n, ok := n.(*ast.Assign); ok {
				var la, ea, ra ir.Assignment

				// Left side of assignment, then check if match
				la, block = g.generate(n.Left, procedure, block)
				ea = g.insertInstruction(block, ir.Inst{
					Kind:  ir.Equals,
					Left:  na,
					Right: la,
				})
				g.insertInstruction(block, ir.Inst{
					Kind:    ir.GotoIf,
					Left:    ea,
					Literal: blocks[idx].Index,
				})

				// Execute command if match, then go to next
				blockBodyIndex := blocks[idx].Index
				ra, blocks[idx] = g.generate(n.Right, procedure, blocks[idx])
				g.insertInstruction(blocks[idx], ir.Inst{
					Kind:    ir.Goto,
					Literal: blockNext.Index,
				})

				blockNext.AddPredecesor(blocks[idx])
				phi = append(phi, ir.PhiLiteral{blockBodyIndex, ra})
			}
		}

		a = g.insertInstruction(blockNext, ir.Inst{
			Kind:    ir.Phi,
			Type:    ir.Type{Kind: kind.None},
			Literal: phi,
		})
		block = blockNext

	case *ast.ProcedureType:
		var ret ir.Assignment

		if node.ReturnType != nil {
			ret, block = g.generate(node.ReturnType, procedure, block)
		}

		a = g.insertInstruction(block, ir.Inst{
			Kind:    ir.ProcedureType,
			Type:    ir.Type{Kind: kind.Type},
			Static:  true,
			Literal: ret,
		})

	case *ast.ProcedureDefinition:
		newProcedure := g.NewProcedure("anon")
		newBlock := g.NewBlock("_start", newProcedure, []*ir.Block{}, true)

		a = g.insertInstruction(block, ir.Inst{
			Kind:    ir.ProcedureDefinition,
			Type:    ir.Type{Kind: kind.Function},
			Static:  true,
			Literal: newProcedure.Index,
		})

		if node.ProcedureType.Name != nil {
			block.Symbols[node.ProcedureType.Name.Value] = a
			newBlock.Symbols[node.ProcedureType.Name.Value] = a
		}

		for _, arg := range node.ProcedureType.Arguments {
			var t ir.Assignment
			t, newBlock = g.generate(arg.Type, newProcedure, newBlock)
			newBlock.Symbols[arg.Name] = g.insertInstruction(newBlock, ir.Inst{
				Kind: ir.Pop,
				Left: t,
			})
		}

		var returnBody ir.Assignment
		returnBody, newBlock = g.generate(node.Body, newProcedure, newBlock)
		_ = g.insertInstruction(newBlock, ir.Inst{
			Kind: ir.Ret,
			Left: returnBody,
		})

	case *ast.Call:
		var proc ir.Assignment
		proc, block = g.generate(node.Procedure, procedure, block)

		for _, node := range node.Arguments {
			var arg ir.Assignment
			arg, block = g.generate(node, procedure, block)
			_ = g.insertInstruction(block, ir.Inst{
				Kind: ir.Push,
				Left: arg,
			})
		}

		a = g.insertInstruction(block, ir.Inst{
			Kind: ir.Call,
			Left: proc,
		})

	case *ast.TrueLiteral:
		a = g.insertInstruction(block, ir.Inst{
			Type:    ir.Type{Kind: kind.Bool},
			Static:  true,
			Kind:    ir.Bool,
			Literal: true,
		})

	case *ast.FalseLiteral:
		a = g.insertInstruction(block, ir.Inst{
			Type:    ir.Type{Kind: kind.Bool},
			Static:  true,
			Kind:    ir.Bool,
			Literal: false,
		})

	case *ast.IntLiteral:
		a = g.insertInstruction(block, ir.Inst{
			Type:    ir.Type{Kind: kind.IntConstant},
			Static:  true,
			Kind:    ir.I64,
			Literal: node.Value,
		})

	case *ast.DefaultLiteral:
		a = g.insertInstruction(block, ir.Inst{
			Kind: ir.Default,
		})

	case *ast.Infix:
		var la, ra ir.Assignment
		la, block = g.generate(node.Left, procedure, block)
		ra, block = g.generate(node.Right, procedure, block)
		switch node.Operator {
		case token.ADD:
			a = g.insertInstruction(block, ir.Inst{
				Kind:  ir.Add,
				Left:  la,
				Right: ra,
			})
		case token.SUB:
			a = g.insertInstruction(block, ir.Inst{
				Kind:  ir.Sub,
				Left:  la,
				Right: ra,
			})
		case token.MUL:
			a = g.insertInstruction(block, ir.Inst{
				Kind:  ir.Mul,
				Left:  la,
				Right: ra,
			})
		case token.QUO:
			a = g.insertInstruction(block, ir.Inst{
				Kind:  ir.Quo,
				Left:  la,
				Right: ra,
			})
		case token.AND:
			a = g.insertInstruction(block, ir.Inst{
				Kind:  ir.And,
				Left:  la,
				Right: ra,
			})
		case token.EQUAL:
			a = g.insertInstruction(block, ir.Inst{
				Kind:  ir.Equals,
				Left:  la,
				Right: ra,
			})
		case token.NOT_EQUAL:
			a = g.insertInstruction(block, ir.Inst{
				Kind:  ir.NotEquals,
				Left:  la,
				Right: ra,
			})
		}

	case *ast.Prefix:
		var exprA ir.Assignment
		exprA, block = g.generate(node.Node, procedure, block)
		switch node.Operator {
		case token.NOT:
			a = g.insertInstruction(block, ir.Inst{
				Kind: ir.Not,
				Left: exprA,
			})
		}
	case *ast.Identifier:
		a = g.lookupSymbol(node.Value, block)

	default:
		fmt.Printf("NOT GENERATED: %T\n", node)
	}
	return a, block
}

func (g *Generator) resolvePhi(name string, phi *ir.Inst, block *ir.Block) ir.Assignment {
	for _, pred := range block.Predecesors {
		res := g.lookupSymbol(name, pred)
		if res != -1 {
			phi.Literal = append(phi.Literal.([]ir.PhiLiteral), ir.PhiLiteral{pred.Index, res})
		}
	}

	return block.Symbols[name]
}

func (g *Generator) lookupSymbol(name string, block *ir.Block) ir.Assignment {
	switch name {
	case "i64":
		//TODO
		return -1
	}

	if a, ok := block.Symbols[name]; ok {
		return a
	}

	block.Symbols[name] = g.insertInstruction(block, ir.Inst{
		Kind:    ir.Phi,
		Symbol:  name,
		Literal: []ir.PhiLiteral{},
	})
	phi := block.Get(block.Symbols[name])
	if !block.Sealed {
		block.IncompletePhis = append(block.IncompletePhis, block.Symbols[name])
		return block.Symbols[name]
	}

	return g.resolvePhi(name, phi, block)
}

func (g *Generator) sealBlock(block *ir.Block) {
	if block.Sealed {
		return
	}
	block.Sealed = true
	for _, a := range block.IncompletePhis {
		phi := block.Get(a)
		g.resolvePhi(phi.Symbol, phi, block)
	}
}

func (g *Generator) insertInstruction(block *ir.Block, inst ir.Inst) ir.Assignment {
	inst.Index = g.counter
	g.counter += 1
	block.Map[inst.Index] = len(block.Instructions)
	block.Instructions = append(block.Instructions, &inst)
	return inst.Index
}
