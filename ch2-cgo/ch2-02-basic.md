# 2.2 CGO 基础

要使用 CGO 特性，需要安装 C/C++ 构建工具链，在 macOS 和 Linux 下是要安装 GCC，在 windows 下是需要安装 MinGW 工具。同时需要保证环境变量 `CGO_ENABLED` 被设置为 1，这表示 CGO 是被启用的状态。在本地构建时 `CGO_ENABLED` 默认是启用的，当交叉构建时 CGO 默认是禁止的。比如要交叉构建 ARM 环境运行的 Go 程序，需要手工设置好 C/C++ 交叉构建的工具链，同时开启 `CGO_ENABLED` 环境变量。然后通过 `import "C"` 语句启用 CGO 特性。

## 2.2.1 `import "C"` 语句

如果在 Go 代码中出现了 `import "C"` 语句则表示使用了 CGO 特性，紧跟在这行语句前面的注释是一种特殊语法，里面包含的是正常的 C 语言代码。当确保 CGO 启用的情况下，还可以在当前目录中包含 C/C++ 对应的源文件。

举个最简单的例子：

```Go
package main

/*
#include <stdio.h>

void printint(int v) {
	printf("printint: %d\n", v);
}
*/
import "C"

func main() {
	v := 42
	C.printint(C.int(v))
}
```

这个例子展示了 cgo 的基本使用方法。开头的注释中写了要调用的 C 函数和相关的头文件，头文件被 include 之后里面的所有的 C 语言元素都会被加入到”C” 这个虚拟的包中。import "C" 导入语句推荐单独一行，不与其他包一同 import（在 Go 早期版本中必须单独 import）。向 C 函数传递参数也很简单，就直接转化成对应 C 语言类型传递就可以。如上例中 `C.int(v)` 用于将一个 Go 中的 int 类型值强制类型转换转化为 C 语言中的 int 类型值，然后调用 C 语言定义的 printint 函数进行打印。

需要注意的是，Go 是强类型语言，所以 cgo 中传递的参数类型必须与声明的类型完全一致，而且传递前必须用”C” 中的转化函数转换成对应的 C 类型，不能直接传入 Go 中类型的变量。同时通过虚拟的 C 包导入的 C 语言符号并不需要是大写字母开头，它们不受 Go 语言的导出规则约束。

cgo 将当前包引用的 C 语言符号都放到了虚拟的 C 包中，同时当前包依赖的其它 Go 语言包内部可能也通过 cgo 引入了相似的虚拟 C 包，但是不同的 Go 语言包引入的虚拟的 C 包之间的类型是不能通用的。这个约束对于要自己构造一些 cgo 辅助函数时有可能会造成一点的影响。

比如我们希望在 Go 中定义一个 C 语言字符指针对应的 CChar 类型，然后增加一个 GoString 方法返回 Go 语言字符串：

```go
package cgo_helper

//#include <stdio.h>
import "C"

type CChar C.char

func (p *CChar) GoString() string {
	return C.GoString((*C.char)(p))
}

func PrintCString(cs *C.char) {
	C.puts(cs)
}
```

现在我们可能会想在其它的 Go 语言包中也使用这个辅助函数：

```go
package main

//static const char* cs = "hello";
import "C"
import "./cgo_helper"

func main() {
	cgo_helper.PrintCString(C.cs)
}
```

这段代码是不能正常工作的，因为当前 main 包引入的 `C.cs` 变量的类型是当前 `main` 包的 cgo 构造的虚拟的 C 包下的 `*char` 类型（具体点是 `*C.char`，更具体点是 `*main.C.char`），它和 cgo_helper 包引入的 `*C.char` 类型（具体点是 `*cgo_helper.C.char`）是不同的。在 Go 语言中方法是依附于类型存在的，不同 Go 包中引入的虚拟的 C 包的类型却是不同的（`main.C` 不等 `cgo_helper.C`），这导致从它们延伸出来的 Go 类型也是不同的类型（`*main.C.char` 不等 `*cgo_helper.C.char`），这最终导致了前面代码不能正常工作。

有 Go 语言使用经验的用户可能会建议参数转型后再传入。但是这个方法似乎也是不可行的，因为 `cgo_helper.PrintCString` 的参数是它自身包引入的 `*C.char` 类型，在外部是无法直接获取这个类型的。换言之，一个包如果在公开的接口中直接使用了 `*C.char` 等类似的虚拟 C 包的类型，其它的 Go 包是无法直接使用这些类型的，除非这个 Go 包同时也提供了 `*C.char` 类型的构造函数。因为这些诸多因素，如果想在 go test 环境直接测试这些 cgo 导出的类型也会有相同的限制。

<!-- 测试代码；需要确实是否有问题 -->

## 2.2.2 `#cgo` 语句

在 `import "C"` 语句前的注释中可以通过 `#cgo` 语句设置编译阶段和链接阶段的相关参数。编译阶段的参数主要用于定义相关宏和指定头文件检索路径。链接阶段的参数主要是指定库文件检索路径和要链接的库文件。

```go
// #cgo CFLAGS: -DPNG_DEBUG=1 -I./include
// #cgo LDFLAGS: -L/usr/local/lib -lpng
// #include <png.h>
import "C"
```

上面的代码中，CFLAGS 部分，`-D` 部分定义了宏 PNG_DEBUG，值为 1；`-I` 定义了头文件包含的检索目录。LDFLAGS 部分，`-L` 指定了链接时库文件检索目录，`-l` 指定了链接时需要链接 png 库。


因为 C/C++ 遗留的问题，C 头文件检索目录可以是相对目录，但是库文件检索目录则需要绝对路径。在库文件的检索目录中可以通过 `${SRCDIR}` 变量表示当前包目录的绝对路径：

```
// #cgo LDFLAGS: -L${SRCDIR}/libs -lfoo
```

上面的代码在链接时将被展开为：

```
// #cgo LDFLAGS: -L/go/src/foo/libs -lfoo
```

`#cgo` 语句主要影响 CFLAGS、CPPFLAGS、CXXFLAGS、FFLAGS 和 LDFLAGS 几个编译器环境变量。LDFLAGS 用于设置链接时的参数，除此之外的几个变量用于改变编译阶段的构建参数 (CFLAGS 用于针对 C 语言代码设置编译参数)。

对于在 cgo 环境混合使用 C 和 C++ 的用户来说，可能有三种不同的编译选项：其中 CFLAGS 对应 C 语言特有的编译选项、CXXFLAGS 对应是 C++ 特有的编译选项、CPPFLAGS 则对应 C 和 C++ 共有的编译选项。但是在链接阶段，C 和 C++ 的链接选项是通用的，因此这个时候已经不再有 C 和 C++ 语言的区别，它们的目标文件的类型是相同的。

`#cgo` 指令还支持条件选择，当满足某个操作系统或某个 CPU 架构类型时后面的编译或链接选项生效。比如下面是分别针对 windows 和非 windows 下平台的编译和链接选项：

```
// #cgo windows CFLAGS: -DX86=1
// #cgo !windows LDFLAGS: -lm
```

其中在 windows 平台下，编译前会预定义 X86 宏为 1；在非 windows 平台下，在链接阶段会要求链接 math 数学库。这种用法对于在不同平台下只有少数编译选项差异的场景比较适用。

如果在不同的系统下 cgo 对应着不同的 c 代码，我们可以先使用 `#cgo` 指令定义不同的 C 语言的宏，然后通过宏来区分不同的代码：

```go
package main

/*
#cgo windows CFLAGS: -DCGO_OS_WINDOWS=1
#cgo darwin CFLAGS: -DCGO_OS_DARWIN=1
#cgo linux CFLAGS: -DCGO_OS_LINUX=1

#if defined(CGO_OS_WINDOWS)
	const char* os = "windows";
#elif defined(CGO_OS_DARWIN)
	const char* os = "darwin";
#elif defined(CGO_OS_LINUX)
	const char* os = "linux";
#else
#	error(unknown os)
#endif
*/
import "C"

func main() {
	print(C.GoString(C.os))
}
```

这样我们就可以用 C 语言中常用的技术来处理不同平台之间的差异代码。

## 2.2.3 build tag 条件编译

build tag 是在 Go 或 cgo 环境下的 C/C++ 文件开头的一种特殊的注释。条件编译类似于前面通过 `#cgo` 指令针对不同平台定义的宏，只有在对应平台的宏被定义之后才会构建对应的代码。但是通过 `#cgo` 指令定义宏有个限制，它只能是基于 Go 语言支持的 windows、darwin 和 linux 等已经支持的操作系统。如果我们希望定义一个 DEBUG 标志的宏，`#cgo` 指令就无能为力了。而 Go 语言提供的 build tag 条件编译特性则可以简单做到。

比如下面的源文件只有在设置 debug 构建标志时才会被构建：

```go
// +build debug

package main

var buildMode = "debug"
```

可以用以下命令构建：

```
go build -tags="debug"
go build -tags="windows debug"
```

我们可以通过 `-tags` 命令行参数同时指定多个 build 标志，它们之间用空格分隔。

当有多个 build tag 时，我们将多个标志通过逻辑操作的规则来组合使用。比如以下的构建标志表示只有在”linux/386“或”darwin 平台下非 cgo 环境 “才进行构建。

```go
// +build linux,386 darwin,!cgo
```

其中 `linux,386` 中 linux 和 386 用逗号链接表示 AND 的意思；而 `linux,386` 和 `darwin,!cgo` 之间通过空白分割来表示 OR 的意思。
