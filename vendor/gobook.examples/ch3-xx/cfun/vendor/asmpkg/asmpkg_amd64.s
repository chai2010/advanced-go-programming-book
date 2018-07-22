// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "textflag.h"

// func CallCAdd(cfun uintptr, a, b int64) int64
TEXT ·CallCAdd(SB), NOSPLIT, $16
	MOVQ cfun+0(FP), AX // cfun
	MOVQ a+8(FP),    BX // a
	MOVQ b+16(FP),   CX // b

	MOVQ BX, 0(SP)
	MOVQ CX, 8(SP)
	CALL AX
	MOVQ AX, ret+24(FP)

	RET
