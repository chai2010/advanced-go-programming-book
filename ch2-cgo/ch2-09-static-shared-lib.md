# 2.9 静态库和动态库

CGO 在使用 C/C++ 资源的时候一般有三种形式：直接使用源码；链接静态库；链接动态库。直接使用源码就是在 `import "C"` 之前的注释部分包含 C 代码，或者在当前包中包含 C/C++ 源文件。链接静态库和动态库的方式比较类似，都是通过在 LDFLAGS 选项指定要链接的库方式链接。本节我们主要关注在 CGO 中如何使用静态库和动态库相关的问题。

## 2.9.1 使用 C 静态库

如果 CGO 中引入的 C/C++ 资源有代码而且代码规模也比较小，直接使用源码是最理想的方式，但很多时候我们并没有源代码，或者从 C/C++ 源代码开始构建的过程异常复杂，这种时候使用 C 静态库也是一个不错的选择。静态库因为是静态链接，最终的目标程序并不会产生额外的运行时依赖，也不会出现动态库特有的跨运行时资源管理的错误。不过静态库对链接阶段会有一定要求：静态库一般包含了全部的代码，里面会有大量的符号，如果不同静态库之间出现了符号冲突则会导致链接的失败。

我们先用纯 C 语言构造一个简单的静态库。我们要构造的静态库名叫 number，库中只有一个 number_add_mod 函数，用于表示数论中的模加法运算。number 库的文件都在 number 目录下。

`number/number.h` 头文件只有一个纯 C 语言风格的函数声明：

```c
int number_add_mod(int a, int b, int mod);
```

`number/number.c` 对应函数的实现：

```c
#include "number.h"

int number_add_mod(int a, int b, int mod) {
	return (a+b)%mod;
}
```

因为 CGO 使用的是 GCC 命令来编译和链接 C 和 Go 桥接的代码。因此静态库也必须是 GCC 兼容的格式。

通过以下命令可以生成一个叫 libnumber.a 的静态库：

```
$ cd ./number
$ gcc -c -o number.o number.c
$ ar rcs libnumber.a number.o
```

生成 libnumber.a 静态库之后，我们就可以在 CGO 中使用该资源了。

创建 main.go 文件如下：

```go
package main

//#cgo CFLAGS: -I./number
//#cgo LDFLAGS: -L${SRCDIR}/number -lnumber
//
//#include "number.h"
import "C"
import "fmt"

func main() {
	fmt.Println(C.number_add_mod(10, 5, 12))
}
```

其中有两个 #cgo 命令，分别是编译和链接参数。CFLAGS 通过 `-I./number` 将 number 库对应头文件所在的目录加入头文件检索路径。LDFLAGS 通过 `-L${SRCDIR}/number` 将编译后 number 静态库所在目录加为链接库检索路径，`-lnumber` 表示链接 libnumber.a 静态库。需要注意的是，在链接部分的检索路径不能使用相对路径（C/C++ 代码的链接程序所限制），我们必须通过 cgo 特有的 `${SRCDIR}` 变量将源文件对应的当前目录路径展开为绝对路径（因此在 windows 平台中绝对路径不能有空白符号）。

因为我们有 number 库的全部代码，所以我们可以用 go generate 工具来生成静态库，或者是通过 Makefile 来构建静态库。因此发布 CGO 源码包时，我们并不需要提前构建 C 静态库。

因为多了一个静态库的构建步骤，这种使用了自定义静态库并已经包含了静态库全部代码的 Go 包无法直接用 go get 安装。不过我们依然可以通过 go get 下载，然后用 go generate 触发静态库构建，最后才是 go install 来完成安装。

为了支持 go get 命令直接下载并安装，我们 C 语言的 `#include` 语法可以将 number 库的源文件链接到当前的包。

创建 `z_link_number_c.c` 文件如下：

```c
#include "./number/number.c"
```

然后在执行 go get 或 go build 之类命令的时候，CGO 就是自动构建 number 库对应的代码。这种技术是在不改变静态库源代码组织结构的前提下，将静态库转化为了源代码方式引用。这种 CGO 包是最完美的。

如果使用的是第三方的静态库，我们需要先下载安装静态库到合适的位置。然后在 #cgo 命令中通过 CFLAGS 和 LDFLAGS 来指定头文件和库的位置。对于不同的操作系统甚至同一种操作系统的不同版本来说，这些库的安装路径可能都是不同的，那么如何在代码中指定这些可能变化的参数呢？

在 Linux 环境，有一个 pkg-config 命令可以查询要使用某个静态库或动态库时的编译和链接参数。我们可以在 #cgo 命令中直接使用 pkg-config 命令来生成编译和链接参数。而且还可以通过 PKG_CONFIG 环境变量定制 pkg-config 命令。因为不同的操作系统对 pkg-config 命令的支持不尽相同，通过该方式很难兼容不同的操作系统下的构建参数。不过对于 Linux 等特定的系统，pkg-config 命令确实可以简化构建参数的管理。关于 pkg-config 的使用细节在此我们不深入展开，大家可以自行参考相关文档。

## 2.9.2 使用 C 动态库

动态库出现的初衷是对于相同的库，多个进程可以共享同一个，以节省内存和磁盘资源。但是在磁盘和内存已经白菜价的今天，这两个作用已经显得微不足道了，那么除此之外动态库还有哪些存在的价值呢？从库开发角度来说，动态库可以隔离不同动态库之间的关系，减少链接时出现符号冲突的风险。而且对于 windows 等平台，动态库是跨越 VC 和 GCC 不同编译器平台的唯一的可行方式。

对于 CGO 来说，使用动态库和静态库是一样的，因为动态库也必须要有一个小的静态导出库用于链接动态库（Linux 下可以直接链接 so 文件，但是在 Windows 下必须为 dll 创建一个 `.a` 文件用于链接）。我们还是以前面的 number 库为例来说明如何以动态库方式使用。

对于在 macOS 和 Linux 系统下的 gcc 环境，我们可以用以下命令创建 number 库的的动态库：

```
$ cd number
$ gcc -shared -o libnumber.so number.c
```

因为动态库和静态库的基础名称都是 libnumber，只是后缀名不同而已。因此 Go 语言部分的代码和静态库版本完全一样：

```go
package main

//#cgo CFLAGS: -I./number
//#cgo LDFLAGS: -L${SRCDIR}/number -lnumber
//
//#include "number.h"
import "C"
import "fmt"

func main() {
	fmt.Println(C.number_add_mod(10, 5, 12))
}
```

编译时 GCC 会自动找到 libnumber.a 或 libnumber.so 进行链接。

对于 windows 平台，我们还可以用 VC 工具来生成动态库（windows 下有一些复杂的 C++ 库只能用 VC 构建）。我们需要先为 number.dll 创建一个 def 文件，用于控制要导出到动态库的符号。

number.def 文件的内容如下：

```
LIBRARY number.dll

EXPORTS
number_add_mod
```

其中第一行的 LIBRARY 指明动态库的文件名，然后的 EXPORTS 语句之后是要导出的符号名列表。

现在我们可以用以下命令来创建动态库（需要进入 VC 对应的 x64 命令行环境）。

```
$ cl /c number.c
$ link /DLL /OUT:number.dll number.obj number.def
```

这时候会为 dll 同时生成一个 number.lib 的导出库。但是在 CGO 中我们无法使用 lib 格式的链接库。

要生成 `.a` 格式的导出库需要通过 mingw 工具箱中的 dlltool 命令完成：

```
$ dlltool -dllname number.dll --def number.def --output-lib libnumber.a
```

生成了 libnumber.a 文件之后，就可以通过 `-lnumber` 链接参数进行链接了。

需要注意的是，在运行时需要将动态库放到系统能够找到的位置。对于 windows 来说，可以将动态库和可执行程序放到同一个目录，或者将动态库所在的目录绝对路径添加到 PATH 环境变量中。对于 macOS 来说，需要设置 DYLD_LIBRARY_PATH 环境变量。而对于 Linux 系统来说，需要设置 LD_LIBRARY_PATH 环境变量。

## 2.9.3 导出 C 静态库

CGO 不仅可以使用 C 静态库，也可以将 Go 实现的函数导出为 C 静态库。我们现在用 Go 实现前面的 number 库的模加法函数。

创建 number.go，内容如下：

```go
package main

import "C"

func main() {}

//export number_add_mod
func number_add_mod(a, b, mod C.int) C.int {
	return (a + b) % mod
}
```

根据 CGO 文档的要求，我们需要在 main 包中导出 C 函数。对于 C 静态库构建方式来说，会忽略 main 包中的 main 函数，只是简单导出 C 函数。采用以下命令构建：

```
$ go build -buildmode=c-archive -o number.a
```

在生成 number.a 静态库的同时，cgo 还会生成一个 number.h 文件。

number.h 文件的内容如下（为了便于显示，内容做了精简）：

```c
#ifdef __cplusplus
extern "C" {
#endif

extern int number_add_mod(int p0, int p1, int p2);

#ifdef __cplusplus
}
#endif
```

其中 `extern "C"` 部分的语法是为了同时适配 C 和 C++ 两种语言。核心内容是声明了要导出的 number_add_mod 函数。

然后我们创建一个 `_test_main.c` 的 C 文件用于测试生成的 C 静态库（用下划线作为前缀名是让为了让 go build 构建 C 静态库时忽略这个文件）：

```c
#include "number.h"

#include <stdio.h>

int main() {
	int a = 10;
	int b = 5;
	int c = 12;

	int x = number_add_mod(a, b, c);
	printf("(%d+%d)%%%d = %d\n", a, b, c, x);

	return 0;
}
```

通过以下命令编译并运行：

```
$ gcc -o a.out _test_main.c number.a
$ ./a.out
```

使用 CGO 创建静态库的过程非常简单。

## 2.9.4 导出 C 动态库

CGO 导出动态库的过程和静态库类似，只是将构建模式改为 `c-shared`，输出文件名改为 `number.so` 而已：

```
$ go build -buildmode=c-shared -o number.so
```

`_test_main.c` 文件内容不变，然后用以下命令编译并运行：

```
$ gcc -o a.out _test_main.c number.so
$ ./a.out
```

## 2.9.5 导出非 main 包的函数

通过 `go help buildmode` 命令可以查看 C 静态库和 C 动态库的构建说明：

```
-buildmode=c-archive
	Build the listed main package, plus all packages it imports,
	into a C archive file. The only callable symbols will be those
	functions exported using a cgo //export comment. Requires
	exactly one main package to be listed.

-buildmode=c-shared
	Build the listed main package, plus all packages it imports,
	into a C shared library. The only callable symbols will
	be those functions exported using a cgo //export comment.
	Requires exactly one main package to be listed.
```

文档说明导出的 C 函数必须是在 main 包导出，然后才能在生成的头文件包含声明的语句。但是很多时候我们可能更希望将不同类型的导出函数组织到不同的 Go 包中，然后统一导出为一个静态库或动态库。

要实现从是从非 main 包导出 C 函数，或者是多个包导出 C 函数（因为只能有一个 main 包），我们需要自己提供导出 C 函数对应的头文件（因为 CGO 无法为非 main 包的导出函数生成头文件）。

假设我们先创建一个 number 子包，用于提供模加法函数：

```go
package number

import "C"

//export number_add_mod
func number_add_mod(a, b, mod C.int) C.int {
	return (a + b) % mod
}
```

然后是当前的 main 包：

```go
package main

import "C"

import (
	"fmt"

	_ "./number"
)

func main() {
	println("Done")
}

//export goPrintln
func goPrintln(s *C.char) {
	fmt.Println("goPrintln:", C.GoString(s))
}
```

其中我们导入了 number 子包，在 number 子包中有导出的 C 函数 number_add_mod，同时我们在 main 包也导出了 goPrintln 函数。

通过以下命令创建 C 静态库：

```
$ go build -buildmode=c-archive -o main.a
```

这时候在生成 main.a 静态库的同时，也会生成一个 main.h 头文件。但是 main.h 头文件中只有 main 包中导出的 goPrintln 函数的声明，并没有 number 子包导出函数的声明。其实 number_add_mod 函数在生成的 C 静态库中是存在的，我们可以直接使用。

创建 `_test_main.c` 测试文件如下：

```c
#include <stdio.h>

void goPrintln(char*);
int number_add_mod(int a, int b, int mod);

int main() {
	int a = 10;
	int b = 5;
	int c = 12;

	int x = number_add_mod(a, b, c);
	printf("(%d+%d)%%%d = %d\n", a, b, c, x);

	goPrintln("done");
	return 0;
}
```

我们并没有包含 CGO 自动生成的 main.h 头文件，而是通过手工方式声明了 goPrintln 和 number_add_mod 两个导出函数。这样我们就实现了从多个 Go 包导出 C 函数了。

