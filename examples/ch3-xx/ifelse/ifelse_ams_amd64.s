// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "textflag.h"

//
// https://github.com/golang/go/issues/14288
//
// from rsc:
// But expanding what I said yesterday just a bit:
// never use MOVB or MOVW with a register destination,
// since it's inefficient (it's a read-modify-write on the target register).
// Instead use MOVL for reg->reg and use MOVBLZX or MOVWLZX for mem->reg;
// those are pure writes on the target register.
//
// 因此, 加载bool型参数到寄存器时, 建议使用 MOVBLZX.
// 如果使用 MOVB 的话, go test 虽然通过了,
// 但是 go run runme.go 则出现错误结果.
//

// func AsmIf(ok bool, a, b int) int
TEXT ·AsmIf(SB), NOSPLIT, $0-32
	MOVBQZX ok+0(FP), AX // ok
	MOVQ a+8(FP), BX     // a
	MOVQ b+16(FP), CX    // b
	CMPQ AX, $0          // test ok
	JEQ  3(PC)           // if !ok, skip 2 line
	MOVQ BX, ret+24(FP)  // return a
	RET
	MOVQ CX, ret+24(FP)  // return b
	RET
