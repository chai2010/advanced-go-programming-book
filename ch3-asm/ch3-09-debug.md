# 3.9 Delve Debugger

Currently, the Go language supports several debuggers for GDB, LLDB, and Delve. Among them, GDB is the earliest supported debugging tool, and LLDB is the standard debugging tool recommended by macOS system. However, GDB and LLDB lack great support for the proprietary features of the Go language, and only Delve is a debugging tool specifically designed for the Go language. And Delve itself is also developed in Go language, and provides the same support for the Windows platform. In this section we explain how to debug the Go assembler based on Delve.

## 3.9.1 Getting Started with Delve

First install the Delve debugger correctly according to the official documentation. We will first construct a simple Go language code to familiarize yourself with the simple usage of Delve.

Create the main.go file, the main function first initializes a slice, and then outputs the contents of the slice:

```go
Package main

Import (
"fmt"
)

Func main() {
Nums := make([]int, 5)
For i := 0; i < len(nums); i++ {
Nums[i] = i * i
}
fmt.Println(nums)
}
```

The command line enters the directory where the package is located, and then enters the `dlv debug` command to enter debugging:

```
$ dlv debug
Type 'help' for list of commands.
(dlv)
```

Enter the help command to see a list of debug commands provided by Delve:

```
(dlv) help
The following commands are available:
    Args ------------------------ Print function arguments.
    Break (alias: b) ------------ Sets a breakpoint.
    Breakpoints (alias: bp) ----- Print out info for active breakpoints.
    Clear ----------------------- Deletes breakpoint.
    Clearall -------------------- Deletes multiple breakpoints.
    Condition (alias: cond) ----- Set breakpoint condition.
    Config ---------------------- Changes configuration parameters.
    Continue (alias: c) --------- Run until breakpoint or program termination.
    Disassemble (alias: disass) - Disassembler.
    Down ------------------------ Move the current frame down.
    Exit (alias: quit | q) ------ Exit the debugger.
    Frame ----------------------- Set the current frame, or execute command...
    Funcs ----------------------- Print list of functions.
    Goroutine ------------------- Shows or changes current goroutine
    Goroutines ------------------ List program goroutines.
    Help (alias: h) ------------- Prints the help message.
    List (alias: ls | l) -------- Show source code.
    Locals ---------------------- Print local variables.
    Next (alias: n) ------------- Step over to next source line.
    On -------------------------- Executes a command when a breakpoint is hit.
    Print (alias: p) ------------ Evaluate an expression.
    Regs ------------------------ Print contents of CPU registers.
    Restart (alias: r) ---------- Restart process.
    Set ------------------------- Changes the value of a variable.
    Source ---------------------- Executes a file containing a list of delve...
    Sources --------------------- Print list of source files.
    Stack (alias: bt) ----------- Print stack trace.
    Step (alias: s) ------------- Single step through program.
    Step-instruction (alias: si) Single step a single cpu instruction.
    Stepout --------------------- Step out of the current function.
    Thread (alias: tr) ---------- Switch to the specified thread.
    Threads --------------------- Print out info for every traced thread.
    Trace (alias: t) ------------ Set tracepoint.
    Types ----------------------- Print list of types
    Up -------------------------- Move the current frame up.
    Vars ------------------------ Print package variables.
    Whatis ---------------------- Prints type of an expression.
Type help followed by a command for full documentation.
(dlv)
```

The entry for each Go program is the main.main function, and we can set a breakpoint here with break:

```
(dlv) break main.main
Breakpoint 1 set at 0x10ae9b8 for main.main() ./main.go:7
```

Then use breakpoints to see all the breakpoints that have been set:

```
(dlv) breakpoints
Breakpoint unrecovered-panic at 0x102a380 for runtime.startpanic()
    /usr/local/go/src/runtime/panic.go:588 (0)
        Print runtime.curg._panic.arg
Breakpoint 1 at 0x10ae9b8 for main.main() ./main.go:7 (0)
```

We found that in addition to the main.main function breakpoint we set ourselves, Delve internally set a breakpoint for the panic exception function.

The vars command allows you to view all package-level variables. Since the final target program may contain a large number of global variables, we can select the global variables we want to view with a regular parameter:

```
(dlv) vars main
Main.initdone· = 2
Runtime.main_init_done = chan bool 0/0
runtime.mainStarted = true
(dlv)
```

Then you can use the continue command to let the program run to the next breakpoint:

```
(dlv) continue
> main.main() ./main.go:7 (hits goroutine(1):1 total:1) (PC: 0x10ae9b8)
     2:
     3: import (
     4: "fmt"
     5: )
     6:
=> 7: func main() {
     8: nums := make([]int, 5)
     9: for i := 0; i < len(nums); i++ {
    10: nums[i] = i * i
    11: }
    12: fmt.Println(nums)
(dlv)
```

Enter the next command to step into the main function:

```
(dlv) next
> main.main() ./main.go:8 (PC: 0x10ae9cf)
     3: import (
     4: "fmt"
     5: )
     6:
     7: func main() {
=> 8: nums := make([]int, 5)
     9: for i := 0; i < len(nums); i++ {
    10: nums[i] = i * i
    11: }
    12: fmt.Println(nums)
    13: }
(dlv)
```

After entering the function, you can view the parameters and local variables of the function through the args and locals commands:

```
(dlv) args
(no args)
(dlv) locals
Nums = []int len: 842350763880, cap: 17491881, nil
```

Because the main function has no arguments, the args command has no output. The locals command outputs the value of the local variable nums slice: the slice has not yet been initialized, the underlying pointer of the slice is nil, and the length and capacity are both random values.

After entering the next command again, you can see the results after the nums slice is initialized:

```
(dlv) next
> main.main() ./main.go:9 (PC: 0x10aea12)
     4: "fmt"
     5: )
     6:
     7: func main() {
     8: nums := make([]int, 5)
=> 9: for i := 0; i < len(nums); i++ {
    10: nums[i] = i * i
    11: }
    12: fmt.Println(nums)
    13: }
(dlv) locals
Nums = []int len: 5, cap: 5, [...]
i = 17601536
(dlv)
```

At this point, because the debugger has reached the for statement line, the local variable has a loop iteration variable i that has not yet been initialized.

Below we use a combination of the break and condition commands to set a conditional breakpoint inside the loop. When the loop variable i equals 3, the breakpoint takes effect:

```
(dlv) break main.go: 10
Breakpoint 2 set at 0x10aea33 for main.main() ./main.go:10
(dlv) condition 2 i==3
(dlv)
```

Then execute the conditional breakpoint just set by continue and output the local variable:

```
(dlv) continue
> main.main() ./main.go:10 (hits goroutine(1):1 total:1) (PC: 0x10aea33)
     5: )
     6:
     7: func main() {
     8: nums := make([]int, 5)
     9: for i := 0; i < len(nums); i++ {
=> 10: nums[i] = i * i
    11: }
    12: fmt.Println(nums)
    13: }
(dlv) locals
Nums = []int len: 5, cap: 5, [...]
i = 3
(dlv) print nums
[]int len: 5, cap: 5, [0,1,4,0,0]
(dlv)
```

We found that when the loop variable i is equal to 3, the first 3 elements of the nums slice have been properly initialized.

We can also view the stack frame information of the current execution function through stack:

```
(dlv) stack
0 0x00000000010aea33 in main.main
   At ./main.go:10
1 0x000000000102bd60 in runtime.main
   At /usr/local/go/src/runtime/proc.go:198
2 0x0000000001053bd1 in runtime.goexit
   At /usr/local/go/src/runtime/asm_amd64.s:2361
(dlv)
```

Or use the goroutine and goroutines commands to view current Goroutine related information:

```
(dlv) goroutine
Thread 101686 at ./main.go:10
Goroutine 1:
  Runtime: ./main.go:10 main.main (0x10aea33)
  User: ./main.go:10 main.main (0x10aea33)
  Go: /usr/local/go/src/runtime/asm_amd64.s:258 runtime.rt0_go (0x1051643)
  Start: /usr/local/go/src/runtime/proc.go:109 runtime.main (0x102bb90)
(dlv) goroutines
[4 goroutines]
* Goroutine 1 - User: ./main.go:10 main.main (0x10aea33) (thread 101686)
  Goroutine 2 - User: /usr/local/go/src/runtime/proc.go:292 \
                Runtime.gopark (0x102c189)
  Goroutine 3 - User: /usr/local/go/src/runtime/proc.go:292 \
                Runtime.gopark (0x102c189)
  Goroutine 4 - User: /usr/local/go/src/runtime/proc.go:292 \
                Runtime.gopark (0x102c189)
(dlv)
```

After completing the debugging work, enter the quit command to exit the debugger. So far we have mastered the simple usage of the Delve debugger.

## 3.9.2 Debugging the assembler

The process of debugging a Go assembler with Delve is much simpler than debugging a Go program. When debugging the assembler, we need to keep an eye on the state of the registers. If we involve function calls or local variables or parameters, we need to focus on the state of the stack register SP.

To compile the demo, we reimplement a simpler main function:

```go
Package main

Func main() { asmSayHello() }

Func asmSayHello()
```

Call the assembly language implementation of the asmSayHello function in the main function to output a string.

The asmSayHello function is implemented in the main_amd64.s file:

```
#include "textflag.h"
#include "funcdata.h"

// "Hello World!\n"
DATA text<>+0(SB)/8,$"Hello Wo"
DATA text<>+8(SB)/8,$"rld!\n"
GLOBL text<>(SB),NOPTR,$16

// func asmSayHello()
TEXT · asmSayHello(SB), $16-0
NO_LOCAL_POINTERS
MOVQ $text<>+0(SB), AX
MOVQ AX, (SP)
MOVQ $16, 8(SP)
CALL runtime·printstring(SB)
RET
```

Referring to the previous debugging process, when executing the breakpoint to the main function, you can disassemble the disassembly command to view the assembly code corresponding to the main function:

```
(dlv) break main.main
Breakpoint 1 set at 0x105011f for main.main() ./main.go:3
(dlv) continue
> main.main() ./main.go:3 (hits goroutine(1):1 total:1) (PC: 0x105011f)
  1: package main
  2:
=>3: func main() { asmSayHello() }
  4:
  5: func asmSayHello()
(dlv) disassemble
TEXT main.main(SB) /path/to/pkg/main.go
  Main.go:3 0x1050110 65488b0c25a0080000 mov rcx, qword ptr g [0x8a0]
  Main.go:3 0x1050119 483b6110 cmp rsp, qword ptr [r +0x10]
  Main.go:3 0x105011d 761a jbe 0x1050139
=> main.go:3 0x105011f* 4883ec08 sub rsp, 0x8
  Main.go:3 0x1050123 48892c24 mov qword ptr [rsp], rbp
  Main.go:3 0x1050127 488d2c24 lea rbp, ptr [rsp]
  Main.go:3 0x105012b e880000000 call $main.asmSayHello
  Main.go:3 0x1050130 488b2c24 mov rbp, qword ptr [rsp]
  Main.go:3 0x1050134 4883c408 add rsp, 0x8
  Main.go:3 0x1050138 c3 ret
  Main.go:3 0x1050139 e87288ffff call $runtime.morestack_noctxt
  Main.go:3 0x105013e ebd0 jmp $main.main
(dlv)
```

Although there is only one line of function call statements inside the main function, a lot of assembly instructions are generated. At the beginning of the function, compare the rsp register to determine whether the stack space is insufficient. If it is insufficient, jump to 0x1050139 address and call the runtime.morestack function to expand the stack. Then jump back to the main function start position and re-test the stack space. Before the asmSayHello function is called, the rsp space is expanded to temporarily store the state of the rbp register. After the function returns, the rbp value is restored by the stack and the temporary stack space is reclaimed. By comparing the Go language code with the corresponding assembly code, we can deepen our understanding of Go assembly language.

From the perspective of assembly language, the working mechanism of the various features of the Go language is also a great help to the debugging work. If you want to debug Go code at the assembly instruction level, Delve also provides a step-instruction to step through the assembly instructions.

Now we still use the break command to set a breakpoint in the asmSayHello function, and enter the continue command to let the debugger execute to the breakpoint position:

```
(dlv) break main.asmSayHello
Breakpoint 2 set at 0x10501bf for main.asmSayHello() ./main_amd64.s:10
(dlv) continue
> main.asmSayHello() ./main_amd64.s:10 (hits goroutine(1):1 total:1) (PC: 0x10501bf)
     5: DATA text<>+0(SB)/8,$"Hello Wo"
     6: DATA text<>+8(SB)/8,$"rld!\n"
     7: GLOBL text<>(SB), NOPTR, $16
     8:
     9: // func asmSayHello()
=> 10: TEXT · asmSayHello(SB), $16-0
    11: NO_LOCAL_POINTERS
    12: MOVQ $text<>+0(SB), AX
    13: MOVQ AX, (SP)
    14: MOVQ $16, 8(SP)
    15: CALL runtime·printstring(SB)
(dlv)
```

At this point we can view all register states via regs:

```
(dlv) regs
       Rax = 0x0000000001050110
       Rbx = 0x0000000000000000
       Rcx = 0x000000c420000300
       Rdx = 0x0000000001070be0
       Rdi = 0x000000c42007c020
       Rsi = 0x0000000000000001
       Rbp = 0x000000c420049f78
       Rsp = 0x000000c420049f70
        R8 = 0x7fffffffffffffff
        R9 = 0xFfffffffffffffff
       R10 = 0x0000000000000100
       R11 = 0x0000000000000286
       R12 = 0x000000c41fffff7c
       R13 = 0x0000000000000000
       R14 = 0x0000000000000178
       R15 = 0x0000000000000004
       Rip = 0x00000000010501bf
    Rflags = 0x0000000000000206
...
(dlv)
```

Because of the many registers of AMD64, the non-universal registers are deliberately omitted from the project information. If you step through the 13 lines, you can find the change in the AX register value.

```
(dlv) regs
       Rax = 0x00000000010a4060
       Rbx = 0x0000000000000000
       Rcx = 0x000000c420000300
...
(dlv)
```

So we can infer that the address of the `text<>` data defined inside the assembler is 0x00000000010a4060. We can use the print command to view the data in the memory:

```
(dlv) print *(*[5]byte)(uintptr(0x00000000010a4060))
[5]uint8 [72,101,108,108,111]
(dlv)
```

We can find that the output of `[5]uint8 [72,101,108,108,111]` happens to correspond to the "Hello" string. In a similar way, we can look at the stack pointer position corresponding to the SP and then look at the value of the local variable in the stack.

So far we have mastered the simple debugging technology of the Go assembly program.