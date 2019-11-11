# 2.4 函数调用

函数是C语言编程的核心，通过CGO技术我们不仅仅可以在Go语言中调用C语言函数，也可以将Go语言函数导出为C语言函数。

## 2.4.1 Go调用C函数

对于一个启用CGO特性的程序，CGO会构造一个虚拟的C包。通过这个虚拟的C包可以调用C语言函数。

```go
/*
static int add(int a, int b) {
	return a+b;
}
*/
import "C"

func main() {
	C.add(1, 1)
}
```

以上的CGO代码首先定义了一个当前文件内可见的add函数，然后通过`C.add`。

## 2.4.2 C函数的返回值

对于有返回值的C函数，我们可以正常获取返回值。

```go
/*
static int div(int a, int b) {
	return a/b;
}
*/
import "C"
import "fmt"

func main() {
	v := C.div(6, 3)
	fmt.Println(v)
}
```

上面的div函数实现了一个整数除法的运算，然后通过返回值返回除法的结果。

不过对于除数为0的情形并没有做特殊处理。如果希望在除数为0的时候返回一个错误，其他时候返回正常的结果。因为C语言不支持返回多个结果，因此`<errno.h>`标准库提供了一个`errno`宏用于返回错误状态。我们可以近似地将`errno`看成一个线程安全的全局变量，可以用于记录最近一次错误的状态码。

改进后的div函数实现如下：

```c
#include <errno.h>

int div(int a, int b) {
	if(b == 0) {
		errno = EINVAL;
		return 0;
	}
	return a/b;
}
```

CGO也针对`<errno.h>`标准库的`errno`宏做的特殊支持：在CGO调用C函数时如果有两个返回值，那么第二个返回值将对应`errno`错误状态。

```go
/*
#include <errno.h>

static int div(int a, int b) {
	if(b == 0) {
		errno = EINVAL;
		return 0;
	}
	return a/b;
}
*/
import "C"
import "fmt"

func main() {
	v0, err0 := C.div(2, 1)
	fmt.Println(v0, err0)

	v1, err1 := C.div(1, 0)
	fmt.Println(v1, err1)
}
```

运行这个代码将会产生以下输出：

```
2 <nil>
0 invalid argument
```

我们可以近似地将div函数看作为以下类型的函数：

```go
func C.div(a, b C.int) (C.int, [error])
```

第二个返回值是可忽略的error接口类型，底层对应 `syscall.Errno` 错误类型。

## 2.4.3 void函数的返回值

C语言函数还有一种没有返回值类型的函数，用void表示返回值类型。一般情况下，我们无法获取void类型函数的返回值，因为没有返回值可以获取。前面的例子中提到，cgo对errno做了特殊处理，可以通过第二个返回值来获取C语言的错误状态。对于void类型函数，这个特性依然有效。

以下的代码是获取没有返回值函数的错误状态码：

```go
//static void noreturn() {}
import "C"
import "fmt"

func main() {
	_, err := C.noreturn()
	fmt.Println(err)
}
```

此时，我们忽略了第一个返回值，只获取第二个返回值对应的错误码。

我们也可以尝试获取第一个返回值，它对应的是C语言的void对应的Go语言类型：

```go
//static void noreturn() {}
import "C"
import "fmt"

func main() {
	v, _ := C.noreturn()
	fmt.Printf("%#v", v)
}
```

运行这个代码将会产生以下输出：

```
main._Ctype_void{}
```

我们可以看出C语言的void类型对应的是当前的main包中的`_Ctype_void`类型。其实也将C语言的noreturn函数看作是返回`_Ctype_void`类型的函数，这样就可以直接获取void类型函数的返回值：

```go
//static void noreturn() {}
import "C"
import "fmt"

func main() {
	fmt.Println(C.noreturn())
}
```

运行这个代码将会产生以下输出：

```
[]
```

其实在CGO生成的代码中，`_Ctype_void`类型对应一个0长的数组类型`[0]byte`，因此`fmt.Println`输出的是一个表示空数值的方括弧。

以上有效特性虽然看似有些无聊，但是通过这些例子我们可以精确掌握CGO代码的边界，可以从更深层次的设计的角度来思考产生这些奇怪特性的原因。


## 2.4.4 C调用Go导出函数

CGO还有一个强大的特性：将Go函数导出为C语言函数。这样的话我们可以定义好C语言接口，然后通过Go语言实现。在本章的第一节快速入门部分我们已经展示过Go语言导出C语言函数的例子。

下面是用Go语言重新实现本节开始的add函数：

```go
import "C"

//export add
func add(a, b C.int) C.int {
	return a+b
}
```

add函数名以小写字母开头，对于Go语言来说是包内的私有函数。但是从C语言角度来看，导出的add函数是一个可全局访问的C语言函数。如果在两个不同的Go语言包内，都存在一个同名的要导出为C语言函数的add函数，那么在最终的链接阶段将会出现符号重名的问题。

CGO生成的 `_cgo_export.h` 文件会包含导出后的C语言函数的声明。我们可以在纯C源文件中包含 `_cgo_export.h` 文件来引用导出的add函数。如果希望在当前的CGO文件中马上使用导出的C语言add函数，则无法引用 `_cgo_export.h` 文件。因为`_cgo_export.h` 文件的生成需要依赖当前文件可以正常构建，而如果当前文件内部循环依赖还未生成的`_cgo_export.h` 文件将会导致cgo命令错误。

```c
#include "_cgo_export.h"

void foo() {
	add(1, 1);
}
```

当导出C语言接口时，需要保证函数的参数和返回值类型都是C语言友好的类型，同时返回值不得直接或间接包含Go语言内存空间的指针。

