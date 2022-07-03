// Code generated by "stringer -type=InstructionKind"; DO NOT EDIT.

package ir

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Undefined-0]
	_ = x[Add-1]
	_ = x[Sub-2]
	_ = x[Mul-3]
	_ = x[Quo-4]
	_ = x[Equals-5]
	_ = x[NotEquals-6]
	_ = x[Move-7]
	_ = x[IfTrueGoto-8]
	_ = x[And-9]
	_ = x[Or-10]
	_ = x[Not-11]
	_ = x[Bool-12]
	_ = x[I8-13]
	_ = x[I16-14]
	_ = x[I32-15]
	_ = x[I64-16]
	_ = x[F32-17]
	_ = x[F64-18]
	_ = x[Phi-19]
	_ = x[Ret-20]
	_ = x[Function-21]
	_ = x[Arg-22]
	_ = x[Call-23]
}

const _InstructionKind_name = "UndefinedAddSubMulQuoEqualsNotEqualsMoveIfTrueGotoAndOrNotBoolI8I16I32I64F32F64PhiRetFunctionArgCall"

var _InstructionKind_index = [...]uint8{0, 9, 12, 15, 18, 21, 27, 36, 40, 50, 53, 55, 58, 62, 64, 67, 70, 73, 76, 79, 82, 85, 93, 96, 100}

func (i InstructionKind) String() string {
	if i < 0 || i >= InstructionKind(len(_InstructionKind_index)-1) {
		return "InstructionKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _InstructionKind_name[_InstructionKind_index[i]:_InstructionKind_index[i+1]]
}
