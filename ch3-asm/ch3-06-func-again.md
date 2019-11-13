In addition to the PC table, the FUNC table is used to record the parameters of the function and the pointer information of the local variables. The FUNCDATA instruction is similar to the format of PCDATA: `FUNCDATA tableid, tableoffset`, the first parameter is the type of the table, and the second is the address of the table. In the current implementation, three FUNC table types are defined: FUNCDATA_ArgsPointerMaps represents a pointer information table of function parameters, FUNCDATA_LocalsPointerMaps represents a local pointer information table, and FUNCDATA_InlTree represents a pointer information table that is inline expanded. Through the FUNC table, the Go garbage collector can track the life cycle of all pointers, and determine whether to move the pointer according to whether the address pointed by the pointer is in the range of the moved stack.

In the previous example of a recursive function, we encountered a NO_LOCAL_POINTERS macro. It is defined as follows:

```c
#define FUNCDATA_ArgsPointerMaps 0 /* garbage collector blocks */
#define FUNCDATA_LocalsPointerMaps 1
#define FUNCDATA_InlTree 2

#define NO_LOCAL_POINTERS FUNCDATA $FUNCDATA_LocalsPointerMaps, runtime·no_pointers_stackmap(SB)
```

Therefore, the NO_LOCAL_POINTERS macro represents the local pointer table corresponding to FUNCDATA_LocalsPointerMaps, and the runtime·no_pointers_stackmap is an empty pointer table, that is, a local variable indicating that the function has no pointer type.

The data of PCDATA and FUNCDATA is generally generated automatically by the compiler, and manual writing is not realistic. If the function already has a Go language declaration, the compiler can automatically output a pointer table of parameters and return values. At the same time, all function calls are generally corresponding to CALL instructions, and the compiler can also assist in generating PCDATA tables. The only thing the compiler can't automatically generate is a table of function local variables, so we generally use the pointer type carefully in the local variables of the assembly function.

Students interested in PCDATA and FUNCDATA details can try to start with the debug/gosym package, refer to the package implementation and test code.

## 3.6.4 Method Functions

The method functions in the Go language are very similar to the global functions, such as the following methods:

```go
Package main

Type MyInt int

Func (v MyInt) Twice() int {
Return int(v)*2
}

Func MyInt_Twice(v MyInt) int {
Return int(v)*2
}
```

The Twice method of the MyInt type is exactly the same as the MyInt_Twice function, except that Twice is modified to the name of the main.MyInt.Twice` in the target file. We can implement this method function in assembly:

```
// func (v MyInt) Twice() int
TEXT · MyInt·Twice (SB), NOSPLIT, $0-16
MOVQ a+0(FP), AX // v
ADDQ AX, AX // AX *= 2
MOVQ AX, ret+8(FP) // return v
RET
```

However, this is just a method function that accepts non-pointer types. Now add a Ptr method that takes a receive argument as a pointer type, and the function returns the pointer passed in:

```go
Func (p *MyInt) Ptr() *MyInt {
Return p
}
```

In the target file, the Ptr method name is modified to `main.(*MyInt).Ptr`, which is corresponding to `·(*MyInt)·Ptr` in the assembly. However, in the Go assembly language, neither the asterisk nor the parentheses can be used as function names, that is, methods that cannot directly use the assembly to receive parameters as pointer types.

There are a lot of special symbols in the final object file that are not supported by Go assembly language (such as double quotes in `type.string."hello"`, which leads to the inability to implement all the handwritten assembly code. Characteristics. Perhaps the Go language officially deliberately limited the features of assembly language.

## 3.6.5 Recursive function: 1 to n summation

Recursive functions are special functions. Recursive functions simplify the handling of many problems by calling itself and saving state on the stack. The power of the recursive function in the Go language is that you don't have to worry about popping the stack, because the stack can be expanded and shrunk as needed.

First implement a 1 to n summation function through the Go recursive function:

```go
// sum = 1+2+...+n
// sum(100) = 5050
Func sum(n int) int {
If n > 0 { return n+sum(n-1) } else { return 0 }
}
```

Then refactor the recursive function above via if/goto to escape to the assembly version:

```go
Func sum(n int) (result int) {
Var AX = n
Var BX int

If n > 0 { goto L_STEP_TO_END }
Goto L_END

L_STEP_TO_END:
AX -= 1
BX = sum(AX)

AX = n // After calling the function, AX is restored to n
BX += AX

Return BX

L_END:
Return 0
}
```

After rewriting, the parameters that are called recursively need to introduce local variables, and the intermediate results need to be introduced to local variables. The state of the call to save the middle through the stack is the core of the recursive function. Because the input parameters are also on the stack, we can save a small amount of state by entering parameters. At the same time we simulated the AX and BX registers, the registers need to be initialized before use, and they need to be reinitialized after the function call.

The following continues to be converted to an assembly language version:

```
// func sum(n int) (result int)
TEXT ·sum(SB), NOSPLIT, $16-16
MOVQ n+0(FP), AX // n
MOVQ result+8(FP), BX // result

CMPQ AX, $0 // test n - 0
JG L_STEP_TO_END // if > 0: goto L_STEP_TO_END
JMP L_END // goto L_STEP_TO_END

L_STEP_TO_END:
SUBQ $1, AX // AX -= 1
MOVQ AX, 0(SP) // arg: n-1
CALL ·sum(SB) // call sum(n-1)
MOVQ 8(SP), BX // BX = sum(n-1)

MOVQ n+0(FP), AX // AX = n
ADDQ AX, BX // BX += AX
MOVQ BX, result+8(FP) // return BX
RET

L_END:
MOVQ $0, result+8(FP) // return 0
RET
```

There is no local variable defined in the assembly version function, only the temporary stack space used to call itself. Because the parameters and return values ​​of the function itself have 16 bytes, the size of the stack frame is also 16 bytes. The L_STEP_TO_END label section is used to handle recursive calls and is a more complex part of the function. L_END is used to handle the part of the recursive finalization.

The parameter that calls the sum function is at the `0(SP)` position, and the return value after the call is finished at the `8(SP)` position. It is necessary to re-inject the value for the required register after the function call, because the called function is likely to destroy the state of the register. The parameter values ​​of the calling function at the same time are also untrustworthy, and the input parameter values ​​may also be modified inside the called function.

In general, there is no difference between using recursive functions and ordinary functions in assembly, of course, without considering the exploding stack. Our function should be able to sum the smaller n, but when n is large enough, that is, the stack reaches a certain depth, there will inevitably be a problem of bursting. Explosive stacks are a feature of the C language and should not appear in even Go assembly language.

The Go language compiler inserts a small piece of code at the beginning of the machine code that generates the function. Because the sum function also requires deep recursive calls, we removed the NOSPLIT flag and let the assembler automatically generate a stack extension code for us:

```
// func sum(n int) int
TEXT ·sum(SB), $16-16
NO_LOCAL_POINTERS

// original code
```

In addition to removing the NOSPLIT flag, we also added a NO_LOCAL_POINTERS statement at the beginning of the function, which indicates that the function does not have a local pointer variable. The expansion of the stack must involve the adjustment of function parameters and partial edit pointers. If the local pointer information is missing, the expansion work cannot be performed. Not only does the stack's expansion require function parameters and local pointer tag tables, it will also be needed for GC garbage collection. The function's parameters and the return state's pointer state can be obtained from the function declaration in the Go language, and the function's local variables need to be specified manually. Because manually specifying a pointer table is a very tedious task, it is generally desirable to avoid local pointers in handwritten assembly.

Readers who like to get to the bottom may have a question: How do the pointers in the registers be maintained if garbage collection or stack adjustments are made? As mentioned earlier, the Go function call is to pass parameters through the stack, and does not use registers to pass parameters. At the same time, all registers after the function call are considered invalid. Therefore, when adjusting and maintaining the pointer, only the pointer data in the memory needs to be scanned. The data in the register needs to be reloaded after the garbage collector function returns, so the register does not need to be scanned.

## 3.6.6 Closure function

The closure function is the most powerful function, because the closure function can capture the local variables of the outer local scope, so the closure function itself has a state. In theory, global functions are also a subset of closure functions, except that global functions do not capture outer variables.

To understand how the closure function works, let's construct the following example:

```go
Package main

Func NewTwiceFunClosure(x int) func() int {
Return func() int {
x *= 2
Return x
}
}

Func main() {
fnTwice := NewTwiceFunClosure(1)

Println(fnTwice()) // 1*2 => 2
Println(fnTwice()) // 2*2 => 4
Println(fnTwice()) // 4*2 => 8
}
```

The `NewTwiceFunClosure` function returns a closure function object, and the returned closure function object captures the outer `x` parameter. The returned closure function object, when executed, multiplies the captured outer variable by 2 and returns. In the `main` function, first call the `NewTwiceFunClosure` function with 1 as a parameter to construct a closure function. The returned closure function is stored in the variable of the `fnTwice` closure function type. Then each call to the `fnTwice` closure function will return the doubled result, which is: 2, 4, 8.

The above code is very easy to understand from the Go language level. But how does the closure function work at the assembly language level? Let's try to construct a closure function manually to show how the closure works. The first is to construct the `FunTwiceClosure` structure type, which is used to represent the closure object:

```go
Type FunTwiceClosure struct {
F uintptr
X int
}

Func NewTwiceFunClosure(x int) func() int {
Var p = &FunTwiceClosure{
F: asmFunTwiceClosureAddr(),
X: x,
}
Return ptrToFunc(unsafe.Pointer(p))
}
```

The `FunTwiceClosure` structure contains two members, the first member `F` represents the address of the function instruction of the closure function, and the second member `X` represents the external variable captured by the closure. If the closure function captures multiple external variables, then `FunTwiceClosure...