# 3.9 Delve调试器

目前Go语言支持GDB、LLDB和Delve几种调试器。其中GDB是最早支持的调试工具，LLDB是macOS系统推荐的标准调试工具。但是GDB和LLDB对Go语言的专有特性都缺乏很大支持，而只有Delve是专门为Go语言设计开发的调试工具。而且Delve本身也是采用Go语言开发，对Windows平台也提供了一样的支持。本节我们基于Delve简单解释如何调试Go汇编程序。

## 3.9.1 Delve入门

首先根据官方的文档正确安装Delve调试器。我们会先构造一个简单的Go语言代码，用于熟悉下Delve的简单用法。

创建main.go文件，main函数先通过循初始化一个切片，然后输出切片的内容：

```go
package main

import (
	"fmt"
)

func main() {
	nums := make([]int, 5)
	for i := 0; i < len(nums); i++ {
		nums[i] = i * i
	}
	fmt.Println(nums)
}
```

命令行进入包所在目录，然后输入`dlv debug`命令进入调试：

```
$ dlv debug
Type 'help' for list of commands.
(dlv)
```

输入help命令可以查看到Delve提供的调试命令列表：

```
(dlv) help
The following commands are available:
    args ------------------------ Print function arguments.
    break (alias: b) ------------ Sets a breakpoint.
    breakpoints (alias: bp) ----- Print out info for active breakpoints.
    clear ----------------------- Deletes breakpoint.
    clearall -------------------- Deletes multiple breakpoints.
    condition (alias: cond) ----- Set breakpoint condition.
    config ---------------------- Changes configuration parameters.
    continue (alias: c) --------- Run until breakpoint or program termination.
    disassemble (alias: disass) - Disassembler.
    down ------------------------ Move the current frame down.
    exit (alias: quit | q) ------ Exit the debugger.
    frame ----------------------- Set the current frame, or execute command...
    funcs ----------------------- Print list of functions.
    goroutine ------------------- Shows or changes current goroutine
    goroutines ------------------ List program goroutines.
    help (alias: h) ------------- Prints the help message.
    list (alias: ls | l) -------- Show source code.
    locals ---------------------- Print local variables.
    next (alias: n) ------------- Step over to next source line.
    on -------------------------- Executes a command when a breakpoint is hit.
    print (alias: p) ------------ Evaluate an expression.
    regs ------------------------ Print contents of CPU registers.
    restart (alias: r) ---------- Restart process.
    set ------------------------- Changes the value of a variable.
    source ---------------------- Executes a file containing a list of delve...
    sources --------------------- Print list of source files.
    stack (alias: bt) ----------- Print stack trace.
    step (alias: s) ------------- Single step through program.
    step-instruction (alias: si)  Single step a single cpu instruction.
    stepout --------------------- Step out of the current function.
    thread (alias: tr) ---------- Switch to the specified thread.
    threads --------------------- Print out info for every traced thread.
    trace (alias: t) ------------ Set tracepoint.
    types ----------------------- Print list of types
    up -------------------------- Move the current frame up.
    vars ------------------------ Print package variables.
    whatis ---------------------- Prints type of an expression.
Type help followed by a command for full documentation.
(dlv)
```

每个Go程序的入口是main.main函数，我们可以用break在此设置一个断点：

```
(dlv) break main.main
Breakpoint 1 set at 0x10ae9b8 for main.main() ./main.go:7
```

然后通过breakpoints查看已经设置的所有断点：

```
(dlv) breakpoints
Breakpoint unrecovered-panic at 0x102a380 for runtime.startpanic()
    /usr/local/go/src/runtime/panic.go:588 (0)
        print runtime.curg._panic.arg
Breakpoint 1 at 0x10ae9b8 for main.main() ./main.go:7 (0)
```

我们发现除了我们自己设置的main.main函数断点外，Delve内部已经为panic异常函数设置了一个断点。

通过vars命令可以查看全部包级的变量。因为最终的目标程序可能含有大量的全局变量，我们可以通过一个正则参数选择想查看的全局变量：

```
(dlv) vars main
main.initdone· = 2
runtime.main_init_done = chan bool 0/0
runtime.mainStarted = true
(dlv)
```

然后就可以通过continue命令让程序运行到下一个断点处：

```
(dlv) continue
> main.main() ./main.go:7 (hits goroutine(1):1 total:1) (PC: 0x10ae9b8)
     2:
     3: import (
     4:         "fmt"
     5: )
     6:
=>   7: func main() {
     8:         nums := make([]int, 5)
     9:         for i := 0; i < len(nums); i++ {
    10:                 nums[i] = i * i
    11:         }
    12:         fmt.Println(nums)
(dlv)
```

输入next命令单步执行进入main函数内部：

```
(dlv) next
> main.main() ./main.go:8 (PC: 0x10ae9cf)
     3: import (
     4:         "fmt"
     5: )
     6:
     7: func main() {
=>   8:         nums := make([]int, 5)
     9:         for i := 0; i < len(nums); i++ {
    10:                 nums[i] = i * i
    11:         }
    12:         fmt.Println(nums)
    13: }
(dlv)
```

进入函数之后可以通过args和locals命令查看函数的参数和局部变量：

```
(dlv) args
(no args)
(dlv) locals
nums = []int len: 842350763880, cap: 17491881, nil
```

因为main函数没有参数，因此args命令没有任何输出。而locals命令则输出了局部变量nums切片的值：此时切片还未完成初始化，切片的底层指针为nil，长度和容量都是一个随机数值。

再次输入next命令单步执行后就可以查看到nums切片初始化之后的结果了：

```
(dlv) next
> main.main() ./main.go:9 (PC: 0x10aea12)
     4:         "fmt"
     5: )
     6:
     7: func main() {
     8:         nums := make([]int, 5)
=>   9:         for i := 0; i < len(nums); i++ {
    10:                 nums[i] = i * i
    11:         }
    12:         fmt.Println(nums)
    13: }
(dlv) locals
nums = []int len: 5, cap: 5, [...]
i = 17601536
(dlv)
```

此时因为调试器已经到了for语句行，因此局部变量出现了还未初始化的循环迭代变量i。

下面我们通过组合使用break和condition命令，在循环内部设置一个条件断点，当循环变量i等于3时断点生效：

```
(dlv) break main.go:10
Breakpoint 2 set at 0x10aea33 for main.main() ./main.go:10
(dlv) condition 2 i==3
(dlv)
```

然后通过continue执行到刚设置的条件断点，并且输出局部变量：

```
(dlv) continue
> main.main() ./main.go:10 (hits goroutine(1):1 total:1) (PC: 0x10aea33)
     5: )
     6:
     7: func main() {
     8:         nums := make([]int, 5)
     9:         for i := 0; i < len(nums); i++ {
=>  10:                 nums[i] = i * i
    11:         }
    12:         fmt.Println(nums)
    13: }
(dlv) locals
nums = []int len: 5, cap: 5, [...]
i = 3
(dlv) print nums
[]int len: 5, cap: 5, [0,1,4,0,0]
(dlv)
```

我们发现当循环变量i等于3时，nums切片的前3个元素已经正确初始化。

我们还可以通过stack查看当前执行函数的栈帧信息：

```
(dlv) stack
0  0x00000000010aea33 in main.main
   at ./main.go:10
1  0x000000000102bd60 in runtime.main
   at /usr/local/go/src/runtime/proc.go:198
2  0x0000000001053bd1 in runtime.goexit
   at /usr/local/go/src/runtime/asm_amd64.s:2361
(dlv)
```

或者通过goroutine和goroutines命令查看当前Goroutine相关的信息：

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
                runtime.gopark (0x102c189)
  Goroutine 3 - User: /usr/local/go/src/runtime/proc.go:292 \
                runtime.gopark (0x102c189)
  Goroutine 4 - User: /usr/local/go/src/runtime/proc.go:292 \
                runtime.gopark (0x102c189)
(dlv)
```

最后完成调试工作后输入quit命令退出调试器。至此我们已经掌握了Delve调试器器的简单用法。

## 3.9.2 调试汇编程序

用Delve调试Go汇编程序的过程比调试Go语言程序更加简单。调试汇编程序时，我们需要时刻关注寄存器的状态，如果涉及函数调用或局部变量或参数还需要重点关注栈寄存器SP的状态。

为了编译演示，我们重新实现一个更简单的main函数：

```go
package main

func main() { asmSayHello() }

func asmSayHello()
```

在main函数中调用汇编语言实现的asmSayHello函数输出一个字符串。

asmSayHello函数在main_amd64.s文件中实现：

```
#include "textflag.h"
#include "funcdata.h"

// "Hello World!\n"
DATA  text<>+0(SB)/8,$"Hello Wo"
DATA  text<>+8(SB)/8,$"rld!\n"
GLOBL text<>(SB),NOPTR,$16

// func asmSayHello()
TEXT ·asmSayHello(SB), $16-0
	NO_LOCAL_POINTERS
	MOVQ $text<>+0(SB), AX
	MOVQ AX, (SP)
	MOVQ $16, 8(SP)
	CALL runtime·printstring(SB)
	RET
```

参考前面的调试流程，在执行到main函数断点时，可以disassemble反汇编命令查看main函数对应的汇编代码：

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
  main.go:3 0x1050110  65488b0c25a0080000 mov rcx, qword ptr g  [0x8a0]
  main.go:3 0x1050119  483b6110           cmp rsp, qword ptr [r  +0x10]
  main.go:3 0x105011d  761a               jbe 0x1050139
=>main.go:3 0x105011f* 4883ec08           sub rsp, 0x8
  main.go:3 0x1050123  48892c24           mov qword ptr [rsp], rbp
  main.go:3 0x1050127  488d2c24           lea rbp, ptr [rsp]
  main.go:3 0x105012b  e880000000         call $main.asmSayHello
  main.go:3 0x1050130  488b2c24           mov rbp, qword ptr [rsp]
  main.go:3 0x1050134  4883c408           add rsp, 0x8
  main.go:3 0x1050138  c3                 ret
  main.go:3 0x1050139  e87288ffff         call $runtime.morestack_noctxt
  main.go:3 0x105013e  ebd0               jmp $main.main
(dlv)
```

虽然main函数内部只有一行函数调用语句，但是却生成了很多汇编指令。在函数的开头通过比较rsp寄存器判断栈空间是否不足，如果不足则跳转到0x1050139地址调用runtime.morestack函数进行栈扩容，然后跳回到main函数开始位置重新进行栈空间测试。而在asmSayHello函数调用之前，先扩展rsp空间用于临时存储rbp寄存器的状态，在函数返回后通过栈恢复rbp的值并回收临时栈空间。通过对比Go语言代码和对应的汇编代码，我们可以加深对Go汇编语言的理解。

从汇编语言角度深刻Go语言各种特性的工作机制对调试工作也是一个很大的帮助。如果希望在汇编指令层面调试Go代码，Delve还提供了一个step-instruction单步执行汇编指令的命令。

现在我们依然用break命令在asmSayHello函数设置断点，并且输入continue命令让调试器执行到断点位置停下：

```
(dlv) break main.asmSayHello
Breakpoint 2 set at 0x10501bf for main.asmSayHello() ./main_amd64.s:10
(dlv) continue
> main.asmSayHello() ./main_amd64.s:10 (hits goroutine(1):1 total:1) (PC: 0x10501bf)
     5: DATA  text<>+0(SB)/8,$"Hello Wo"
     6: DATA  text<>+8(SB)/8,$"rld!\n"
     7: GLOBL text<>(SB),NOPTR,$16
     8:
     9: // func asmSayHello()
=>  10: TEXT ·asmSayHello(SB), $16-0
    11:         NO_LOCAL_POINTERS
    12:         MOVQ $text<>+0(SB), AX
    13:         MOVQ AX, (SP)
    14:         MOVQ $16, 8(SP)
    15:         CALL runtime·printstring(SB)
(dlv)
```

此时我们可以通过regs查看全部的寄存器状态：

```
(dlv) regs
       rax = 0x0000000001050110
       rbx = 0x0000000000000000
       rcx = 0x000000c420000300
       rdx = 0x0000000001070be0
       rdi = 0x000000c42007c020
       rsi = 0x0000000000000001
       rbp = 0x000000c420049f78
       rsp = 0x000000c420049f70
        r8 = 0x7fffffffffffffff
        r9 = 0xffffffffffffffff
       r10 = 0x0000000000000100
       r11 = 0x0000000000000286
       r12 = 0x000000c41fffff7c
       r13 = 0x0000000000000000
       r14 = 0x0000000000000178
       r15 = 0x0000000000000004
       rip = 0x00000000010501bf
    rflags = 0x0000000000000206
...
(dlv)
```

因为AMD64的各种寄存器非常多，项目的信息中刻意省略了非通用的寄存器。如果再单步执行到13行时，可以发现AX寄存器值的变化。

```
(dlv) regs
       rax = 0x00000000010a4060
       rbx = 0x0000000000000000
       rcx = 0x000000c420000300
...
(dlv)
```

因此我们可以推断汇编程序内部定义的`text<>`数据的地址为0x00000000010a4060。我们可以用过print命令来查看该内存内的数据：

```
(dlv) print *(*[5]byte)(uintptr(0x00000000010a4060))
[5]uint8 [72,101,108,108,111]
(dlv)
```

我们可以发现输出的`[5]uint8 [72,101,108,108,111]`刚好是对应“Hello”字符串。通过类似的方法，我们可以通过查看SP对应的栈指针位置，然后查看栈中局部变量的值。

至此我们就掌握了Go汇编程序的简单调试技术。
