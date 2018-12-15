// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "textflag.h"

// func AsmSumInt16Slice(v []int16) int16
TEXT ·AsmSumInt16Slice(SB), NOSPLIT, $0-26
	MOVQ v_base+0(FP), R8
	MOVQ v_len+8(FP), R9
	SHLQ $1, R9
	ADDQ R8, R9
	MOVQ $0, R10

loop:
	CMPQ R8, R9
	JE   end
	ADDW (R8), R10
	ADDQ $2, R8
	JMP  loop

end:
	MOVW R10, ret+24(FP)
	RET

// func AsmSumIntSlice(s []int) int
TEXT ·AsmSumIntSlice(SB), NOSPLIT, $0-32
	MOVQ s+0(FP), AX     // &s[0]
	MOVQ s_len+8(FP), BX // len(s)
	MOVQ $0, CX          // sum = 0

loop:
	CMPQ BX, $0   // compare cnt,0
	JLE  end      // if cnt <= 0: goto end
	DECQ BX       // cnt--
	ADDQ (AX), CX // sum += s[i]
	ADDQ $8, AX   // i++
	JMP  loop     // goto loop

end:
	MOVQ CX, ret+24(FP)  // return sum
	RET

// func AsmSumIntSliceV2(s []int) int
TEXT ·AsmSumIntSliceV2(SB), NOSPLIT, $0-32
	MOVQ s+0(FP), AX     // p := &s[0]
	MOVQ s_len+8(FP), BX
	LEAQ 0(AX)(BX*8), BX // p_end := &s[len(s)]
	MOVQ $0, CX          // sum = 0

loop:
	CMPQ AX, BX   // compare p,p_end
	JGE  end      // if p >= p_end: goto end
	ADDQ (AX), CX // sum += s[i]
	ADDQ $8, AX   // p++
	JMP  loop     // goto loop

end:
	MOVQ CX, ret+24(FP)  // return sum
	RET
