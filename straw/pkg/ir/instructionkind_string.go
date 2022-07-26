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
	_ = x[Mod-5]
	_ = x[Equals-6]
	_ = x[NotEquals-7]
	_ = x[Move-8]
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
	_ = x[End-21]
	_ = x[Function-22]
	_ = x[IfTrueGoto-23]
	_ = x[Goto-24]
	_ = x[Call-25]
	_ = x[Push-26]
	_ = x[Pop-27]
}

const _InstructionKind_name = "UndefinedAddSubMulQuoModEqualsNotEqualsMoveAndOrNotBoolI8I16I32I64F32F64PhiRetEndFunctionIfTrueGotoGotoCallPushPop"

var _InstructionKind_index = [...]uint8{0, 9, 12, 15, 18, 21, 24, 30, 39, 43, 46, 48, 51, 55, 57, 60, 63, 66, 69, 72, 75, 78, 81, 89, 99, 103, 107, 111, 114}

func (i InstructionKind) String() string {
	if i < 0 || i >= InstructionKind(len(_InstructionKind_index)-1) {
		return "InstructionKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _InstructionKind_name[_InstructionKind_index[i]:_InstructionKind_index[i+1]]
}
