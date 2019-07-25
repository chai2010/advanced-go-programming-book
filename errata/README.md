# 勘误·第一版第一次印刷

## 1. ch3.4最后一个图有错误

sum函数的`ret+24(FP)`改为`ret+16(FP)`

## 2. ch3.5 控制流 - for例子有错误 #438

171页第一个代码段改为：

```go
func LoopAdd(cnt, v0, step int) int {
	result, vi := 0, v0
	for i := 0; i < cnt; i++ {
		result, vi = result+vi, vi+step
	}
	return result
}
```

改动的代码有2行:
1. `result := v0`改为`result, vi := 0, v0`
2. `result += step`改为`result, vi = result+vi, vi+step`

171页第二个代码段改为:

```go
func LoopAdd(cnt, v0, step int) int {
	var vi = v0
	var result = 0

// LOOP_BEGIN:
	var i = 0

LOOP_IF:
	if i < cnt { goto LOOP_BODY }
	goto LOOP_END

LOOP_BODY:
	i = i+1
	result = result + vi
	vi = vi + step
	goto LOOP_IF

LOOP_END:

	return result
}
```

改动的部分：
1. `var i = 0`改为`var vi = v0`
2. `LOOP_BEGIN:`变成注释，其后的代码改为`var i = 0`
3. LOOP_BODY中`result = result + step`改为`result = result + vi`
4. LOOP_BODY中增加一行`vi = vi + step`

第三个代码段改为：

```go
#include "textflag.h"

// func LoopAdd(cnt, v0, step int) int
TEXT ·LoopAdd(SB), NOSPLIT,  $0-32
	MOVQ $0, BX          // result
	MOVQ cnt+0(FP), AX   // cnt
	MOVQ v0+8(FP), DI    // vi = v0
	MOVQ step+16(FP), CX // step

LOOP_BEGIN:
	MOVQ $0, DX          // i

LOOP_IF:
	CMPQ DX, AX          // compare i, cnt
	JL   LOOP_BODY       // if i < cnt: goto LOOP_BODY
	JMP LOOP_END

LOOP_BODY:
	ADDQ DI, BX          // result += vi
	ADDQ CX, DI          // vi += step
	ADDQ $1, DX          // i++
	JMP LOOP_IF

LOOP_END:

	MOVQ BX, ret+24(FP)  // return result
	RET
```

以上三个代码段其实是同一个程序的不同版本，他们的改动都是相同的问题。

## 3. ch1.1时间有错误

第1页：“到了2010年，Go语言已经逐步趋于稳定。同年9月，Go语言正式发布并开源了代码。”

其中2010改为2009.


