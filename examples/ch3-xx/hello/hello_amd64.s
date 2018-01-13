// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "textflag.h"
#include "funcdata.h"

// "Hello World!\n"
DATA  text<>+0(SB)/8,$"Hello Wo"
DATA  text<>+8(SB)/8,$"rld!\n"
GLOBL text<>(SB),NOPTR,$16

// utf8: "你好, 世界!\n"
// hex: e4bda0e5a5bd2c20 e4b896e7958c210a
// len: 16
DATA  text_zh<>+0(SB)/8,$"\xe4\xbd\xa0\xe5\xa5\xbd\x2c\x20"
DATA  text_zh<>+8(SB)/8,$"\xe4\xb8\x96\xe7\x95\x8c\x21\x0a"
GLOBL text_zh<>(SB),NOPTR,$16

// func PrintHelloWorld_var()
TEXT ·PrintHelloWorld_var(SB), $16-0
	NO_LOCAL_POINTERS
	CALL runtime·printlock(SB)
	MOVQ ·text+0(SB), AX
	MOVQ AX, (SP)
	MOVQ ·text+8(SB), AX
	MOVQ AX, 8(SP)
	CALL runtime·printstring(SB)
	CALL runtime·printunlock(SB)
	RET

// func PrintHelloWorld()
TEXT ·PrintHelloWorld(SB), $16-0
	NO_LOCAL_POINTERS
	CALL runtime·printlock(SB)
	MOVQ $text<>+0(SB), AX
	MOVQ AX, (SP)
	MOVQ $16, 8(SP)
	CALL runtime·printstring(SB)
	CALL runtime·printunlock(SB)
	RET

// func PrintHelloWorld_zh()
TEXT ·PrintHelloWorld_zh(SB), $16-0
	NO_LOCAL_POINTERS
	CALL runtime·printlock(SB)
	MOVQ $text_zh<>+0(SB), AX
	MOVQ AX, (SP)
	MOVQ $16, 8(SP)
	CALL runtime·printstring(SB)
	CALL runtime·printunlock(SB)
	RET
