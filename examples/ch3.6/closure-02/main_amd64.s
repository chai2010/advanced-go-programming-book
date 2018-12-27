// Copyright © 2018 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "textflag.h"

TEXT ·ptrToFunc(SB), NOSPLIT, $0-16
	MOVQ ptr+0(FP), AX // AX = ptr
	MOVQ AX, ret+8(FP) // return AX
	RET

TEXT ·asmFunTwiceClosureAddr(SB), NOSPLIT, $0-8
	LEAQ ·asmFunTwiceClosureBody(SB), AX // AX = ·asmFunTwiceClosureBody(SB)
	MOVQ AX, ret+0(FP)                   // return AX
	RET

TEXT ·asmFunTwiceClosureBody(SB), NOSPLIT|NEEDCTXT, $0-8
	MOVQ 8(DX), AX
	ADDQ AX   , AX        // AX *= 2
	MOVQ AX   , 8(DX)     // ctx.X = AX
	MOVQ AX   , ret+0(FP) // return AX
	RET

