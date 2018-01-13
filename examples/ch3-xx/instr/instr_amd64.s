// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "textflag.h"

// func Add2(n, m int64) int32
TEXT ·Add2(SB), NOSPLIT, $0-24
	MOVQ n+0(FP), AX
	MOVQ m+8(FP), BX
	ADDQ AX, BX
	MOVQ BX, ret+16(FP)
	RET

// func BSF(n int64) int
TEXT ·BSF(SB), NOSPLIT, $0
	BSFQ n+0(FP), AX
	JEQ  allZero
	MOVQ AX, ret+8(FP)
	RET

allZero:
	MOVQ $-1, ret+8(FP)
	RET

// func BSF32(n int32) int32
TEXT ·BSF32(SB), NOSPLIT, $0
	BSFL n+0(FP), AX
	JEQ  allZero32
	MOVL AX, ret+8(FP)
	RET

allZero32:
	MOVL $-1, ret+8(FP)
	RET

// func Sum2(s []int64) int64
TEXT ·Sum2(SB), NOSPLIT, $0
	MOVQ $0, DX
	MOVQ s_base+0(FP), AX
	MOVQ s_len+8(FP), DI
	MOVQ $0, CX
	CMPQ CX, DI
	JGE  Sum2End

Sum2Loop:
	MOVQ (AX), BP
	ADDQ BP, DX
	ADDQ $8, AX
	INCQ CX
	CMPQ CX, DI
	JL   Sum2Loop

Sum2End:
	MOVQ DX, ret+24(FP)
	RET

// vim: set ft=txt:
