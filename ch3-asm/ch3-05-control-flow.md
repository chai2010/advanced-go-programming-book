# 3.5 控制流

程序主要有顺序、分支和循环几种执行流程。本节主要讨论如何将Go语言的控制流比较直观地转译为汇编程序，或者说如何以汇编思维来编写Go语言代码。

## 3.5.1 顺序执行

顺序执行是我们比较熟悉的工作模式，类似俗称流水账编程。所有不含分支、循环和goto语句，并且没有递归调用的Go函数一般都是顺序执行的。

比如有如下顺序执行的代码：

```go
func main() {
	var a = 10
	println(a)

	var b = (a+a)*a
	println(b)
}
```

我们尝试用Go汇编的思维改写上述函数。因为X86指令中一般只有2个操作数，因此在用汇编改写时要求出现的变量表达式中最多只能有一个运算符。同时对于一些函数调用，也需要用汇编中可以调用的函数来改写。

第一步改写依然是使用Go语言，只不过是用汇编的思维改写：

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

首选模仿C语言的处理方式在函数入口处声明全部的局部变量。然后根据MOV、ADD、MUL等指令的风格，将之前的变量表达式展开为用`=`、`+=`和`*=`几种运算表达的多个指令。最后用runtime包内部的printint和printnl函数代替之前的println函数输出结果。

经过用汇编的思维改写过后，上述的Go函数虽然看着繁琐了一点，但是还是比较容易理解的。下面我们进一步尝试将改写后的函数继续转译为汇编函数：

```
TEXT ·main(SB), $24-0
	MOVQ $0, a-8*2(SP) // a = 0
	MOVQ $0, b-8*1(SP) // b = 0

	// 将新的值写入a对应内存
	MOVQ $10, AX       // AX = 10
	MOVQ AX, a-8*2(SP) // a = AX

	// 以a为参数调用函数
	MOVQ AX, 0(SP)
	CALL runtime·printint(SB)
	CALL runtime·printnl(SB)

	// 函数调用后, AX/BX 寄存器可能被污染, 需要重新加载
	MOVQ a-8*2(SP), AX // AX = a
	MOVQ b-8*1(SP), BX // BX = b

	// 计算b值, 并写入内存
	MOVQ AX, BX        // BX = AX  // b = a
	ADDQ BX, BX        // BX += BX // b += a
	IMULQ AX, BX       // BX *= AX // b *= a
	MOVQ BX, b-8*1(SP) // b = BX

	// 以b为参数调用函数
	MOVQ BX, 0(SP)
	CALL runtime·printint(SB)
	CALL runtime·printnl(SB)

	RET
```

汇编实现main函数的第一步是要计算函数栈帧的大小。因为函数内有a、b两个int类型变量，同时调用的runtime·printint函数参数是一个int类型并且没有返回值，因此main函数的栈帧是3个int类型组成的24个字节的栈内存空间。

在函数的开始处先将变量初始化为0值，其中`a-8*2(SP)`对应a变量、`a-8*1(SP)`对应b变量（因为a变量先定义，因此a变量的地址更小）。

然后给a变量分配一个AX寄存器，并且通过AX寄存器将a变量对应的内存设置为10，AX也是10。为了输出a变量，需要将AX寄存器的值放到`0(SP)`位置，这个位置的变量将在调用runtime·printint函数时作为它的参数被打印。因为我们之前已经将AX的值保存到a变量内存中了，因此在调用函数前并不需要再进行寄存器的备份工作。

在调用函数返回之后，全部的寄存器将被视为可能被调用的函数修改，因此我们需要从a、b对应的内存中重新恢复寄存器AX和BX。然后参考上面Go语言中b变量的计算方式更新BX对应的值，计算完成后同样将BX的值写入到b对应的内存。

需要说明的是，上面的代码中`IMULQ AX, BX`使用了`IMULQ`指令来计算乘法。没有使用`MULQ`指令的原因是`MULQ`指令默认使用`AX`保存结果。读者可以自己尝试用`MULQ`指令改写上述代码。

最后以b变量作为参数再次调用runtime·printint函数进行输出工作。所有的寄存器同样可能被污染，不过main函数马上就返回了，因此不再需要恢复AX、BX等寄存器了。

重新分析汇编改写后的整个函数会发现里面很多的冗余代码。我们并不需要a、b两个临时变量分配两个内存空间，而且也不需要在每个寄存器变化之后都要写入内存。下面是经过优化的汇编函数：

```
TEXT ·main(SB), $16-0
	// var temp int

	// 将新的值写入a对应内存
	MOVQ $10, AX        // AX = 10
	MOVQ AX, temp-8(SP) // temp = AX

	// 以a为参数调用函数
	CALL runtime·printint(SB)
	CALL runtime·printnl(SB)

	// 函数调用后, AX 可能被污染, 需要重新加载
	MOVQ temp-8*1(SP), AX // AX = temp

	// 计算b值, 不需要写入内存
	MOVQ AX, BX        // BX = AX  // b = a
	ADDQ BX, BX        // BX += BX // b += a
	IMULQ AX, BX       // BX *= AX // b *= a

	// ...
```

首先是将main函数的栈帧大小从24字节减少到16字节。唯一需要保存的是a变量的值，因此在调用runtime·printint函数输出时全部的寄存器都可能被污染，我们无法通过寄存器备份a变量的值，只有在栈内存中的值才是安全的。然后在BX寄存器并不需要保存到内存。其它部分的代码基本保持不变。

## 3.5.2 if/goto跳转

Go语言刚刚开源的时候并没有goto语句，后来Go语言虽然增加了goto语句，但是并不推荐在编程中使用。有一个和cgo类似的原则：如果可以不使用goto语句，那么就不要使用goto语句。Go语言中的goto语句是有严格限制的：它无法跨越代码块，并且在被跨越的代码中不能含有变量定义的语句。虽然Go语言不推荐goto语句，但是goto确实每个汇编语言码农的最爱。因为goto近似等价于汇编语言中的无条件跳转指令JMP，配合if条件goto就组成了有条件跳转指令，而有条件跳转指令正是构建整个汇编代码控制流的基石。

为了便于理解，我们用Go语言构造一个模拟三元表达式的If函数：

```go
func If(ok bool, a, b int) int {
	if ok { return a } else { return b }
}
```

比如求两个数最大值的三元表达式`(a>b)?a:b`用If函数可以这样表达：`If(a>b, a, b)`。因为语言的限制，用来模拟三元表达式的If函数不支持泛型（可以将a、b和返回类型改为空接口，不过使用会繁琐一些）。

这个函数虽然看似只有简单的一行，但是包含了if分支语句。在改用汇编实现前，我们还是先用汇编的思维来重新审视If函数。在改写时同样要遵循每个表达式只能有一个运算符的限制，同时if语句的条件部分必须只有一个比较符号组成，if语句的body部分只能是一个goto语句。

用汇编思维改写后的If函数实现如下：

```go
func If(ok int, a, b int) int {
	if ok == 0 { goto L }
	return a
L:
	return b
}
```

因为汇编语言中没有bool类型，我们改用int类型代替bool类型（真实的汇编是用byte表示bool类型，可以通过MOVBQZX指令加载byte类型的值，这里做了简化处理）。当ok参数非0时返回变量a，否则返回变量b。我们将ok的逻辑反转下：当ok参数为0时，表示返回b，否则返回变量a。在if语句中，当ok参数为0时goto到L标号指定的语句，也就是返回变量b。如果if条件不满足，也就是ok参数非0，执行后面的语句返回变量a。

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

首先是将三个参数加载到寄存器中，ok参数对应CX寄存器，a、b分别对应AX、BX寄存器。然后使用CMPQ比较指令将CX寄存器和常数0进行比较。如果比较的结果为0，那么下一条JZ为0时跳转指令将跳转到L标号对应的语句，也就是返回变量b的值。如果比较的结果不为0，那么JZ指令将没有效果，继续执行后面的指令，也就是返回变量a的值。

在跳转指令中，跳转的目标一般是通过一个标号表示。不过在有些通过宏实现的函数中，更希望通过相对位置跳转，这时候可以通过PC寄存器的偏移量来计算临近跳转的位置。

## 3.5.3 for循环

Go语言的for循环有多种用法，我们这里只选择最经典的for结构来讨论。经典的for循环由初始化、结束条件、迭代步长三个部分组成，再配合循环体内部的if条件语言，这种for结构可以模拟其它各种循环类型。

基于经典的for循环结构，我们定义一个LoopAdd函数，可以用于计算任意等差数列的和：

```go
func LoopAdd(cnt, v0, step int) int {
	result := v0
	for i := 0; i < cnt; i++ {
		result += step
	}
	return result
}
```

比如`1+2+...+100`等差数列可以这样计算`LoopAdd(100, 1, 1)`，而`10+8+...+0`等差数列则可以这样计算`LoopAdd(5, 10, -2)`。在用汇编彻底重写之前先采用前面`if/goto`类似的技术来改造for循环。

新的LoopAdd函数只有if/goto语句构成：

```go

func LoopAdd(cnt, v0, step int) int {
	var i = 0
	var result = 0

LOOP_BEGIN:
	result = v0

LOOP_IF:
	if i < cnt { goto LOOP_BODY }
	goto LOOP_END

LOOP_BODY
	i = i+1
	result = result + step
	goto LOOP_IF

LOOP_END:

	return result
}
```

函数的开头先定义两个局部变量便于后续代码使用。然后将for语句的初始化、结束条件、迭代步长三个部分拆分为三个代码段，分别用LOOP_BEGIN、LOOP_IF、LOOP_BODY三个标号表示。其中LOOP_BEGIN循环初始化部分只会执行一次，因此该标号并不会被引用，可以省略。最后LOOP_END语句表示for循环的结束。四个标号分隔出的三个代码段分别对应for循环的初始化语句、循环条件和循环体，其中迭代语句被合并到循环体中了。

下面用汇编语言重新实现LoopAdd函数

```
// func LoopAdd(cnt, v0, step int) int
TEXT ·LoopAdd(SB), NOSPLIT, $0-32
	MOVQ cnt+0(FP), AX   // cnt
	MOVQ v0+8(FP), BX    // v0/result
	MOVQ step+16(FP), CX // step

LOOP_BEGIN:
	MOVQ $0, DX          // i

LOOP_IF:
	CMPQ DX, AX          // compare i, cnt
	JL   LOOP_BODY       // if i < cnt: goto LOOP_BODY
	goto LOOP_END

LOOP_BODY:
	ADDQ $1, DX          // i++
	ADDQ CX, BX          // result += step
	goto LOOP_IF

LOOP_END:

	MOVQ BX, ret+24(FP)  // return result
	RET
```

其中v0和result变量复用了一个BX寄存器。在LOOP_BEGIN标号对应的指令部分，用MOVQ将DX寄存器初始化为0，DX对应变量i，循环的迭代变量。在LOOP_IF标号对应的指令部分，使用CMPQ指令比较DX和AX，如果循环没有结束则跳转到LOOP_BODY部分，否则跳转到LOOP_END部分结束循环。在LOOP_BODY部分，更新迭代变量并且执行循环体中的累加语句，然后直接跳转到LOOP_IF部分进入下一轮循环条件判断。LOOP_END标号之后就是返回累加结果的语句。

循环是最复杂的控制流，循环中隐含了分支和跳转语句。掌握了循环的写法基本也就掌握了汇编语言的基础写法。更极客的玩法是通过汇编语言打破传统的控制流，比如跨越多层函数直接返回，比如参考基因编辑的手段直接执行一个从C语言构建的代码片段等。总之掌握规律之后，你会发现其实汇编语言编程会变得异常简单和有趣。
