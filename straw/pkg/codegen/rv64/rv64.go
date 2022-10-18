package rv64

import (
	"fmt"

	"github.com/yjp20/turtle/straw/pkg/ir"
)

func Compile(program ir.Prog) []Instruction {
	insts := codegen{}
	for _, procedure := range program.Procedures {
		insts.compileProcedure(procedure)
	}
	return insts
}

type codegen []Instruction

func (c *codegen) compileProcedure(procedure *ir.Proc) {
	for _, block := range procedure.Blocks {
		c.compileBlock(block.Instructions)
	}
}

func (c *codegen) compileBlock(block []*ir.Inst) {
	am := addressMap{addresses: make([]addressDescriptor, len(block))}
	for i := 0; i < len(block); i++ {
		am.addresses[block[i].Left].LastUsed = ir.Assignment(i + 1)
		am.addresses[block[i].Right].LastUsed = ir.Assignment(i + 1)
	}

	for _, inst := range block {
		switch inst.Kind {
		case ir.ProcedureDefinition:
			fmt.Fprintf(&sb, "%s:\n", inst.Literal.(string))
		case ir.Pop:
			am.addresses[inst.Index].MemoryAddress = 4 + stackAddress(inst.Literal.(int))
		case ir.Add:
			dest, r1, r2 := am.getReg(&sb, inst)
			fmt.Fprintf(&sb, "  add %s, %s, %s\n", dest, r1, r2)
		case ir.Mul:
			dest, r1, r2 := am.getReg(&sb, inst)
			fmt.Fprintf(&sb, "  mul %s, %s, %s\n", dest, r1, r2)
		case ir.I64:
			dest, _, _ := am.getReg(&sb, inst)
			fmt.Fprintf(&sb, "  li %s, %d\n", dest, inst.Literal.(int64))
		default:
			fmt.Printf("NOT HANDLED IN COMPILE rv64: %s\n", inst.Kind)
		}
	}
}

func (c *codegen) getReg(am *addressMap, inst *ir.Inst) (dest registerAddress, r1 registerAddress, r2 registerAddress) {
	for i := registerAddress(5); i < 32; i++ {
		am.registers[i].Used = false
	}

	// If left is already register allocated, then assign that value
	if inst.Left != 0 {
		for i := registerAddress(5); i < 32; i++ {
			if am.registers[i].Assignment == inst.Left {
				am.registers[i].Used = true
				r1 = i
				break
			}
		}
	}
	// If right is already register allocated, then assign that value
	if inst.Right != 0 {
		for i := registerAddress(5); i < 32; i++ {
			if am.registers[i].Assignment == inst.Right {
				am.registers[i].Used = true
				r2 = i
				break
			}
		}
	}

	if inst.Left != 0 && r1 == 0 {
		r1 = c.getBest(am, inst.Index)
		fmt.Fprintf(sb, "  lw %s %d(sp)\n", r1, am.addresses[inst.Left].MemoryAddress)
	}
	if inst.Right != 0 && r2 == 0 {
		r2 = c.getBest(am, inst.Index)
		fmt.Fprintf(sb, "  lw %s %d(sp)\n", r2, am.addresses[inst.Right].MemoryAddress)
	}

	for i := registerAddress(5); i < 32; i++ {
		am.registers[i].Used = false
	}

	dest = c.getBest(am, inst.Index)

	if inst.Left != 0 {
		am.registers[r1].Assignment = inst.Left
	}
	if inst.Right != 0 {
		am.registers[r2].Assignment = inst.Right
	}

	am.registers[dest].Assignment = inst.Index
	return dest, r1, r2
}

func (c *codegen) getBest(am *addressMap, current ir.Assignment) registerAddress {
	best := 0x3FFFFFFF
	idx := registerAddress(0)

	for i := registerAddress(5); i < 10; i++ {
		score := 0
		if am.registers[i].Used {
			score += 0xFFFFFF
		}
		if am.registers[i].Assignment != 0 && am.addresses[am.registers[i].Assignment].LastUsed > current {
			score += 0xFFFF
		}
		// fmt.Fprintf(sb, "! i:%d score:%d a:%d current:%d lu:%d\n", i, score, am.registers[i].Assignment, current, am.addresses[am.registers[i].Assignment].LastUsed)
		if score < best {
			idx = i
			best = score
		}
	}

	// TODO: Handle if need spill
	am.registers[idx].Used = true
	return idx
}

func (c *codegen) spill(am *addressMap) {

}

func (c *codegen) Binary() {
	for _, i := range *c {
		i.Binary()
	}
}

type addressMap struct {
	addresses  []addressDescriptor
	registers  [32]registerDescriptor
	stackUsage int
}

type addressDescriptor struct {
	LastUsed        ir.Assignment
	Assignment      ir.Assignment
	RegisterAddress registerAddress
	MemoryAddress   stackAddress
}

type registerDescriptor struct {
	Used       bool
	Assignment ir.Assignment
}

type registerAddress int
type stackAddress int
