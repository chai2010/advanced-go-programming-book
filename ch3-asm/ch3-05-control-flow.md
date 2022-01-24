# 3.5 控制流

程序主要有顺序、分支和循环几种执行流程。本节主要讨论如何将 Go 语言的控制流比较直观地转译为汇编程序，或者说如何以汇编思维来编写 Go 语言代码。

## 3.5.1 顺序执行

顺序执行是我们比较熟悉的工作模式，类似俗称流水账编程。所有不含分支、循环和 goto 语句，并且没有递归调用的 Go 函数一般都是顺序执行的。

比如有如下顺序执行的代码：

```go
func main() {
	var a = 10
	println(a)

	var b = (a+a)*a
	println(b)
}
```

我们尝试用 Go 汇编的思维改写上述函数。因为 X86 指令中一般只有 2 个操作数，因此在用汇编改写时要求出现的变量表达式中最多只能有一个运算符。同时对于一些函数调用，也需要用汇编中可以调用的函数来改写。

第一步改写依然是使用 Go 语言，只不过是用汇编的思维改写：

```
func main() {
	var a, b int

	a = 10
	runtime.printint(a)
	runtime.printnl()

	b = a
	b += b
	b *= a
	runtime.printint(b)
	runtime.printnl()
}
```

首选模仿 C 语言的处理方式在函数入口处声明全部的局部变量。然后根据 MOV、ADD、MUL 等指令的风格，将之前的变量表达式展开为用 `=`、`+=` 和 `*=` 几种运算表达的多个指令。最后用 runtime 包内部的 printint 和 printnl 函数代替之前的 println 函数输出结果。

经过用汇编的思维改写过后，上述的 Go 函数虽然看着繁琐了一点，但是还是比较容易理解的。下面我们进一步尝试将改写后的函数继续转译为汇编函数：

```
TEXT ·main(SB), $24-0
	MOVQ $0, a-8*2(SP) // a = 0
	MOVQ $0, b-8*1(SP) // b = 0

	// 将新的值写入 a 对应内存
	MOVQ $10, AX       // AX = 10
	MOVQ AX, a-8*2(SP) // a = AX

	// 以 a 为参数调用函数
	MOVQ AX, 0(SP)
	CALL runtime·printint(SB)
	CALL runtime·printnl(SB)

	// 函数调用后, AX/BX 寄存器可能被污染, 需要重新加载
	MOVQ a-8*2(SP), AX // AX = a
	MOVQ b-8*1(SP), BX // BX = b

	// 计算 b 值, 并写入内存
	MOVQ AX, BX        // BX = AX  // b = a
	ADDQ BX, BX        // BX += BX // b += a
	IMULQ AX, BX       // BX *= AX // b *= a
	MOVQ BX, b-8*1(SP) // b = BX

	// 以 b 为参数调用函数
	MOVQ BX, 0(SP)
	CALL runtime·printint(SB)
	CALL runtime·printnl(SB)

	RET
```

汇编实现 main 函数的第一步是要计算函数栈帧的大小。因为函数内有 a、b 两个 int 类型变量，同时调用的 runtime·printint 函数参数是一个 int 类型并且没有返回值，因此 main 函数的栈帧是 3 个 int 类型组成的 24 个字节的栈内存空间。

在函数的开始处先将变量初始化为 0 值，其中 `a-8*2(SP)` 对应 a 变量、`a-8*1(SP)` 对应 b 变量（因为 a 变量先定义，因此 a 变量的地址更小）。

然后给 a 变量分配一个 AX 寄存器，并且通过 AX 寄存器将 a 变量对应的内存设置为 10，AX 也是 10。为了输出 a 变量，需要将 AX 寄存器的值放到 `0(SP)` 位置，这个位置的变量将在调用 runtime·printint 函数时作为它的参数被打印。因为我们之前已经将 AX 的值保存到 a 变量内存中了，因此在调用函数前并不需要再进行寄存器的备份工作。

在调用函数返回之后，全部的寄存器将被视为可能被调用的函数修改，因此我们需要从 a、b 对应的内存中重新恢复寄存器 AX 和 BX。然后参考上面 Go 语言中 b 变量的计算方式更新 BX 对应的值，计算完成后同样将 BX 的值写入到 b 对应的内存。

需要说明的是，上面的代码中 `IMULQ AX, BX` 使用了 `IMULQ` 指令来计算乘法。没有使用 `MULQ` 指令的原因是 `MULQ` 指令默认使用 `AX` 保存结果。读者可以自己尝试用 `MULQ` 指令改写上述代码。

最后以 b 变量作为参数再次调用 runtime·printint 函数进行输出工作。所有的寄存器同样可能被污染，不过 main 函数马上就返回了，因此不再需要恢复 AX、BX 等寄存器了。

重新分析汇编改写后的整个函数会发现里面很多的冗余代码。我们并不需要 a、b 两个临时变量分配两个内存空间，而且也不需要在每个寄存器变化之后都要写入内存。下面是经过优化的汇编函数：

```
TEXT ·main(SB), $16-0
	// var temp int

	// 将新的值写入 a 对应内存
	MOVQ $10, AX        // AX = 10
	MOVQ AX, temp-8(SP) // temp = AX

	// 以 a 为参数调用函数
	CALL runtime·printint(SB)
	CALL runtime·printnl(SB)

	// 函数调用后, AX 可能被污染, 需要重新加载
	MOVQ temp-8*1(SP), AX // AX = temp

	// 计算 b 值, 不需要写入内存
	MOVQ AX, BX        // BX = AX  // b = a
	ADDQ BX, BX        // BX += BX // b += a
	IMULQ AX, BX       // BX *= AX // b *= a

	// ...
```

首先是将 main 函数的栈帧大小从 24 字节减少到 16 字节。唯一需要保存的是 a 变量的值，因此在调用 runtime·printint 函数输出时全部的寄存器都可能被污染，我们无法通过寄存器备份 a 变量的值，只有在栈内存中的值才是安全的。然后在 BX 寄存器并不需要保存到内存。其它部分的代码基本保持不变。

## 3.5.2 if/goto 跳转

Go 语言刚刚开源的时候并没有 goto 语句，后来 Go 语言虽然增加了 goto 语句，但是并不推荐在编程中使用。有一个和 cgo 类似的原则：如果可以不使用 goto 语句，那么就不要使用 goto 语句。Go 语言中的 goto 语句是有严格限制的：它无法跨越代码块，并且在被跨越的代码中不能含有变量定义的语句。虽然 Go 语言不推荐 goto 语句，但是 goto 确实每个汇编语言码农的最爱。因为 goto 近似等价于汇编语言中的无条件跳转指令 JMP，配合 if 条件 goto 就组成了有条件跳转指令，而有条件跳转指令正是构建整个汇编代码控制流的基石。

为了便于理解，我们用 Go 语言构造一个模拟三元表达式的 If 函数：

```go
func If(ok bool, a, b int) int {
	if ok {return a} else { return b }
}
```

比如求两个数最大值的三元表达式 `(a>b)?a:b` 用 If 函数可以这样表达：`If(a>b, a, b)`。因为语言的限制，用来模拟三元表达式的 If 函数不支持泛型（可以将 a、b 和返回类型改为空接口，不过使用会繁琐一些）。

这个函数虽然看似只有简单的一行，但是包含了 if 分支语句。在改用汇编实现前，我们还是先用汇编的思维来重新审视 If 函数。在改写时同样要遵循每个表达式只能有一个运算符的限制，同时 if 语句的条件部分必须只有一个比较符号组成，if 语句的 body 部分只能是一个 goto 语句。

用汇编思维改写后的 If 函数实现如下：

```go
func If(ok int, a, b int) int {
	if ok == 0 {goto L}
	return a
L:
	return b
}
```

因为汇编语言中没有 bool 类型，我们改用 int 类型代替 bool 类型（真实的汇编是用 byte 表示 bool 类型，可以通过 MOVBQZX 指令加载 byte 类型的值，这里做了简化处理）。当 ok 参数非 0 时返回变量 a，否则返回变量 b。我们将 ok 的逻辑反转下：当 ok 参数为 0 时，表示返回 b，否则返回变量 a。在 if 语句中，当 ok 参数为 0 时 goto 到 L 标号指定的语句，也就是返回变量 b。如果 if 条件不满足，也就是 ok 参数非 0，执行后面的语句返回变量 a。

上述函数的实现已经非常接近汇编语言，下面是改为汇编实现的代码：

```
TEXT ·If(SB), NOSPLIT, $0-32
	MOVQ ok+8*0(FP), CX // ok
	MOVQ a+8*1(FP), AX  // a
	MOVQ b+8*2(FP), BX  // b

	CMPQ CX, $0         // test ok
	JZ   L              // if ok == 0, goto L
	MOVQ AX, ret+24(FP) // return a
	RET

L:
	MOVQ BX, ret+24(FP) // return b
	RET
```

首先是将三个参数加载到寄存器中，ok 参数对应 CX 寄存器，a、b 分别对应 AX、BX 寄存器。然后使用 CMPQ 比较指令将 CX 寄存器和常数 0 进行比较。如果比较的结果为 0，那么下一条 JZ 为 0 时跳转指令将跳转到 L 标号对应的语句，也就是返回变量 b 的值。如果比较的结果不为 0，那么 JZ 指令将没有效果，继续执行后面的指令，也就是返回变量 a 的值。

在跳转指令中，跳转的目标一般是通过一个标号表示。不过在有些通过宏实现的函数中，更希望通过相对位置跳转，这时候可以通过 PC 寄存器的偏移量来计算临近跳转的位置。

## 3.5.3 for 循环

Go 语言的 for 循环有多种用法，我们这里只选择最经典的 for 结构来讨论。经典的 for 循环由初始化、结束条件、迭代步长三个部分组成，再配合循环体内部的 if 条件语言，这种 for 结构可以模拟其它各种循环类型。

基于经典的 for 循环结构，我们定义一个 LoopAdd 函数，可以用于计算任意等差数列的和：

```go
func LoopAdd(cnt, v0, step int) int {
	result := v0
	for i := 0; i < cnt; i++ {
		result += step
	}
	return result
}
```

比如 `1+2+...+100` 等差数列可以这样计算 `LoopAdd(100, 1, 1)`，而 `10+8+...+0` 等差数列则可以这样计算 `LoopAdd(5, 10, -2)`。在用汇编彻底重写之前先采用前面 `if/goto` 类似的技术来改造 for 循环。

新的 LoopAdd 函数只有 if/goto 语句构成：

```go
func LoopAdd(cnt, v0, step int) int {
	var i = 0
	var result = 0

LOOP_BEGIN:
	result = v0

LOOP_IF:
	if i <cnt { goto LOOP_BODY}
	goto LOOP_END

LOOP_BODY
	i = i+1
	result = result + step
	goto LOOP_IF

LOOP_END:

	return result
}
```

函数的开头先定义两个局部变量便于后续代码使用。然后将 for 语句的初始化、结束条件、迭代步长三个部分拆分为三个代码段，分别用 LOOP_BEGIN、LOOP_IF、LOOP_BODY 三个标号表示。其中 LOOP_BEGIN 循环初始化部分只会执行一次，因此该标号并不会被引用，可以省略。最后 LOOP_END 语句表示 for 循环的结束。四个标号分隔出的三个代码段分别对应 for 循环的初始化语句、循环条件和循环体，其中迭代语句被合并到循环体中了。

下面用汇编语言重新实现 LoopAdd 函数

```
#include "textflag.h"

// func LoopAdd(cnt, v0, step int) int
TEXT ·LoopAdd(SB), NOSPLIT,  $0-32
	MOVQ cnt+0(FP), AX   // cnt
	MOVQ v0+8(FP), BX    // v0/result
	MOVQ step+16(FP), CX // step

LOOP_BEGIN:
	MOVQ $0, DX          // i

LOOP_IF:
	CMPQ DX, AX          // compare i, cnt
	JL   LOOP_BODY       // if i < cnt: goto LOOP_BODY
	JMP LOOP_END

LOOP_BODY:
	ADDQ $1, DX          // i++
	ADDQ CX, BX          // result += step
	JMP LOOP_IF

LOOP_END:

	MOVQ BX, ret+24(FP)  // return result
	RET
```

其中 v0 和 result 变量复用了一个 BX 寄存器。在 LOOP_BEGIN 标号对应的指令部分，用 MOVQ 将 DX 寄存器初始化为 0，DX 对应变量 i，循环的迭代变量。在 LOOP_IF 标号对应的指令部分，使用 CMPQ 指令比较 DX 和 AX，如果循环没有结束则跳转到 LOOP_BODY 部分，否则跳转到 LOOP_END 部分结束循环。在 LOOP_BODY 部分，更新迭代变量并且执行循环体中的累加语句，然后直接跳转到 LOOP_IF 部分进入下一轮循环条件判断。LOOP_END 标号之后就是返回累加结果的语句。

循环是最复杂的控制流，循环中隐含了分支和跳转语句。掌握了循环的写法基本也就掌握了汇编语言的基础写法。更极客的玩法是通过汇编语言打破传统的控制流，比如跨越多层函数直接返回，比如参考基因编辑的手段直接执行一个从 C 语言构建的代码片段等。总之掌握规律之后，你会发现其实汇编语言编程会变得异常简单和有趣。
