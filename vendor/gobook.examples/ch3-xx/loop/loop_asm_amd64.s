// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "textflag.h"

// func AsmLoopAdd(cnt, v0, step int) int
TEXT ·AsmLoopAdd(SB), NOSPLIT, $0-32
	MOVQ cnt+0(FP), AX   // cnt
	MOVQ v0+8(FP), BX    // v0
	MOVQ step+16(FP), CX // step

loop:
	CMPQ AX, $0 // compare cnt,0
	JLE  end    // if cnt <= 0: go end
	DECQ AX     // cnt--
	ADDQ CX, BX // v0 += step
	JMP  loop   // goto loop

end:
	MOVQ BX, ret+24(FP)  // return v0
	RET
