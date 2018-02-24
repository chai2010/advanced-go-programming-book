// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "textflag.h"

// func GetPkgValue() int
TEXT ·GetPkgValue(SB), NOSPLIT, $0-8
	MOVQ ·gopkgValue(SB), AX
	MOVQ AX, ret+0(FP)
	RET

// func GetPkgInfo() PkgInfo
TEXT ·GetPkgInfo(SB), NOSPLIT, $0-24
	MOVBLZX ·gInfo+0(SB), AX      // .V0 byte
	MOVQ    AX, ret+0(FP)
	MOVWLZX ·gInfo+2(SB), AX      // .V1 uint16
	MOVQ    AX, ret+2(FP)
	MOVLQZX ·gInfo+4(SB), AX      // .V2 int32
	MOVQ    AX, ret+4(FP)
	MOVQ    ·gInfo+8(SB), AX      // .V3 int32
	MOVQ    AX, ret+8(FP)
	MOVBLZX ·gInfo+(16+0)(SB), AX // .V4 bool
	MOVQ    AX, ret+(16+0)(FP)
	MOVBLZX ·gInfo+(16+1)(SB), AX // .V5 bool
	MOVQ    AX, ret+(16+1)(FP)
	RET
