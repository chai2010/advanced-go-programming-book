# 3.1. 快速入门

在第一章的“Hello, World 的革命”一节中，我们已经见过一个Go汇编程序。本节我们将通过由浅入深的一系列小例子来快速掌握Go汇编的简单用法。

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

## 没有分号

- 分号用于分隔多个汇编语句
- 行末尾自动添加分号

<!-- 宏中的分号和注释 -->

