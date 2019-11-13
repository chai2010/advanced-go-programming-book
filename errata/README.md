# errata · first edition first print

## 1. ch3.4 The last figure has an error

The `ret+24(FP)` of the sum function is changed to `ret+16(FP)`

## 2. ch3.5 Control Flow - for example has error #438

The first code snippet on page 171 is changed to:

```go
Func LoopAdd(cnt, v0, step int) int {
Result, vi := 0, v0
For i := 0; i < cnt; i++ {
Result, vi = result+vi, vi+step
}
Return result
}
```

The changed code has 2 lines:
1. `result := v0` to `result, vi := 0, v0`
2. `result += step` is changed to `result, vi = result+vi, vi+step`

The second code snippet on page 171 is changed to:

```go
Func LoopAdd(cnt, v0, step int) int {
Var vi = v0
Var result = 0

// LOOP_BEGIN:
Var i = 0

LOOP_IF:
If i < cnt { goto LOOP_BODY }
Goto LOOP_END

LOOP_BODY:
i = i+1
Result = result + vi
Vi = vi + step
Goto LOOP_IF

LOOP_END:

Return result
}
```

The changed part:
1. `var i = 0` is changed to `var vi = v0`
2. `LOOP_BEGIN:` becomes a comment, followed by the code changed to `var i = 0`
3. In LOOP_BODY, `result = result + step` is changed to `result = result + vi`
4. Add a line to LOOP_BODY `vi = vi + step`

The third code snippet is changed to:

```go
#include "textflag.h"

// func LoopAdd(cnt, v0, step int) int
TEXT ·LoopAdd(SB), NOSPLIT, $0-32
MOVQ $0, BX // result
MOVQ cnt+0(FP), AX // cnt
MOVQ v0+8(FP), DI // vi = v0
MOVQ step+16(FP), CX // step

LOOP_BEGIN:
MOVQ $0, DX // i

LOOP_IF:
CMPQ DX, AX // compare i, cnt
JL LOOP_BODY // if i < cnt: goto LOOP_BODY
JMP LOOP_END

LOOP_BODY:
ADDQ DI, BX // result += vi
ADDQ CX, DI // vi += step
ADDQ $1, DX // i++
JMP LOOP_IF

LOOP_END:

MOVQ BX, ret+24(FP) // return result
RET
```

The above three code segments are actually different versions of the same program, and their changes are the same.

## 3. ch1.1 time has error

Page 1: "In 2010, the Go language has gradually stabilized. In September of the same year, the Go language was officially released and open source code."

Among them, 2010 was changed to 2009.