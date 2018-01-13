// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

TEXT ·Find+0(SB),$0
	MOVQ $0, SI         // zero the iterator
	MOVQ vec+0(FP), BX  // BX = &vec[0]
	MOVQ vec+8(FP), CX  // len(vec)
	MOVQ num+24(FP), DX

start:
	CMPQ SI, CX
	JG   notfound
	CMPQ (BX), DX
	JNE  notequal
	JE   found

found:
	MOVQ $1, return+32(FP)
	RET

notequal:
	INCQ SI
	LEAQ +8(BX), BX
	JMP  start

notfound:
	MOVQ $0, return+32(FP)
	RET
