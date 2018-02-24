// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "textflag.h"

// func AsmAdd(a, b int) int
TEXT ·AsmAdd(SB), NOSPLIT, $0-24
	MOVQ a+0(FP), AX    // a
	MOVQ b+8(FP), BX    // b
	ADDQ AX, BX         // a+b
	MOVQ BX, ret+16(FP) // return a+b
	RET

// func AsmAddSlice(dst, a, b []int)
TEXT ·AsmAddSlice__todo(SB), NOSPLIT, $0-72
	MOVQ dst+0(FP), AX     // AX: dst
	MOVQ a+24(FP), BX      // BX: &a
	MOVQ b+48(FP), CX      // CX: &b
	MOVQ dst_len+8(FP), DX // DX: len(dst)
	MOVQ a_len+32(FP), R8  // R8: len(a)
	MOVQ b_len+56(FP), R9  // R9: len(b)
	// TODO: DX = min(DX,R8,R9)
	RET
