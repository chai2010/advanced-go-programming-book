// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

TEXT ·SumVec+0(SB), $0
	MOVQ vec1+0(FP), BX  // Move the first vector into BX
	MOVQ vec2+24(FP), CX // Move the second vector into BX
	MOVUPS (BX), X0
	MOVUPS (CX), X1
	ADDPS X0, X1
	MOVUPS X1, result+48(FP)
	RET
