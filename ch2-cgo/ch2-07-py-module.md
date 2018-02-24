# 2.7 Go实现Python模块

前面章节我们已经讲述了如何通过CGO来引用和创建C动态库和静态库。实现了对C动态库和静态库的支持，理论上就可以应用到动态库的绝大部分场景。Python语言作为当下最红的语言，本节我们将演示如何通过Go语言来为Python脚本语言编写扩展模块。

## 基于ctypes

Python内置了非常丰富的模块，其中ctypes支持直接从C动态库调用函数。为了演示如何基于ctypes技术来扩展模块，我们需要先用Go语言创建一个C动态库。

我们使用的是之前出现过的例子：

```go
// main.go
package main

import "C"
import "fmt"

func main() {}

//export SayHello
func SayHello(name *C.char) {
	fmt.Printf("hello %s!\n", C.GoString(name))
}
```

其中只导出了一个SayHello函数，用于打印字符串。通过以下命令基于上述Go代码创建say-hello.so动态库：

```
go build -buildmode=c-shared -o say-hello.so main.go
```

现在我们就可以通过ctypes模块调用say-hello.so动态库中的SayHello函数了：

```python
// hello.py
import ctypes

libso = ctypes.CDLL("./say-hello.so")

SayHello = libso.SayHello
SayHello.argtypes = [ctypes.c_char_p]
SayHello.restype = None

SayHello(ctypes.c_char_p(b"hello"))
```

我们首先通过ctypes.CDLL加载动态库到libso，并通过libso.SayHello来获取SayHello函数。获取到SayHello函数之后设置函数的输入参数为一个C语言类型的字符串，该函数没有返回值。然后我们通过`ctypes.c_char_p(b"hello")`将Python字节串转为C语言格式的字符串作为参数调用SayHello。如果一切正常的话就可以输出字符串了。

从这个例子可以看出，给予ctypes构造Python扩展模块非常简单，本质上只是在构建一个纯C语言规格的动态库。比较复杂的部分在ctypes的具体使用，关于ctypes的具体细节就不详细展开的，用户可以自行参考Python自带的官方文档。

## 基于Python C接口创建

在前面的例子中，通过ctypes创建的模块必须要用Python再包装一层，否则就要直接面对C语言风格的接口。如果基于基于Python C接口，我们可以完全再Go和C语言层面创建灵活强大的模块，重点是不再需要在Python中重新包装。

基于Python C接口创建模块和使用C语言的静态库的流程类似：

```
package main

/*
// macOS:
#cgo darwin pkg-config: python3

// linux
#cgo linux pkg-config: python3

// windows
// should generate libpython3.a from python3.lib

#define Py_LIMITED_API
#include <Python.h>

extern PyObject* PyInit_gopkg();
extern PyObject* Py_gopkg_sum(PyObject *, PyObject *);

static int cgo_PyArg_ParseTuple_ii(PyObject *arg, int *a, int *b) {
	return PyArg_ParseTuple(arg, "ii", a, b);
}

static PyObject* cgo_PyInit_gopkg(void) {
	static PyMethodDef methods[] = {
		{"sum", Py_gopkg_sum, METH_VARARGS, "Add two numbers."},
		{NULL, NULL, 0, NULL},
	};
	static struct PyModuleDef module = {
		PyModuleDef_HEAD_INIT, "gopkg", NULL, -1, methods,
	};
	return PyModule_Create(&module);
}
*/
import "C"

func main() {}

//export PyInit_gopkg
func PyInit_gopkg() *C.PyObject {
	return C.cgo_PyInit_gopkg()
}

//export Py_gopkg_sum
func Py_gopkg_sum(self, args *C.PyObject) *C.PyObject {
	var a, b C.int
	if C.cgo_PyArg_ParseTuple_ii(args, &a, &b) == 0 {
		return nil
	}
	return C.PyLong_FromLong(C.long(a + b))
}
```

因为Python的链接参数要复杂了很多，我们借助pkg-config工具来获取编译参数和链接参数。然后我们在Go语言中分别导出了PyInit_gopkg和Py_gopkg_sum函数，其中PyInit_gopkg函数用于初始化名为gopkg的Python模块，而Py_gopkg_sum函数则是模块中sum方法的实现。

因此PyArg_ParseTuple是可变参数类型，CGO中无法使用可变参数的C函数，因此我们通过增加一个cgo_PyArg_ParseTuple_ii辅助函数小消除可变参数的影响。同样，模块的方法列表必须在C语言内存空间创建，因为CGO是禁止将Go语言内存直接返回到C语言空间的。

然后通过以下命令创建gopkg.so动态库：

```
go build -buildmode=c-shared -o gopkg.so main.go
```

这里需要注意几个出现gopkg名字的地方。gopkg是我们创建的Python模块的名字，因此它对应一个gopkg.so动态库。再gopkg.so动态库中必须有一个PyInit_gopkg函数，该函数是模块的初始化函数。在PyInit_gopkg函数初始化模块时，同样需要指定模块的名字时gopkg。模块中的方法函数是通过函数指针访问，具体的名字没有影响。

### macOS环境构建

因为在macOS中，pkg-config不支持Python3版本。不过macOS有一个python3-config的命令可以实现pkg-config类似的功能。不过python3-config生成的编译参数无法直接用于CGO编译选项（因为GCC不能识别部分参数会导致错误构建）。

我们在python3-config的基础只是又包装了一个工具，在通过python3-config获取到编译参数之后将GCC不支持的参数剔除掉。

创建py3-config.go文件：

```go
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

cgo中的pkg-config只需要两个参数`--cflags`和`--libs`。其中`--libs`选项的输出我们采用的是`python3-config --ldflags`的输出，因为`--libs`选项没有包含库的检索路径，而`--ldflags`选项则是在指定链接库参数的基础上增加了库的检索路径。

基于py3-config.go可以创建一个py3-config命令。然后通过PKG_CONFIG环境变量将cgo使用的pkg-config命令指定为我们订制的命令：

```
PKG_CONFIG=./py3-config go build -buildmode=c-shared -o gopkg.so main.go
```

对于不支持pkg-config的平台我们都可以基于类似的方法处理。
