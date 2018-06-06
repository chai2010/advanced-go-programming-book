# 3.5. 控制流(Doing)

程序执行的流程主要有顺序、分支和循环几种执行流程。本节主要讨论如何将Go语言的控制流比较直观地转译为汇编程序，或者说如何以汇编思维来编写Go语言代码。

## 顺序执行

顺序执行是我们比较熟悉的工作模式，类似俗称流水账编程。所有不含分支、循环和goto语言，并且每一递归调用的Go函数一般都是顺序执行的。

比如有如下顺序执行的代码：

```go
func main() {
	var a = 10
	println(a)

	var b = (a+a)*a
	println(b)
}
```

我们尝试用Go汇编的思维改写上述函数。因为X86指令中一般只有2个操作数，因此在用汇编改写时要求出现的变量表达式中最多只能有一个运算符。同时对于一些函数调用，也需要该用汇编中可以调用的函数来改写。

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

首选模仿C语言的处理方式在函数入口出声明全部的局部变量。然后将根据MOV、ADD、MUL等指令的风格，将之前的变量表达式展开为用`=`、`+=`和`*=`几种运算表达的多个指令。最后用runtime包内部的printint和printnl函数代替之前的println函数输出结果。

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
	CALL runtime·printint
	CALL runtime·printnl

	// 函数调用后, AX/BX 可能被污染, 需要重新加载
	MOVQ a-8*2(SP), AX // AX = a
	MOVQ b-8*1(SP), BX // BX = b

	// 计算b值, 并写入内存
	MOVQ AX, BX        // BX = AX  // b = a
	ADDQ BX, BX        // BX += BX // b += a
	MULQ AX, BX        // BX *= AX // b *= a
	MOVQ BX, b-8*1(SP) // b = BX

	// 以b为参数调用函数
	MOVQ BX, 0(SP)
	CALL runtime·printint
	CALL runtime·printnl

	RET
```

汇编实现main函数的第一步是要计算函数栈帧的大小。因为函数内有a、b两个int类型变量，同时调用的runtime·printint函数参数是一个int类型并且没有返回值，因此main函数的栈帧是3个int类型组成的24个字节的栈内存空间。

在函数的开始处先将变量初始化为0值，其中`a-8*2(SP)`对应a变量、`a-8*1(SP)`对应b变量（因为a变量先定义，因此a变量的地址更小）。

然后给a变量分配一个AX寄存器，并且通过AX寄存器将a变量对应的内存设置为10，AX也是10。为了输出a变量，需要将AX寄存器的值放到`0(SP)`位置，这个位置的变量将在调用runtime·printint函数时作为它的参数被打印。因为我们之前已经将AX的值保存到a变量内存中了，因此在调用函数前并不需要在进行寄存器的备份工作。

在调用函数返回之后，全部的寄存器将被视为被调用的函数修改，因此我们需要从a、b对应的内存中重新恢复寄存器AX和BX。然后参考上面Go语言中b变量的计算方式更新BX对应的值，计算完成后同样将BX的值写入到b对应的内存。

最后以b变量作为参数再次调用runtime·printint函数进行输出工作。所有的寄存器通样可能被污染，不过main马上就返回不在需要使用AX、BX等寄存器，因此就不需要再次恢复寄存器的值了。

重新分析汇编改写后的整个函数会发现里面很多的冗余代码。我们并不需要a、b两个临时变量分配两个内存空间，而且也不需要在每个寄存器变化之后都要写入内存。下面是经过优化的汇编函数：

```
TEXT ·main(SB), $16-0
	// var temp int

	// 将新的值写入a对应内存
	MOVQ $10, AX        // AX = 10
	MOVQ AX, temp-8(SP) // temp = AX

	// 以a为参数调用函数
	// ...

	// 函数调用后, AX 可能被污染, 需要重新加载
	MOVQ temp-8*1(SP), AX // AX = temp

	// 计算b值, 不需要写入内存
	MOVQ AX, BX        // BX = AX  // b = a
	ADDQ BX, BX        // BX += BX // b += a
	MULQ AX, BX        // BX *= AX // b *= a

	// ...
```

首先是将main函数的栈帧大小从24字节减少到16字节。唯一需要保存的是a变量的值，因此在调用runtime·printint函数输出时全部的寄存器都可能被污染，我们无法通过寄存器备份a变量的值，只有在栈内存中的值才是安全的。然后在BX寄存器并不需要保存到内存。其它部分的代码基本保持不变。

## goto跳转

TODO

## if分支

TODO

## for循环

TODO
