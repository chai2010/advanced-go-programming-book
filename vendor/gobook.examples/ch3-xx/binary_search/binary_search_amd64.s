// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

TEXT ·BinarySearch+0(SB),$0

start:
	MOVQ	arr+0(FP), CX
	MOVQ	len+8(FP), AX
	JMP		find_index

find_index:
	XORQ	DX, DX
	MOVQ	$2, BX
	IDIVQ	BX
	JMP		comp

comp:
	LEAQ	(AX * 8), BX
	ADDQ	BX, CX
	MOVQ	num+24(FP), DX
	CMPQ	DX, (CX)
	JE		found
	JG		right
	JL		left
	JMP		not_found

left:
	CMPQ	len+8(FP), $1
	JE		not_found
	MOVQ	AX, len+8(FP)
	JMP 	start

right:
	CMPQ	len+8(FP), $1
	JE		not_found
	MOVQ	CX, arr+0(FP)
	JMP 	start

not_found:
	MOVQ	$0, ret+32(FP)
	RET

found:
	MOVQ	$1, ret+32(FP)
	RET
