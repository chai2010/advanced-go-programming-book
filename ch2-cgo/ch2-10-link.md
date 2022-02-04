# 2.10 编译和链接参数

编译和链接参数是每一个 C/C++ 程序员需要经常面对的问题。构建每一个 C/C++ 应用均需要经过编译和链接两个步骤，CGO 也是如此。
本节我们将简要讨论 CGO 中经常用到的编译和链接参数的用法。

## 2.10.1 编译参数：CFLAGS/CPPFLAGS/CXXFLAGS

编译参数主要是头文件的检索路径，预定义的宏等参数。理论上来说 C 和 C++ 是完全独立的两个编程语言，它们可以有着自己独立的编译参数。
但是因为 C++ 语言对 C 语言做了深度兼容，甚至可以将 C++ 理解为 C 语言的超集，因此 C 和 C++ 语言之间又会共享很多编译参数。
因此 CGO 提供了 CFLAGS/CPPFLAGS/CXXFLAGS 三种参数，其中 CFLAGS 对应 C 语言编译参数 (以 `.c` 后缀名)、
CPPFLAGS 对应 C/C++ 代码编译参数 (*.c,*.cc,*.cpp,*.cxx)、CXXFLAGS 对应纯 C++ 编译参数 (*.cc,*.cpp,*.cxx)。

## 2.10.2 链接参数：LDFLAGS

链接参数主要包含要链接库的检索目录和要链接库的名字。因为历史遗留问题，链接库不支持相对路径，我们必须为链接库指定绝对路径。
cgo 中的 ${SRCDIR} 为当前目录的绝对路径。经过编译后的 C 和 C++ 目标文件格式是一样的，因此 LDFLAGS 对应 C/C++ 共同的链接参数。

## 2.10.3 pkg-config

为不同 C/C++ 库提供编译和链接参数是一项非常繁琐的工作，因此 cgo 提供了对应 `pkg-config` 工具的支持。
我们可以通过 `#cgo pkg-config xxx` 命令来生成 xxx 库需要的编译和链接参数，其底层通过调用
`pkg-config xxx --cflags` 生成编译参数，通过 `pkg-config xxx --libs` 命令生成链接参数。
需要注意的是 `pkg-config` 工具生成的编译和链接参数是 C/C++ 公用的，无法做更细的区分。

`pkg-config` 工具虽然方便，但是有很多非标准的 C/C++ 库并没有实现对其支持。
这时候我们可以手工为 `pkg-config` 工具创建对应库的编译和链接参数实现支持。

比如有一个名为 xxx 的 C/C++ 库，我们可以手工创建 `/usr/local/lib/pkgconfig/xxx.pc` 文件：

```
Name: xxx
Cflags:-I/usr/local/include
Libs:-L/usr/local/lib –lxxx2
```

其中 Name 是库的名字，Cflags 和 Libs 行分别对应 xxx 使用库需要的编译和链接参数。如果 `pc` 文件在其它目录，
可以通过 PKG_CONFIG_PATH 环境变量指定 `pkg-config` 工具的检索目录。

而对应 cgo 来说，我们甚至可以通过 PKG_CONFIG 环境变量可指定自定义的 pkg-config 程序。
如果是自己实现 CGO 专用的 pkg-config 程序，只要处理 `--cflags` 和 `--libs` 两个参数即可。

下面的程序是 macos 系统下生成 Python3 的编译和链接参数：

```go
// py3-config.go
func main() {
	for _, s := range os.Args {
		if s == "--cflags" {
			out, _ := exec.Command("python3-config", "--cflags").CombinedOutput()
			out = bytes.Replace(out, []byte("-arch"), []byte{}, -1)
			out = bytes.Replace(out, []byte("i386"), []byte{}, -1)
			out = bytes.Replace(out, []byte("x86_64"), []byte{}, -1)
			fmt.Print(string(out))
			return
		}
		if s == "--libs" {
			out, _ := exec.Command("python3-config", "--ldflags").CombinedOutput()
			fmt.Print(string(out))
			return
		}
	}
}
```

然后通过以下命令构建并使用自定义的 `pkg-config` 工具：

```
$ go build -o py3-config py3-config.go
$ PKG_CONFIG=./py3-config go build -buildmode=c-shared -o gopkg.so main.go
```

具体的细节可以参考 Go 实现 Python 模块章节。

## 2.10.4 go get 链

在使用 `go get` 获取 Go 语言包的同时会获取包依赖的包。比如 A 包依赖 B 包，B 包依赖 C 包，C 包依赖 D 包：
`pkgA -> pkgB -> pkgC -> pkgD -> ...`。再 go get 获取 A 包之后会依次线获取 BCD 包。
如果在获取 B 包之后构建失败，那么将导致链条的断裂，从而导致 A 包的构建失败。

链条断裂的原因有很多，其中常见的原因有：

- 不支持某些系统, 编译失败
- 依赖 cgo, 用户没有安装 gcc
- 依赖 cgo, 但是依赖的库没有安装
- 依赖 pkg-config, windows 上没有安装
- 依赖 pkg-config, 没有找到对应的 bc 文件
- 依赖 自定义的 pkg-config, 需要额外的配置
- 依赖 swig, 用户没有安装 swig, 或版本不对

仔细分析可以发现，失败的原因中和 CGO 相关的问题占了绝大多数。这并不是偶然现象，
自动化构建 C/C++ 代码一直是一个世界难题，到目前位置也没有出现一个大家认可的统一的 C/C++ 管理工具。

因为用了 cgo，比如 gcc 等构建工具是必须安装的，同时尽量要做到对主流系统的支持。
如果依赖的 C/C++ 包比较小并且有源代码的前提下，可以优先选择从代码构建。

比如 `github.com/chai2010/webp` 包通过为每个 C/C++ 源文件在当前包建立关键文件实现零配置依赖：

```
// z_libwebp_src_dec_alpha.c
#include "./internal/libwebp/src/dec/alpha.c"
```

因此在编译 `z_libwebp_src_dec_alpha.c` 文件时，会编译 libweb 原生的代码。
其中的依赖是相对目录，对于不同的平台支持可以保持最大的一致性。

## 2.10.5 多个非 main 包中导出 C 函数

官方文档说明导出的 Go 函数要放 main 包，但是真实情况是其它包的 Go 导出函数也是有效的。
因为导出后的 Go 函数就可以当作 C 函数使用，所以必须有效。但是不同包导出的 Go 函数将在同一个全局的名字空间，因此需要小心避免重名的问题。
如果是从不同的包导出 Go 函数到 C 语言空间，那么 cgo 自动生成的 `_cgo_export.h` 文件将无法包含全部导出的函数声明，
我们必须通过手写头文件的方式声明导出的全部函数。
