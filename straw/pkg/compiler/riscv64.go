package compiler

import (
	"fmt"
	"strings"

	"github.com/yjp20/turtle/straw/ir"
)

type AddressMap struct {
	addresses  []AddressDescriptor
	registers  [32]RegisterDescriptor
	stackUsage int
}

type AddressDescriptor struct {
	LastUsed        ir.Assignment
	Assignment      ir.Assignment
	RegisterAddress RegisterAddress
	MemoryAddress   StackAddress
}

type RegisterDescriptor struct {
	Used       bool
	Assignment ir.Assignment
}

type RegisterAddress int

func (ra RegisterAddress) String() string { return fmt.Sprintf("x%d", ra) }

type StackAddress int

func (am *AddressMap) GetReg(sb *strings.Builder, inst ir.Instruction) (dest RegisterAddress, r1 RegisterAddress, r2 RegisterAddress) {
	for i := RegisterAddress(5); i < 32; i++ {
		am.registers[i].Used = false
	}

	// If left is already register allocated, then assign that value
	if inst.Left != 0 {
		for i := RegisterAddress(5); i < 32; i++ {
			if am.registers[i].Assignment == inst.Left {
				am.registers[i].Used = true
				r1 = i
				break
			}
		}
	}
	// If right is already register allocated, then assign that value
	if inst.Right != 0 {
		for i := RegisterAddress(5); i < 32; i++ {
			if am.registers[i].Assignment == inst.Right {
				am.registers[i].Used = true
				r2 = i
				break
			}
		}
	}

	if inst.Left != 0 && r1 == 0 {
		r1 = am.getBest(sb, inst.Index)
		fmt.Fprintf(sb, "  lw %s %d(sp)\n", r1, am.addresses[inst.Left].MemoryAddress)
	}
	if inst.Right != 0 && r2 == 0 {
		r2 = am.getBest(sb, inst.Index)
		fmt.Fprintf(sb, "  lw %s %d(sp)\n", r2, am.addresses[inst.Right].MemoryAddress)
	}

	for i := RegisterAddress(5); i < 32; i++ {
		am.registers[i].Used = false
	}

	dest = am.getBest(sb, inst.Index)

	if inst.Left != 0 {
		am.registers[r1].Assignment = inst.Left
	}
	if inst.Right != 0 {
		am.registers[r2].Assignment = inst.Right
	}

	am.registers[dest].Assignment = inst.Index
	return dest, r1, r2
}

func (am *AddressMap) getBest(sb *strings.Builder, current ir.Assignment) RegisterAddress {
	best := 0x3FFFFFFF
	idx := RegisterAddress(0)

	for i := RegisterAddress(5); i < 10; i++ {
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

func (am *AddressMap) spill() {

}

func Compile(insts []ir.Instruction) string {
	return CompileBlock(insts)
}

func CompileBlock(block []ir.Instruction) string {
	sb := strings.Builder{}
	am := AddressMap{addresses: make([]AddressDescriptor, len(block))}
	for i := 0; i < len(block); i++ {
		am.addresses[block[i].Left].LastUsed = ir.Assignment(i+1)
		am.addresses[block[i].Right].LastUsed = ir.Assignment(i+1)
	}

	for _, inst := range block {
		switch inst.Kind {
		case ir.Function:
			fmt.Fprintf(&sb, "%s:\n", inst.Literal.(string))
		case ir.Arg:
			am.addresses[inst.Index].MemoryAddress = 4 + StackAddress(inst.Literal.(int))
		case ir.Add:
			dest, r1, r2 := am.GetReg(&sb, inst)
			fmt.Fprintf(&sb, "  add %s, %s, %s\n", dest, r1, r2)
		case ir.Mul:
			dest, r1, r2 := am.GetReg(&sb, inst)
			fmt.Fprintf(&sb, "  mul %s, %s, %s\n", dest, r1, r2)
		case ir.I64:
			dest, _, _ := am.GetReg(&sb, inst)
			fmt.Fprintf(&sb, "  li %s, %d\n", dest, inst.Literal.(int64))
		default:
			fmt.Printf("NOT HANDLED IN COMPILE rv64: %s\n", inst.Kind)
		}
	}
	return sb.String()
}
