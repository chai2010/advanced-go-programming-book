// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "textflag.h"

// func AsmMin(a, b int) int
TEXT ·AsmMin(SB), NOSPLIT, $0-24
	MOVQ a+0(FP), AX    // a
	MOVQ b+8(FP), BX    // b
	CMPQ AX, BX         // compare a, b
	JGT  3(PC)          // if a>b, skip 2 line
	MOVQ AX, ret+16(FP) // return a
	RET
	MOVQ BX, ret+16(FP) // return b
	RET

// func AsmMax(a, b int) int
TEXT ·AsmMax(SB), NOSPLIT, $0-24
	MOVQ a+0(FP), AX    // a
	MOVQ b+8(FP), BX    // b
	CMPQ AX, BX         // compare a, b
	JLT  3(PC)          // if a<b, skip 2 line
	MOVQ AX, ret+16(FP) // return a
	RET
	MOVQ BX, ret+16(FP) // return b
	RET
