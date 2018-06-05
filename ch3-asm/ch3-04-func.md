# 3.4. 函数(Doing)

终于到函数了！因为Go汇编语言中，可以也建议通过Go语言来定义全局变量，那么剩下的也就是函数了。只有掌握了汇编函数的基本用法，才能真正算是Go汇编语言入门。本章将简单讨论Go汇编中函数的定义和用法。

## 基本语法

函数标识符通过TEXT汇编指令定义，表示该行开始的指令定义在TEXT内存段。TEXT语句后的指令一般对应函数的实现，但是对于TEXT指令本身来说并不关心后面是否有指令。我个人绝对TEXT和LABEL定义的符号是类似的，区别只是LABEL是用于跳转标号，但是本质上他们都是通过标识符映射一个内存地址。

函数的定义的语法如下：

```
TEXT symbol(SB), [flags,] $framesize[-argsize]
```

函数的定义部分由5个部分组成：TEXT指令、函数名、可选的flags标志、函数帧大小和可选的函数参数大小。

其中Text用于定义函数符号，函数名中当前包的路径可以省略。函数的名字后面是`(SB)`，表示是相对于的函数名符号对相对于SB伪寄存器的偏移量，二者组合在一起最终是绝对地址。作为全局的标识符的全局变量和全局函数的名字一般都是基于SB伪寄存器的相对地址。标志部分用于指示函数的一些特殊行为，常见的NOSPLIT主要用于指示叶子函数不进行栈分裂。framesize部分表示函数的局部变量需要多少栈空间，其中包含调用其它函数是准备调用参数的隐式栈空间。最后是可以省略的参数大小，之所以可以省略是因为编译器可以从Go语言的函数声明中推导出函数参数的大小。

下面是在main包中Add在汇编中两种定义方式：

```
// func Add(a, b int) int
TEXT main·Add(SB), NOSPLIT, $0-24

// func Add(a, b int) int
TEXT ·Add(SB), $0
```

第一种是最完整的写法：函数名部分包含了当前包的路径，同时指明了函数的参数大小为24个字节（对应参数和返回值的3个int类型）。第二种写法则比较简洁，省略了当前包的路径和参数的大小。需要注意的是，标志参数中的NOSPLIT如果在Go语言函数声明中通过注释指明了标志，应该也是可以省略的（需要确认下）。

目前可能遇到的函数函数标志有NOSPLIT、WRAPPER和NEEDCTXT几个。其中NOSPLIT不会生成或包含栈分裂代码，这一般用于没有任何其它函数调用的叶子函数，这样可以适当提高性能。WRAPPER标志则表示这个是一个包装函数，在panic或runtime.caller等某项处理函数帧的地方不会增加函数帧计数。最后的NEEDCTXT表示需要一个上下午参数，一般用于闭包函数。

需要注意的是函数也没有类型，上面定义的Add函数签名可以下面任意一种格式：

```
func Add(a, b int) int
func Add(a, b, c int)
func Add() (a, b, c int)
func Add() (a []int) // reflect.SliceHeader 切片头刚好也是 3 个 int 成员
// ...
```

对于汇编函数来说，只要是函数的名字和参数大小一致就可以是相同的函数了。而且在Go汇编语言中，输入参数和返回值参数是没有任何的区别的。

## 函数参数和返回值

对于函数来说，最重要是是函数对外提供的API约定，包含函数的名称、参数和返回值。当名称和参数返回都确定之后，如何精确计算参数和返回值的大小是第一个需要解决的问题。

比如有一个Foo函数的签名如下：

```go
func Foo(a, b int) (c int)
```

对于这个函数，我们可以轻易看出它需要3个int类型的空间，参数和返回值的大小也就是24个字节：

```
TEXT ·Foo(SB), $0-24
```

那么如何在汇编中引用这3个参数呢？为此Go汇编中引入了一个FP伪寄存器，表示函数当前帧的地址，也就是第一个参数的地址。因此我们以通过`+0(FP)`、`+8(FP)`和`+16(FP)`来分别引用a、b、c三个参数。

但是在汇编代码中，我们并不能直接使用`+0(FP)`来使用参数。为了编写易于维护的汇编代码，Go汇编语言要求，任何通过FP寄存器访问的变量必和一个临时标识符前缀组合后才能有效，一般使用参数对应的变量名作为前缀。

下面的代码演示了如何在汇编函数中使用参数和返回值：

```
TEXT ·Foo(SB), $0
	MOVEQ a+0(FP), AX  // a
	MOVEQ b+8(FP), BX  // b
	MOVEQ c+16(FP), CX // c
	RET
```

如果是参数和返回值类型比较复杂的情况改如何处理呢？下面我们再尝试一个更复杂的函数参数和返回值的计算。比如有以下一个函数：

```go
func SomeFunc(a, b int, c bool) (d float64, err error) int
```

函数的参数有不同的类型，同时含义多个返回值，而且返回值中含有更复杂的接口类型。我们该如何计算每个参数的位置和总的大小呢？

其实函数参数和返回值的大小以及对齐问题和结构体的大小和成员对齐问题是一致的。我们先看看如果用Go语言函数来模拟Foo函数中参数和返回值的地址：

```go
func Foo(FP *struct{a, b, c int}) {
	_ = unsafe.Offsetof(FP.a) + uintptr(FP) // a
	_ = unsafe.Offsetof(FP.b) + uintptr(FP) // b
	_ = unsafe.Offsetof(FP.c) + uintptr(FP) // c

	_ = unsafe.Sizeof(*FP) // argsize

	return
}
```

我们尝试将全部的参数和返回值以同样的顺序放到一个结构体中，将FP伪寄存器作为唯一的一个指针参数，而每个成员的地址也就是对应原来参数的地址。

用同样的策略可以很容易计算前面的SomeFunc函数的参数和返回值的地址和总大小。

因为SomeFunc函数的参数比较多，我们临时定一个`SomeFunc_args_and_returns`结构体用于对应参数和返回值：

```go
type SomeFunc_args_and_returns struct {
	a int
	b int
	c bool
	d float64
	e error
}
```

然后将SomeFunc原来的参数替换为结构体形式，并且只保留唯一的FP作为参数：

```go
func SomeFunc(FP *SomeFunc_args_and_returns) {
	_ = unsafe.Offsetof(FP.a) + uintptr(FP) // a
	_ = unsafe.Offsetof(FP.b) + uintptr(FP) // b
	_ = unsafe.Offsetof(FP.c) + uintptr(FP) // c
	_ = unsafe.Offsetof(FP.d) + uintptr(FP) // d
	_ = unsafe.Offsetof(FP.e) + uintptr(FP) // e

	_ = unsafe.Sizeof(*FP) // argsize

	return
}
```

代码完全和Foo函数参数的方式类似。唯一的差异是每个函数的偏移量，这有`unsafe.Offsetof`函数自动计算生成。因为Go结构体中的每个成员已经满足了对齐要求，因此采用通用方式得到每个参数的偏移量也是满足对齐要求的。


## 函数中的局部变量

TODO

## 函数中的(真)(伪)SP

<!-- 伪SP是当前func栈的起点，不会发生变化，便于局部变量定位 -->

TODO

## 调用其它函数

<!-- 基于栈传参和返回值，性能稍低，但是栈帧非常规整，好分析 -->


TODO

## 宏函数

<!-- swap 方法，几个版本 -->

TODO
