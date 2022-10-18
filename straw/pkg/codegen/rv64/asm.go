package rv64

type Instruction interface {
	Label() bool
	Binary() uint32
}

type R struct {
	funct7 uint8
	funct3 uint8
	rd     uint8
	rs1    uint8
	rs2    uint8
	opcode uint8
}

func (r *R) Binary() uint32 {
	return uint32(r.funct7)<<25 +
		uint32(r.rs2)<<20 +
		uint32(r.rs1)<<15 +
		uint32(r.funct3)<<12 +
		uint32(r.rd)<<7 +
		uint32(r.opcode)
}

type I struct {
	imm    uint32
	funct3 uint8
	rd     uint8
	rs1    uint8
	opcode uint8
}

func (i *I) Binary() uint32 {
	return (i.imm)<<20 +
		uint32(i.rs1)<<15 +
		uint32(i.funct3)<<12 +
		uint32(i.rd)<<7 +
		uint32(i.opcode)
}

type S struct {
	imm    uint32
	funct3 uint8
	rs1    uint8
	rs2    uint8
	opcode uint8
}

func (s *S) Binary() uint32 {
	return (s.imm>>5)<<25 +
		uint32(s.rs2)<<20 +
		uint32(s.rs1)<<15 +
		uint32(s.funct3)<<12 +
		uint32(s.rs1)<<7 +
		uint32(s.opcode)
}

type SB struct {
	imm    uint32
	funct3 uint8
	rs1    uint8
	rs2    uint8
	opcode uint8
}

func (sb *SB) Binary() uint32 {
	return ((sb.imm>>12)&0x1)<<30 +
		((sb.imm>>5)&0x3f)<<25 +
		uint32(sb.rs2)<<20 +
		uint32(sb.rs1)<<15 +
		uint32(sb.funct3)<<12 +
		((sb.imm>>1)&0x1f)<<8 +
		((sb.imm>>11)&0x1)<<7 +
		uint32(sb.opcode)
}

type U struct {
	imm    uint32
	rd     uint8
	opcode uint8
}

func (u *U) Binary() uint32 {
	return (u.imm>>12)<<12 +
		uint32(u.rd<<7) +
		uint32(u.opcode)
}

type UJ struct {
	imm    uint32
	rd     uint8
	opcode uint8
}

func (uj *UJ) Binary() uint32 {
	return ((uj.imm>>20)&0x1)<<30 +
		((uj.imm>>1)&0x3ff)<<21 +
		((uj.imm>>11)&0x1)<<20 +
		((uj.imm>>12)&0xff)<<12 +
		uint32(uj.rd)<<7 +
		uint32(uj.opcode)
}
