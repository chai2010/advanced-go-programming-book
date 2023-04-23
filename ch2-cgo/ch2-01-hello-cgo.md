# 2.1 快速入门

本节我们将通过一系列由浅入深的小例子来快速掌握 CGO 的基本用法。

## 2.1.1 最简 CGO 程序

真实的 CGO 程序一般都比较复杂。不过我们可以由浅入深，一个最简的 CGO 程序该是什么样的呢？要构造一个最简 CGO 程序，首先要忽视一些复杂的 CGO 特性，同时要展示 CGO 程序和纯 Go 程序的差别来。下面是我们构建的最简 CGO 程序：

```go
// hello.go
package main

import "C"

func main() {
	println("hello cgo")
}
```

代码通过 `import "C"` 语句启用 CGO 特性，主函数只是通过 Go 内置的 println 函数输出字符串，其中并没有任何和 CGO 相关的代码。虽然没有调用 CGO 的相关函数，但是 `go build` 命令会在编译和链接阶段启动 gcc 编译器，这已经是一个完整的 CGO 程序了。

## 2.1.2 基于 C 标准库函数输出字符串

第一章那个 CGO 程序还不够简单，我们现在来看看更简单的版本：

```go
// hello.go
package main

//#include <stdio.h>
import "C"

func main() {
	C.puts(C.CString("Hello, World\n"))
}
```

我们不仅仅通过 `import "C"` 语句启用 CGO 特性，同时包含 C 语言的 `<stdio.h>` 头文件。然后通过 CGO 包的 `C.CString` 函数将 Go 语言字符串转为 C 语言字符串，最后调用 CGO 包的 `C.puts` 函数向标准输出窗口打印转换后的 C 字符串。

相比 “Hello, World 的革命” 一节中的 CGO 程序最大的不同是：我们没有在程序退出前释放 `C.CString` 创建的 C 语言字符串；还有我们改用 `puts` 函数直接向标准输出打印，之前是采用 `fputs` 向标准输出打印。

没有释放使用 `C.CString` 创建的 C 语言字符串会导致内存泄漏。但是对于这个小程序来说，这样是没有问题的，因为程序退出后操作系统会自动回收程序的所有资源。

## 2.1.3 使用自己的 C 函数

前面我们使用了标准库中已有的函数。现在我们先自定义一个叫 `SayHello` 的 C 函数来实现打印，然后从 Go 语言环境中调用这个 `SayHello` 函数：

```go
// hello.go
package main

/*
#include <stdio.h>

static void SayHello(const char* s) {
	puts(s);
}
*/
import "C"

func main() {
	C.SayHello(C.CString("Hello, World\n"))
}
```

除了 `SayHello` 函数是我们自己实现的之外，其它的部分和前面的例子基本相似。

我们也可以将 `SayHello` 函数放到当前目录下的一个 C 语言源文件中（后缀名必须是 `.c`）。因为是编写在独立的 C 文件中，为了允许外部引用，所以需要去掉函数的 `static` 修饰符。

```c
// hello.c

#include <stdio.h>

void SayHello(const char* s) {
	puts(s);
}
```

然后在 CGO 部分先声明 `SayHello` 函数，其它部分不变：

```go
// hello.go
package main

//void SayHello(const char* s);
import "C"

func main() {
	C.SayHello(C.CString("Hello, World\n"))
}
```

注意，如果之前运行的命令是 `go run hello.go` 或 `go build hello.go` 的话，此处须使用 `go run "your/package"` 或 `go build "your/package"` 才可以。若本就在包路径下的话，也可以直接运行 `go run .` 或 `go build`。

既然 `SayHello` 函数已经放到独立的 C 文件中了，我们自然可以将对应的 C 文件编译打包为静态库或动态库文件供使用。如果是以静态库或动态库方式引用 `SayHello` 函数的话，需要将对应的 C 源文件移出当前目录（CGO 构建程序会自动构建当前目录下的 C 源文件，从而导致 C 函数名冲突）。关于静态库等细节将在稍后章节讲解。

## 2.1.4 C 代码的模块化

在编程过程中，抽象和模块化是将复杂问题简化的通用手段。当代码语句变多时，我们可以将相似的代码封装到一个个函数中；当程序中的函数变多时，我们将函数拆分到不同的文件或模块中。而模块化编程的核心是面向程序接口编程（这里的接口并不是 Go 语言的 interface，而是 API 的概念）。

在前面的例子中，我们可以抽象一个名为 hello 的模块，模块的全部接口函数都在 hello.h 头文件定义：

```c
// hello.h
void SayHello(const char* s);
```

其中只有一个 SayHello 函数的声明。但是作为 hello 模块的用户来说，就可以放心地使用 SayHello 函数，而无需关心函数的具体实现。而作为 SayHello 函数的实现者来说，函数的实现只要满足头文件中函数的声明的规范即可。下面是 SayHello 函数的 C 语言实现，对应 hello.c 文件：

```c
// hello.c

#include "hello.h"
#include <stdio.h>

void SayHello(const char* s) {
	puts(s);
}
```

在 hello.c 文件的开头，实现者通过 `#include "hello.h"` 语句包含 SayHello 函数的声明，这样可以保证函数的实现满足模块对外公开的接口。

接口文件 hello.h 是 hello 模块的实现者和使用者共同的约定，但是该约定并没有要求必须使用 C 语言来实现 SayHello 函数。我们也可以用 C++ 语言来重新实现这个 C 语言函数：

```c++
// hello.cpp

#include <iostream>

extern "C" {
	#include "hello.h"
}

void SayHello(const char* s) {
	std::cout << s;
}
```

在 C++ 版本的 SayHello 函数实现中，我们通过 C++ 特有的 `std::cout` 输出流输出字符串。不过为了保证 C++ 语言实现的 SayHello 函数满足 C 语言头文件 hello.h 定义的函数规范，我们需要通过 `extern "C"` 语句指示该函数的链接符号遵循 C 语言的规则。

在采用面向 C 语言 API 接口编程之后，我们彻底解放了模块实现者的语言枷锁：实现者可以用任何编程语言实现模块，只要最终满足公开的 API 约定即可。我们可以用 C 语言实现 SayHello 函数，也可以使用更复杂的 C++ 语言来实现 SayHello 函数，当然我们也可以用汇编语言甚至 Go 语言来重新实现 SayHello 函数。


## 2.1.5 用 Go 重新实现 C 函数

其实 CGO 不仅仅用于 Go 语言中调用 C 语言函数，还可以用于导出 Go 语言函数给 C 语言函数调用。在前面的例子中，我们已经抽象一个名为 hello 的模块，模块的全部接口函数都在 hello.h 头文件定义：

```c
// hello.h
void SayHello(/*const*/ char* s);
```

现在我们创建一个 hello.go 文件，用 Go 语言重新实现 C 语言接口的 SayHello 函数:

```go
// hello.go
package main

import "C"

import "fmt"

//export SayHello
func SayHello(s *C.char) {
	fmt.Print(C.GoString(s))
}
```

我们通过 CGO 的 `//export SayHello` 指令将 Go 语言实现的函数 `SayHello` 导出为 C 语言函数。为了适配 CGO 导出的 C 语言函数，我们禁止了在函数的声明语句中的 const 修饰符。需要注意的是，这里其实有两个版本的 `SayHello` 函数：一个 Go 语言环境的；另一个是 C 语言环境的。cgo 生成的 C 语言版本 SayHello 函数最终会通过桥接代码调用 Go 语言版本的 SayHello 函数。

通过面向 C 语言接口的编程技术，我们不仅仅解放了函数的实现者，同时也简化的函数的使用者。现在我们可以将 SayHello 当作一个标准库的函数使用（和 puts 函数的使用方式类似）：

```go
package main

//#include "hello.h"
import "C"

func main() {
	C.SayHello(C.CString("Hello, World\n"))
}
```

一切似乎都回到了开始的 CGO 代码，但是代码内涵更丰富了。

## 2.1.6 面向 C 接口的 Go 编程

在开始的例子中，我们的全部 CGO 代码都在一个 Go 文件中。然后，通过面向 C 接口编程的技术将 SayHello 分别拆分到不同的 C 文件，而 main 依然是 Go 文件。再然后，是用 Go 函数重新实现了 C 语言接口的 SayHello 函数。但是对于目前的例子来说只有一个函数，要拆分到三个不同的文件确实有些繁琐了。

正所谓合久必分、分久必合，我们现在尝试将例子中的几个文件重新合并到一个 Go 文件。下面是合并后的成果：

```go
package main

//void SayHello(char* s);
import "C"

import (
	"fmt"
)

func main() {
	C.SayHello(C.CString("Hello, World\n"))
}

//export SayHello
func SayHello(s *C.char) {
	fmt.Print(C.GoString(s))
}
```

现在版本的 CGO 代码中 C 语言代码的比例已经很少了，但是我们依然可以进一步以 Go 语言的思维来提炼我们的 CGO 代码。通过分析可以发现 `SayHello` 函数的参数如果可以直接使用 Go 字符串是最直接的。在 Go1.10 中 CGO 新增加了一个 `_GoString_` 预定义的 C 语言类型，用来表示 Go 语言字符串。下面是改进后的代码：

```go
// +build go1.10

package main

//void SayHello(_GoString_ s);
import "C"

import (
	"fmt"
)

func main() {
	C.SayHello("Hello, World\n")
}

//export SayHello
func SayHello(s string) {
	fmt.Print(s)
}
```

虽然看起来全部是 Go 语言代码，但是执行的时候是先从 Go 语言的 `main` 函数，到 CGO 自动生成的 C 语言版本 `SayHello` 桥接函数，最后又回到了 Go 语言环境的 `SayHello` 函数。这个代码包含了 CGO 编程的精华，读者需要深入理解。

*思考题: main 函数和 SayHello 函数是否在同一个 Goroutine 里执行？*
