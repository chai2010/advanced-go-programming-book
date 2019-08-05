#include "textflag.h"

// func getg() interface{}
TEXT ·getg(SB), NOSPLIT, $32-16
	// get runtime.g
	MOVQ (TLS), AX
	// get runtime.g type
	MOVQ $type·runtime·g(SB), BX

	// convert (*g) to interface{}
	MOVQ AX, 8(SP)
	MOVQ BX, 0(SP)
	CALL ·runtime_convT2E_hack(SB)
	MOVQ 16(SP), AX
	MOVQ 24(SP), BX

	// return interface{}
	MOVQ AX, ret+0(FP)
	MOVQ BX, ret+8(FP)
	RET
