# 3.5 Control Flow

The program mainly has several execution processes of order, branch and loop. This section focuses on how to translate the control flow of the Go language into an assembler, or how to write Go code in assembly thinking.

## 3.5.1 Sequential execution

Sequential execution is a familiar working mode, similar to the so-called running account programming. All Go functions that do not have branches, loops, and goto statements, and are not recursively called, are generally executed sequentially.

For example, the code executed in the following order:

```go
Func main() {
Var a = 10
Println(a)

Var b = (a+a)*a
Println(b)
}
```

We try to rewrite the above function with the idea of ​​Go assembly. Since there are generally only 2 operands in an X86 instruction, there can be at most one operator in a variable expression that is required to be rewritten with assembly. At the same time, for some function calls, you also need to use the functions that can be called in the assembly to rewrite.

The first step to rewrite is still to use the Go language, but to rewrite it with assembly thinking:

```
Func main() {
Var a, b int

a = 10
Runtime.printint(a)
Runtime.printnl()

b = a
b += b
b *= a
Runtime.printint(b)
Runtime.printnl()
}
```

The preferred way of mimicking the C language is to declare all local variables at the function entry. Then according to the style of MOV, ADD, MUL and other instructions, the previous variable expression is expanded into multiple instructions expressed by several operations of `=`, `+=`, and `*=`. Finally, use the printint and printnl functions inside the runtime package to replace the output of the previous println function.

After rewriting with the assembly mind, the Go function described above is a bit cumbersome, but it is still relatively easy to understand. Below we further try to continue to translate the rewritten function into an assembly function:

```
TEXT · main(SB), $24-0
MOVQ $0, a-8*2(SP) // a = 0
MOVQ $0, b-8*1(SP) // b = 0

// Write the new value to a corresponding memory
MOVQ $10, AX // AX = 10
MOVQ AX, a-8*2(SP) // a = AX

// call function with a as argument
MOVQ AX, 0(SP)
CALL runtime·printint(SB)
CALL runtime·printnl(SB)

// After the function call, the AX/BX register may be contaminated and needs to be reloaded
MOVQ a-8*2(SP), AX // AX = a
MOVQ b-8*1(SP), BX // BX = b

/ / Calculate the b value, and write to the memory
MOVQ AX, BX // BX = AX // b = a
ADDQ BX, BX // BX += BX // b += a
IMULQ AX, BX // BX *= AX // b *= a
MOVQ BX, b-8*1(SP) // b = BX

// call function with b as argument
MOVQ BX, 0(SP)
CALL runtime·printint(SB)
CALL runtime·printnl(SB)

RET
```

The first step in assembling the main function is to calculate the size of the function stack frame. Because there are two int type variables in the function, the runtime and printint function parameters are an int type and there is no return value, so the stack frame of the main function is a stack of 24 bytes of int type. space.

Initialize the variable to a value of 0 at the beginning of the function, where `a-8*2(SP)` corresponds to a variable, `a-8*1(SP)` corresponds to b variable (because a variable is defined first, so a The address of the variable is smaller).

Then assign an AX register to the a variable, and set the memory corresponding to the a variable to 10 through the AX register, and AX is also 10. In order to output the a variable, the value of the AX register needs to be placed at the `0(SP)` position. The variable at this position will be printed as its argument when the runtime·printint function is called. Because we have previously saved the value of AX to the a variable memory, there is no need to perform a backup of the register before calling the function.

After the callback of the function, all registers will be treated as function modifications that may be called, so we need to restore registers AX and BX from the memory corresponding to a and b. Then refer to the calculation method of the b variable in the Go language above to update the value corresponding to BX. After the calculation is completed, the value of BX is also written to the memory corresponding to b.

It should be noted that in the above code, `IMULQ AX, BX` uses the `IMULQ` instruction to calculate the multiplication. The reason for not using the `MULQ` instruction is that the `MULQ` instruction uses `AX` to save the result by default. Readers can try to rewrite the above code with the `MULQ` command.

Finally, use the b variable as a parameter to call the runtime·printint function again for output work. All registers may also be contaminated, but the main function returns immediately, so there is no need to restore AX, BX, etc. registers.

Reanalysing the entire function after assembly rewriting will reveal a lot of redundant code inside. We don't need two temporary variables a and b to allocate two memory spaces, and we don't need to write to memory after each register change. The following is an optimized assembly function:

```
TEXT · main(SB), $16-0
// var temp int

// Write the new value to a corresponding memory
MOVQ $10, AX // AX = 10
MOVQ AX, temp-8(SP) // temp = AX

// call function with a as argument
CALL runtime·printint(SB)
CALL runtime·printnl(SB)

// After the function is called, AX may be contaminated and needs to be reloaded
MOVQ temp-8*1(SP), AX // AX = temp

/ / Calculate b value, no need to write to memory
MOVQ AX, BX // BX = AX // b = a
ADDQ BX, BX // BX += BX // b += a
IMULQ AX, BX // BX *= AX // b *= a

// ...
```

The first is to reduce the stack frame size of the main function from 24 bytes to 16 bytes. The only thing that needs to be saved is the value of the a variable, so all registers may be contaminated when calling the output of the runtime·printint function. We cannot back up the value of the a variable through the register, only the value in the stack memory is safe. Then the BX register does not need to be saved to memory. The rest of the code remains basically the same.

## 3.5.2 if/goto jump

When the Go language was just open source, there was no goto statement. Later, although the Go language added a goto statement, it is not recommended for use in programming. There is a similar principle to cgo: if you can not use the goto statement, then don't use the goto statement. The goto statement in the Go language is strictly limited: it cannot cross code blocks and cannot contain variable-defined statements in the code being spanned. Although the Go language does not recommend the goto statement, goto does have the favorite of every assembly language. Because goto is approximately equivalent to the unconditional jump instruction JMP in assembly language, the conditional jump instruction is formed with the if condition goto, and the conditional jump instruction is the cornerstone for constructing the entire assembly code control flow.

To make it easier to understand, we construct an If function that simulates a ternary expression in Go:

```go
Func If(ok bool, a, b int) int {
If ok { return a } else { return b }
}
```

For example, to find the ternary expression of the maximum number of two numbers `(a>b)?a:b` can be expressed by the If function: `If(a>b, a, b)`. Because of the language limitations, the If function used to simulate ternary expressions does not support generics (you can change the a, b, and return types to empty interfaces, but the use is cumbersome).

Although this function seems to have only a simple line, it contains an if branch statement. Before switching to assembly implementation, we still use the assembly thinking to re-examine the If function. In rewriting, you must also follow the constraint that each expression can only have one operator. At the same time, the conditional part of the if statement must have only one comparison symbol. The body part of the if statement can only be a goto statement.

The If function rewritten with assembly thinking is implemented as follows:

```go
Func If(ok int, a, b int) int {
If ok == 0 { goto L }
Return a
L:
Return b
}
```

Because there is no bool type in assembly language, we use the int type instead of the bool type (the real assembly uses byte to represent the bool type, and the byte type value can be loaded by the MOVBQZX instruction, which is simplified here). The variable a is returned when the ok parameter is non-zero, otherwise the variable b is returned. We reverse the logic of ok: when the ok parameter is 0, it means return b, otherwise it returns the variable a. In the if statement, when the ok parameter is 0, goto to the statement specified by the L label, that is, return the variable b. If the if condition is not satisfied, that is, the ok parameter is non-zero, executing the following statement returns the variable a.

The implementation of the above function is very close to the assembly language, the following is the code to the assembly implementation:

```
TEXT ·If(SB), NOSPLIT, $0-32
MOVQ ok+8*0(FP), CX // ok
MOVQ a+8*1(FP), AX // a
MOVQ b+8*2(FP), BX // b

CMPQ CX, $0 // test ok
JZ L // if ok == 0, goto L
MOVQ AX, ret+24(FP) // return a
RET

L:
MOVQ BX, ret+24(FP) // return b
RET
```

The first is to load three parameters into the register, the ok parameter corresponds to the CX register, and a and b correspond to the AX and BX registers, respectively. The CX register is then compared to the constant 0 using the CMPQ compare instruction. If the result of the comparison is 0, then the next JZ is 0, the jump instruction will jump to the statement corresponding to the L label, that is, the value of the variable b is returned. If the result of the comparison is not 0, then the JZ instruction will have no effect and continue to execute the following instructions, that is, return the value of the variable a.

In a jump instruction, the target of the jump is generally indicated by a label. However, in some functions implemented by macros, it is more desirable to jump through relative positions. At this time, the offset of the PC register can be used to calculate the position of the adjacent jump.

## 3.5.3 for loop

The Go language's for loop has many uses, and we only choose the most classic for structure to discuss. The classic for loop consists of three parts: initialization, end condition, and iteration step size. It is combined with the if condition language inside the loop body. This for structure can simulate other various loop types.

Based on the classic for loop structure, we define a LoopAdd function that can be used to calculate the sum of any arithmetic progression:

```go
Func LoopAdd(cnt, v0, step int) int {
Result := v0
For i := 0; i < cnt; i++ {
Result += step
}
Return result
}
```

For example, the `1+2+...+100` isolating sequence can calculate `LoopAdd(100, 1, 1)`, and the `10+8+...+0` isolating sequence can be calculated as `LoopAdd (5, 10, -2)`. Before the complete rewrite with assembly, use the technique similar to `if/goto` to modify the for loop.

The new LoopAdd function consists of only if/goto statements:

```go
Func LoopAdd(cnt, v0, step int) int {
Var i = 0
Var result = 0

LOOP_BEGIN:
Result = v0

LOOP_IF:
If i < cnt { goto LOOP_BODY }
Goto LOOP_END

LOOP_BODY
i = i+1
Result = result + step
Goto LOOP_IF

LOOP_END:

Return result
}
```

At the beginning of the function, two local variables are defined to facilitate subsequent code usage. Then the three parts of the for statement initialization, end condition, and iteration step are split into three code segments, using LOOP_BEGIN, LOOP_IF, and LOOP_BODY.Show. The LOOP_BEGIN loop initialization part will only be executed once, so the label will not be referenced and can be omitted. The last LOOP_END statement indicates the end of the for loop. The three code segments separated by four labels correspond to the initialization statement of the for loop, the loop condition, and the loop body, where the iteration statement is merged into the loop body.

The following implementation of the LoopAdd function in assembly language

```
#include "textflag.h"

// func LoopAdd(cnt, v0, step int) int
TEXT ·LoopAdd(SB), NOSPLIT, $0-32
MOVQ cnt+0(FP), AX // cnt
MOVQ v0+8(FP), BX // v0/result
MOVQ step+16(FP), CX // step

LOOP_BEGIN:
MOVQ $0, DX // i

LOOP_IF:
CMPQ DX, AX // compare i, cnt
JL LOOP_BODY // if i < cnt: goto LOOP_BODY
JMP LOOP_END

LOOP_BODY:
ADDQ $1, DX // i++
ADDQ CX, BX // result += step
JMP LOOP_IF

LOOP_END:

MOVQ BX, ret+24(FP) // return result
RET
```

The v0 and result variables are multiplexed with a BX register. In the instruction part corresponding to the LOOP_BEGIN label, the DX register is initialized to 0 with MOVQ, the DX corresponds to the variable i, and the iteration variable of the loop. In the instruction part corresponding to the LOOP_IF label, use the CMPQ instruction to compare DX and AX. If the loop does not end, jump to the LOOP_BODY part, otherwise jump to the LOOP_END part to end the loop. In the LOOP_BODY section, the iterator variable is updated and the accumulation statement in the loop body is executed, and then jumps directly to the LOOP_IF section to proceed to the next round of loop condition determination. The LOOP_END label is followed by a statement that returns the result of the accumulation.

Loops are the most complex control flow, and branches and jump statements are implicit in the loop. Mastering the looping method basically grasps the basic writing method of assembly language. The more geek gameplay is to break the traditional control flow through assembly language, such as directly returning across multiple layers of functions, such as reference gene editing directly to execute a code fragment constructed from C language. In short, after mastering the rules, you will find that assembly language programming becomes extremely simple and interesting.