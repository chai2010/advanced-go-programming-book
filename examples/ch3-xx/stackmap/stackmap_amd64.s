// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "funcdata.h"
#include "textflag.h"

// func X(b []byte) []byte
TEXT ·X(SB), $48-48
	MOVQ b_base+0(FP), BX
	MOVQ b_len+8(FP), CX
	MOVQ b_cap+16(FP), DX

	CMPQ CX, DX
	JL   afterGrow

	// Set up the growSlice call.
	MOVQ BX, gs_base-48(SP)
	MOVQ CX, gs_len-40(SP)
	MOVQ DX, gs_cap-32(SP)

	CALL ·growSlice(SB)

	MOVQ gs_base-24(SP), BX
	MOVQ gs_len-16(SP), CX
	MOVQ gs_cap-8(SP), DX

afterGrow:
	// At this point, we have adequate capacity to increase len + 1 and the
	// following register scheme:
	//   BX - b_base
	//   CX - b_len
	//   DX - b_cap

	// Write base/cap results.
	MOVQ BX, ret_base+24(FP)
	MOVQ DX, ret_cap+40(FP)

	// Write new element to b and increment the length.
	LEAQ (BX)(CX*1), BX
	MOVB $3, (BX)
	ADDQ $1, CX
	MOVQ CX, ret_len+32(FP)

	RET
