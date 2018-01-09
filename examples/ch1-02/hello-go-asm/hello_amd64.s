// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "textflag.h"
#include "funcdata.h"

// "Hello World!\n"
DATA  text<>+0(SB)/8,$"Hello Wo"
DATA  text<>+8(SB)/8,$"rld!\n"
GLOBL text<>(SB),NOPTR,$16

// func main()
TEXT ·main(SB), $16-0
    NO_LOCAL_POINTERS
    MOVQ $text<>+0(SB), AX
    MOVQ AX, (SP)
    MOVQ $16, 8(SP)
    CALL runtime·printstring(SB)
    RET
