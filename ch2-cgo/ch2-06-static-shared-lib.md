# 2.6. 静态库和动态库

CGO在使用C/C++资源的时候一般有三种形式：直接使用源码；链接静态库；链接动态库。直接使用源码就是在`import "C"`之前的注释部分包含C代码，或者在当前包中包含C/C++源文件。链接静态库和动态库的方式比较类似，都是通过在LDFLAGS选项指定要链接的库方式链接。本节我们主要关注CGO中如何处理静态库和动态库相关的问题。

## 使用C静态库

如果CGO中引入的C/C++资源有代码而且代码规模也比较小，直接使用源码是最理想的方式。很多时候我们并没有源代码，或者从C/C++源代码的构建过程异常复杂，这种时候使用C静态库也是一个不错的选择。静态库因为是静态链接，最终的目标程序并不会产生额外的运行时依赖，而且也不会出现动态库特有的跨运行时资源管理的错误。不过静态库对链接阶段会有一定要求：静态库一般包含了全部的代码，里面会有大量的符号，如果不同静态库之间出现符号冲突会导致链接的失败。

我们先用纯C语言构造一个简单的静态库。我们要构造的静态库名叫number，库中只有一个number_add_mod函数，用于表示数论中的模加法运算。number库的文件都在number目录下。

`number/number.h`头文件只有一个纯C语言风格的函数声明：

```c
int number_add_mod(int a, int b, int mod);
```

`number/number.c`对应函数的实现：

```c
#include "number.h"

int number_add_mod(int a, int b, int mod) {
	return (a+b)%mod;
}
```

因为CGO使用的是GCC命令来编译和链接C和Go桥接的代码。因此静态库也必须是GCC兼容的格式。

通过以下命令可以生成一个libnumber.a的静态库：

```
$ cd ./number
$ gcc -c -o number.o number.c
$ ar rcs libnumber.a number.o
```

生成libnumber.a静态库之后，我们就可以在CGO中使用该资源了。

创建main.go文件如下：

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

其中有两个#cgo命令，分别是编译和链接参数。CFLAGS通过`-I./number`将number库对应头文件所在的目录加入头文件检索路径。LDFLAGS通过`-L${SRCDIR}/number`将编译后number静态库所在目录加为链接库检索路径，`-lnumber`表示链接libnumber.a静态库。需要注意的是，在链接部分的检索路径不能使用相对路径（这是由于C/C++代码的链接程序所限制），我们必须通过cgo特有的`${SRCDIR}`变量将源文件对应的当前目录路径展开为绝对路径（因此在windows平台中绝对路径不能有空白符号）。

因为我们有number库的全部代码，我们可以用go generate工具来生成静态库，或者是通过Makefile来构建静态库。因此发布CGO源码包时，我们并不需要提前构建C静态库。

因为多了一个静态库的构建步骤，这种使用了自定义静态库并已经包含了静态库全部代码的Go包无法直接用go get安装。不过我们依然可以通过go get下载，然后用go generate触发静态库构建，最后才是go install安装来完成。

为了支持go get命令直接下载并安装，我们C语言的`#include`语法可以将number库的源文件链接到当前的包。

创建`z_link_number_c.c`文件如下：

```c
#include "./number/number.c"
```

然后在执行go get或go build之类命令的时候，CGO就是自动构建number库对应的代码。这种技术是在不改变静态库源代码组织结构的前提下，将静态库转化为了源代码方式引用。这种CGO包是最完美的。

如果使用的是第三方的静态库，我们需要先下载安装静态库到合适的位置。然后在#cgo命令中通过CFLAGS和LDFLAGS来指定头文件和库的位置。对于不同的操作系统甚至同一种操作系统的不同版本来说，这些库的安装路径可能都是不同的，那么如何在代码中指定这些可能变化的参数呢？

在Linux环境，有个一个pkg-config命令可以查询要使用某个静态库或动态库时的编译和链接参数。我们可以在#cgo命令中直接使用pkg-config命令来生成编译和链接参数。而且还可以通过PKG_CONFIG环境变量订制pkg-config命令。因为不同的操作系统对pkg-config命令的支持不尽相同，通过该方式很难兼容不同的操作系统下的构建参数。不过对于Linux等特定的系统，pkg-config命令确实可以简化构建参数的管理。关于pkg-config的使用细节我们不再深入展开，大家可以自行参考相关文档。

## 使用C动态库

好处是跨编译器
坏处是多了一个依赖


## 导出C静态库

## 导出C动态库

## 导出非main包的函数

## Plugin

静态注入
C动态库注入
Goplugin特性

<!--
使用c静态库
使用c动态库
创建c静态库
创建c动态库

跨越多个go包导出函数

动态库的注意点
-->
