#include "textflag.h"

// func LoopAdd(cnt, v0, step int) int
TEXT ·LoopAdd(SB), NOSPLIT,  $0-32
	MOVQ $0, BX          // result
	MOVQ cnt+0(FP), AX   // cnt
	MOVQ v0+8(FP), DI    // vi = v0
	MOVQ step+16(FP), CX // step

LOOP_BEGIN:
	MOVQ $0, DX          // i

LOOP_IF:
	CMPQ DX, AX          // compare i, cnt
	JL   LOOP_BODY       // if i < cnt: goto LOOP_BODY
	JMP LOOP_END

LOOP_BODY:
	ADDQ DI, BX          // result += vi
	ADDQ CX, DI          // vi += step
	ADDQ $1, DX          // i++
	JMP LOOP_IF

LOOP_END:

	MOVQ BX, ret+24(FP)  // return result
	RET
