# 2.1. 你好, CGO!

在第一章的“Hello, World 的革命”一节中，我们已经见过一个CGO程序。这一节我们将重新给出三个版本的CGO实现，来简单展示CGO的用法。

## 基于C标准库

第一章那个CGO程序还不够简单，我们现在来看看更简单的版本：

```go
package main

//#include <stdio.h>
import "C"

func main() {
	C.puts(C.CString("Hello, World\n"))
}
```

我们通过`import "C"`语句启用CGO特性，同时包含C语言的`<stdio.h>`头文件。然后通过CGO包的`C.CString`函数将Go语言字符串转为C语言字符串，最后调用C语言的`C.puts`函数向标准输出窗口打印转换后的C字符串。

相比“Hello, World 的革命”一节中的CGO程序最大的不同是：我们没有在程序退出前释放`C.CString`创建的C语言字符串；还有我们改用`puts`函数直接向标准输出打印，之前是采用`fputs`向标准输出打印。

没有释放使用`C.CString`创建的C语言字符串会导致内存泄露。但是对于这个小程序来说，这样是没有问题的，因为程序退出后操作系统会自动回收程序的所有资源。

## 使用自己的C函数

前面我们使用了标准库中已有的函数。现在我们先自定义一个叫`SayHello`的C函数来实现打印，然后从Go语言环境中调用这个`SayHello`函数：

```go
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

除了`SayHello`函数是我们自己实现的之外，其它的部分和前面的例子基本相似。

我们也可以将`SayHello`函数放到当前目录下的一个C语言源文件中（后缀名必须是`.c`）。因为是编写在独立的C文件中，为了允许外部引用，所以需要去掉函数的`static`修饰符。

```c
// hello.c

#include <stdio.h>

void SayHello(const char* s) {
	puts(s);
}
```

然后在CGO部分先声明`SayHello`函数，其它部分不变：

```go
package main

//void SayHello(const char* s);
import "C"

func main() {
	C.SayHello(C.CString("Hello, World\n"))
}
```

既然`SayHello`函数已经放到独立的C文件中了，我们自然可以将对应的C文件编译打包为静态库或动态库文件供使用。如果是以静态库或动态库方式引用`SayHello`函数的话，需要将对应的C源文件移出当前目录（CGO构建程序会自动构建当前目录下的C源文件，从而导致C函数名冲突）。关于静态库等细节将在稍后章节讲解。

## 用Go来实现C函数

其实CGO不仅仅用于Go语言中调用C语言函数，还可以用于导出Go语言函数给C语言函数调用。

```go
package main

/*
#include <stdio.h>

void cgoPuts(char* s);

static void SayHello(const char* s) {
	cgoPuts((char*)(s));
}
*/
import "C"

func main() {
	C.SayHello(C.CString("Hello, World\n"))
}

//export cgoPuts
func cgoPuts(s *C.char) {
	fmt.Print(C.GoString(s))
}
```

我们通过CGO的`//export cgoPuts`指令将Go语言实现的函数`cgoPuts`导出给C语言函数使用。然后在C语言版本的`SayHello`函数中，用`cgoPuts`替换之前的`puts`函数调用。在使用之前，同样要先声明`cgoPuts`函数。

需要主要的是，这里其实有两个版本的`cgoPuts`函数：一个Go语言环境的；另一个是C语言环境的。在C语言环境中，`SayHello`调用的也是C语言环境的`cgoPuts`函数；这是CGO自动生成的桥接函数，内部会调用Go语言环境的`cgoPuts`函数。因此，我们也可以直接在Go语言环境中调用C语言环境的`cgoPuts`函数。

现在我们可以改用Go语言重新实现C语言接口的`SayHello`函数，然后在`main`函数中还是和之前一样调用`C.SayHello`实现输出：

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

虽然看起来全部是Go语言代码，但是执行的时候是先从Go语言的`main`函数，到CGO自动生成的C语言版本`SayHello`桥接函数，最后又回到了Go语言环境的`SayHello`函数。虽然看起来有点绕，但CGO确实是这样运行的。

需要注意的是，CGO导出Go语言函数时，函数参数中不再支持C语言中`const`修饰符。
