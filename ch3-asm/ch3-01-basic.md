# 3.1. 快速入门

在第一章的“Hello, World 的革命”一节中，我们已经见过一个Go汇编程序。本节我们将通过分析简单的Go程序输出的汇编代码，然后照猫画虎用汇编实现一个简单的输出程序。

## 实现和声明

Go汇编语言并不是一个独立的语言，主要原因是因为Go汇编程序无法独立使用。Go汇编代码必须以Go包的方式被组织，同时包中至少要有一个Go语言文件。如果Go汇编代码中定义的变量和函数要被其它Go语言代码引用，还需要通过Go语言代码将汇编中定义的符号声明出来。用于变量的定义和函数的定义Go汇编文件类似于C语言中的.c文件。而用于导出汇编中定义符号的Go源文件类似于C语言的.h文件。

## 定义整数变量

为了简单，我们先用Go语言定义并赋值一个整数变量，然后查看生成的汇编代码。

创建pkg.go文件，内容如下：

```go
package pkg

var Id = 9527
```

然后用以下命令查看的Go语言程序对应的伪汇编代码：

```
$ go tool compile -S pkg.go
"".Id SNOPTRDATA size=8
        0x0000 37 25 00 00 00 00 00 00                          '.......
```

输出的汇编比较简单，其中`"".Id`对应Id变量符号，变量的内存大小为8个字节。变量的初始化内容为`37 25 00 00 00 00 00 00`，对应十六进制格式的0x2537，对应十进制为9527。SNOPTRDATA是相关的标志，暂时忽略。

以上的内容只是目标文件对于的汇编，和Go汇编语言虽然相似当并不完全等价。Go语言官网自带了一个Go汇编语言的入门教程，地址在：https://golang.org/doc/asm。

Go汇编语言提供了DATA命令用于初始化变量，DATA命令的语法如下：

```
DATA symbol+offset(SB)/width, value
```

其中symbol为变量在汇编语言中对应的符号，offset是符号开始地址的偏移量，width是要初始化内存的宽度大小，value是要初始化的那天。其中当前包中Go语言定义的符号symbol，在汇编代码中对应`·symbol`，其中·为一个特殊的unicode符号。

采用以下命令可以给Id变量初始化为十六进制的0x2537，对应十进制的9527，常量需要以美元符号$开头表示：

```
DATA ·Id+0(SB)/1,$0x37
DATA ·Id+1(SB)/1,$0x25
```

变量定义好之后需要导出以共其它代码引用。Go汇编语言提供了GLOBL命令用于将符号导出：

```
GLOBL symbol(SB), width
```

其中symbol对应汇编中符号的名字，width为符号对应内存的大小。用以下命令将汇编中的·Id变量导出：

```
GLOBL ·Id, $8
```

现在已经出版完成了用汇编定义一个整数变量的工作。

为了便于其它包使用该Id变量，我们还需要在Go代码中声明该变量，同时也给变量指定一个合适的类型。修改pkg.go的内容如下：

```go
package pkg

var Id int
```

表示声明一个一个int类型的Id变量。因为该变量已经在汇编中定义，因此Go语言部分只是声明变量，声明的变量不能含义初始化的操作。

完整的汇编代码在pkg_amd64.s中：

```
GLOBL ·Id(SB),$8

DATA ·Id+0(SB)/1,$0x37
DATA ·Id+1(SB)/1,$0x25
DATA ·Id+2(SB)/1,$0x00
DATA ·Id+3(SB)/1,$0x00
DATA ·Id+4(SB)/1,$0x00
DATA ·Id+5(SB)/1,$0x00
DATA ·Id+6(SB)/1,$0x00
DATA ·Id+7(SB)/1,$0x00
```

文件名pkg_amd64.s表示为AMD64环境下的汇编代码文件。

虽然pkg包改用汇编实现，但是用法和之前完全一样：

```go
package main

import pkg "pkg包的路径"

func main() {
	println(pkg.Id)
}
```

对于Go包的用户来说，用Go汇编语言或Go语言实现并无区别。

## 定义字符串变量

TODO

## Go语言版本

```go
package pkg

var helloworld = "Hello World!"

func HelloWorld() {
	println(helloworld)
}
```

## Go汇编版本

```
#include "textflag.h"

// var helloworld string
GLOBL ·helloworld(SB),NOPTR,$32                  // var helloworld [32]byte
	DATA ·helloworld+0(SB)/8,$·helloworld+16(SB) // StringHeader.Data
	DATA ·helloworld+8(SB)/8,$12                 // StringHeader.Len
	DATA ·helloworld+16(SB)/8,$"Hello Wo"        // ...string data...
	DATA ·helloworld+24(SB)/8,$"rld!"            // ...string data...

// func HelloWorld()
TEXT ·HelloWorld(SB), $16-0
	MOVQ ·helloworld+0(SB), AX; MOVQ AX, 0(SP)
	MOVQ ·helloworld+8(SB), BX; MOVQ BX, 8(SP)
	CALL runtime·printstring(SB)
	CALL runtime·printnl(SB)
	RET
```

## 汇编语法

- 变量要在Go语言中声明, 但不能赋值
- 函数要在Go语言中声明, 但不包含函数实现

- Go语言中的标识符x对应汇编语言中的·x

- GLOBL: 定义全局标识符, 分配内存空间
- DATA: 初始化对应内存空间
- TEXT: 定义函数

## 字符串的结构

```go
var helloworld string // 只能声明, 不能赋值
```

```
// +---------------------------+              ·helloworld+0(SB)
// | reflect.StringHeader.Data | ----------\ $·helloworld+16(SB)
// +---------------------------+           |
// | reflect.StringHeader.Len  |           |
// +---------------------------+ <---------/  ·helloworld+16(SB)
// | "Hello World!"            |
// +---------------------------+
```

- 字符串的数据紧挨字符串头结构体
- $·helloworld+16(SB) 表示符号地址
- ·helloworld+16(SB) 表示符号地址内的数据

## HelloWorld函数

```go
func HelloWorld() // 只能声明, 不能定义
```

```
TEXT ·HelloWorld(SB), $16-0
	MOVQ ·helloworld+0(SB), AX; MOVQ AX, 0(SP)
	MOVQ ·helloworld+8(SB), BX; MOVQ BX, 8(SP)
	CALL runtime·printstring(SB)
	CALL runtime·printnl(SB)
	RET
```

- $16-0中的16: 表示函数内部有16字节用于局部变量
- $16-0中的0: 表示函数参数和返回值总大小为0

- printstring的参数类型为StringHeader
- 0(SP)为StringHeader.Data
- 8(SP)为StringHeader.Len

## 简化: 在Go中定义变量

```go
var helloworld string = "你好, 中国!"

func HelloWorld()
```

```
TEXT ·HelloWorld(SB), $16-0
	MOVQ ·helloworld+0(SB), AX; MOVQ AX, 0(SP)
	MOVQ ·helloworld+8(SB), BX; MOVQ BX, 8(SP)
	CALL runtime·printstring(SB)
	CALL runtime·printnl(SB)
	RET
```

- 汇编定义变量没有太多优势, 性价比较低
- 汇编的优势是挖掘芯片的功能和性能

## 特殊字符

TODO

## 没有分号

- 分号用于分隔多个汇编语句
- 行末尾自动添加分号

<!-- 宏中的分号和注释 -->

