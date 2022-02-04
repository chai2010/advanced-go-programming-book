# 2.7 CGO 内存模型

CGO 是架接 Go 语言和 C 语言的桥梁，它使二者在二进制接口层面实现了互通，但是我们要注意因两种语言的内存模型的差异而可能引起的问题。如果在 CGO 处理的跨语言函数调用时涉及到了指针的传递，则可能会出现 Go 语言和 C 语言共享某一段内存的场景。我们知道 C 语言的内存在分配之后就是稳定的，但是 Go 语言因为函数栈的动态伸缩可能导致栈中内存地址的移动 (这是 Go 和 C 内存模型的最大差异)。如果 C 语言持有的是移动之前的 Go 指针，那么以旧指针访问 Go 对象时会导致程序崩溃。

## 2.7.1 Go 访问 C 内存

C 语言空间的内存是稳定的，只要不是被人为提前释放，那么在 Go 语言空间可以放心大胆地使用。在 Go 语言访问 C 语言内存是最简单的情形，我们在之前的例子中已经见过多次。

因为 Go 语言实现的限制，我们无法在 Go 语言中创建大于 2GB 内存的切片（具体请参考 makeslice 实现代码）。不过借助 cgo 技术，我们可以在 C 语言环境创建大于 2GB 的内存，然后转为 Go 语言的切片使用：

```go
package main

/*
#include <stdlib.h>

void* makeslice(size_t memsize) {
	return malloc(memsize);
}
*/
import "C"
import "unsafe"

func makeByteSlice(n int) []byte {
	p := C.makeslice(C.size_t(n))
	return ((*[1 << 31]byte)(p))[0:n:n]
}

func freeByteSlice(p []byte) {
	C.free(unsafe.Pointer(&p[0]))
}

func main() {
	s := makeByteSlice(1<<32+1)
	s[len(s)-1] = 255
	print(s[len(s)-1])
	freeByteSlice(s)
}
```

例子中我们通过 makeByteSlice 来创建大于 4G 内存大小的切片，从而绕过了 Go 语言实现的限制（需要代码验证）。而 freeByteSlice 辅助函数则用于释放从 C 语言函数创建的切片。

因为 C 语言内存空间是稳定的，基于 C 语言内存构造的切片也是绝对稳定的，不会因为 Go 语言栈的变化而被移动。

## 2.7.2 C 临时访问传入的 Go 内存

cgo 之所以存在的一大因素是为了方便在 Go 语言中接纳吸收过去几十年来使用 C/C++ 语言软件构建的大量的软件资源。C/C++ 很多库都是需要通过指针直接处理传入的内存数据的，因此 cgo 中也有很多需要将 Go 内存传入 C 语言函数的应用场景。

假设一个极端场景：我们将一块位于某 goroutine 的栈上的 Go 语言内存传入了 C 语言函数后，在此 C 语言函数执行期间，此 goroutinue 的栈因为空间不足的原因发生了扩展，也就是导致了原来的 Go 语言内存被移动到了新的位置。但是此时此刻 C 语言函数并不知道该 Go 语言内存已经移动了位置，仍然用之前的地址来操作该内存——这将将导致内存越界。以上是一个推论（真实情况有些差异），也就是说 C 访问传入的 Go 内存可能是不安全的！

当然有 RPC 远程过程调用的经验的用户可能会考虑通过完全传值的方式处理：借助 C 语言内存稳定的特性，在 C 语言空间先开辟同样大小的内存，然后将 Go 的内存填充到 C 的内存空间；返回的内存也是如此处理。下面的例子是这种思路的具体实现：

```go
package main

/*
#include <stdlib.h>
#include <stdio.h>

void printString(const char* s) {
	printf("%s", s);
}
*/
import "C"
import "unsafe"

func printString(s string) {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))

	C.printString(cs)
}

func main() {
	s := "hello"
	printString(s)
}
```

在需要将 Go 的字符串传入 C 语言时，先通过 `C.CString` 将 Go 语言字符串对应的内存数据复制到新创建的 C 语言内存空间上。上面例子的处理思路虽然是安全的，但是效率极其低下（因为要多次分配内存并逐个复制元素），同时也极其繁琐。

为了简化并高效处理此种向 C 语言传入 Go 语言内存的问题，cgo 针对该场景定义了专门的规则：在 CGO 调用的 C 语言函数返回前，cgo 保证传入的 Go 语言内存在此期间不会发生移动，C 语言函数可以大胆地使用 Go 语言的内存！

根据新的规则我们可以直接传入 Go 字符串的内存：

```go
package main

/*
#include<stdio.h>

void printString(const char* s, int n) {
	int i;
	for(i = 0; i < n; i++) {
		putchar(s[i]);
	}
	putchar('\n');
}
*/
import "C"

func printString(s string) {
	p := (*reflect.StringHeader)(unsafe.Pointer(&s))
	C.printString((*C.char)(unsafe.Pointer(p.Data)), C.int(len(s)))
}

func main() {
	s := "hello"
	printString(s)
}
```

现在的处理方式更加直接，且避免了分配额外的内存。完美的解决方案！

任何完美的技术都有被滥用的时候，CGO 的这种看似完美的规则也是存在隐患的。我们假设调用的 C 语言函数需要长时间运行，那么将会导致被他引用的 Go 语言内存在 C 语言返回前不能被移动，从而可能间接地导致这个 Go 内存栈对应的 goroutine 不能动态伸缩栈内存，也就是可能导致这个 goroutine 被阻塞。因此，在需要长时间运行的 C 语言函数（特别是在纯 CPU 运算之外，还可能因为需要等待其它的资源而需要不确定时间才能完成的函数），需要谨慎处理传入的 Go 语言内存。

不过需要小心的是在取得 Go 内存后需要马上传入 C 语言函数，不能保存到临时变量后再间接传入 C 语言函数。因为 CGO 只能保证在 C 函数调用之后被传入的 Go 语言内存不会发生移动，它并不能保证在传入 C 函数之前内存不发生变化。

以下代码是错误的：

```go
// 错误的代码
tmp := uintptr(unsafe.Pointer(&x))
pb := (*int16)(unsafe.Pointer(tmp))
*pb = 42
```

因为 tmp 并不是指针类型，在它获取到 Go 对象地址之后 x 对象可能会被移动，但是因为不是指针类型，所以不会被 Go 语言运行时更新成新内存的地址。在非指针类型的 tmp 保持 Go 对象的地址，和在 C 语言环境保持 Go 对象的地址的效果是一样的：如果原始的 Go 对象内存发生了移动，Go 语言运行时并不会同步更新它们。

## 2.7.3 C 长期持有 Go 指针对象

作为一个 Go 程序员在使用 CGO 时潜意识会认为总是 Go 调用 C 函数。其实 CGO 中，C 语言函数也可以回调 Go 语言实现的函数。特别是我们可以用 Go 语言写一个动态库，导出 C 语言规范的接口给其它用户调用。当 C 语言函数调用 Go 语言函数的时候，C 语言函数就成了程序的调用方，Go 语言函数返回的 Go 对象内存的生命周期也就自然超出了 Go 语言运行时的管理。简言之，我们不能在 C 语言函数中直接使用 Go 语言对象的内存。

虽然 Go 语言禁止在 C 语言函数中长期持有 Go 指针对象，但是这种需求是切实存在的。如果需要在 C 语言中访问 Go 语言内存对象，我们可以将 Go 语言内存对象在 Go 语言空间映射为一个 int 类型的 id，然后通过此 id 来间接访问和控制 Go 语言对象。

以下代码用于将 Go 对象映射为整数类型的 ObjectId，用完之后需要手工调用 free 方法释放该对象 ID：

```go
package main

import "sync"

type ObjectId int32

var refs struct {
	sync.Mutex
	objs map[ObjectId]interface{}
	next ObjectId
}

func init() {
	refs.Lock()
	defer refs.Unlock()

	refs.objs = make(map[ObjectId]interface{})
	refs.next = 1000
}

func NewObjectId(obj interface{}) ObjectId {
	refs.Lock()
	defer refs.Unlock()

	id := refs.next
	refs.next++

	refs.objs[id] = obj
	return id
}

func (id ObjectId) IsNil() bool {
	return id == 0
}

func (id ObjectId) Get() interface{} {
	refs.Lock()
	defer refs.Unlock()

	return refs.objs[id]
}

func (id *ObjectId) Free() interface{} {
	refs.Lock()
	defer refs.Unlock()

	obj := refs.objs[*id]
	delete(refs.objs, *id)
	*id = 0

	return obj
}
```

我们通过一个 map 来管理 Go 语言对象和 id 对象的映射关系。其中 NewObjectId 用于创建一个和对象绑定的 id，而 id 对象的方法可用于解码出原始的 Go 对象，也可以用于结束 id 和原始 Go 对象的绑定。

下面一组函数以 C 接口规范导出，可以被 C 语言函数调用：

```go
package main

/*
extern char* NewGoString(char*);
extern void FreeGoString(char*);
extern void PrintGoString(char*);

static void printString(const char* s) {
	char* gs = NewGoString(s);
	PrintGoString(gs);
	FreeGoString(gs);
}
*/
import "C"

//export NewGoString
func NewGoString(s *C.char) *C.char {
	gs := C.GoString(s)
	id := NewObjectId(gs)
	return (*C.char)(unsafe.Pointer(uintptr(id)))
}

//export FreeGoString
func FreeGoString(p *C.char) {
	id := ObjectId(uintptr(unsafe.Pointer(p)))
	id.Free()
}

//export PrintGoString
func PrintGoString(s *C.char) {
	id := ObjectId(uintptr(unsafe.Pointer(p)))
	gs := id.Get().(string)
	print(gs)
}

func main() {
	C.printString("hello")
}
```

在 printString 函数中，我们通过 NewGoString 创建一个对应的 Go 字符串对象，返回的其实是一个 id，不能直接使用。我们借助 PrintGoString 函数将 id 解析为 Go 语言字符串后打印。该字符串在 C 语言函数中完全跨越了 Go 语言的内存管理，在 PrintGoString 调用前即使发生了栈伸缩导致的 Go 字符串地址发生变化也依然可以正常工作，因为该字符串对应的 id 是稳定的，在 Go 语言空间通过 id 解码得到的字符串也就是有效的。

## 2.7.4 导出 C 函数不能返回 Go 内存

在 Go 语言中，Go 是从一个固定的虚拟地址空间分配内存。而 C 语言分配的内存则不能使用 Go 语言保留的虚拟内存空间。在 CGO 环境，Go 语言运行时默认会检查导出返回的内存是否是由 Go 语言分配的，如果是则会抛出运行时异常。

下面是 CGO 运行时异常的例子：

```go
/*
extern int* getGoPtr();

static void Main() {
	int* p = getGoPtr();
	*p = 42;
}
*/
import "C"

func main() {
	C.Main()
}

//export getGoPtr
func getGoPtr() *C.int {
	return new(C.int)
}
```

其中 getGoPtr 返回的虽然是 C 语言类型的指针，但是内存本身是从 Go 语言的 new 函数分配，也就是由 Go 语言运行时统一管理的内存。然后我们在 C 语言的 Main 函数中调用了 getGoPtr 函数，此时默认将发送运行时异常：

```
$ go run main.go
panic: runtime error: cgo result has Go pointer

goroutine 1 [running]:
main._cgoexpwrap_cfb3840e3af2_getGoPtr.func1(0xc420051dc0)
  command-line-arguments/_obj/_cgo_gotypes.go:60 +0x3a
main._cgoexpwrap_cfb3840e3af2_getGoPtr(0xc420016078)
  command-line-arguments/_obj/_cgo_gotypes.go:62 +0x67
main._Cfunc_Main()
  command-line-arguments/_obj/_cgo_gotypes.go:43 +0x41
main.main()
  /Users/chai/go/src/github.com/chai2010 \
  /advanced-go-programming-book/examples/ch2-xx \
  /return-go-ptr/main.go:17 +0x20
exit status 2
```

异常说明 cgo 函数返回的结果中含有 Go 语言分配的指针。指针的检查操作发生在 C 语言版的 getGoPtr 函数中，它是由 cgo 生成的桥接 C 语言和 Go 语言的函数。

下面是 cgo 生成的 C 语言版本 getGoPtr 函数的具体细节（在 cgo 生成的 `_cgo_export.c` 文件定义）：

```c
int* getGoPtr()
{
	__SIZE_TYPE__ _cgo_ctxt = _cgo_wait_runtime_init_done();
	struct {
		int* r0;
	} __attribute__((__packed__)) a;
	_cgo_tsan_release();
	crosscall2(_cgoexp_95d42b8e6230_getGoPtr, &a, 8, _cgo_ctxt);
	_cgo_tsan_acquire();
	_cgo_release_context(_cgo_ctxt);
	return a.r0;
}
```

其中 `_cgo_tsan_acquire` 是从 LLVM 项目移植过来的内存指针扫描函数，它会检查 cgo 函数返回的结果是否包含 Go 指针。

需要说明的是，cgo 默认对返回结果的指针的检查是有代价的，特别是 cgo 函数返回的结果是一个复杂的数据结构时将花费更多的时间。如果已经确保了 cgo 函数返回的结果是安全的话，可以通过设置环境变量 `GODEBUG=cgocheck=0` 来关闭指针检查行为。

```
$ GODEBUG=cgocheck=0 go run main.go
```

关闭 cgocheck 功能后再运行上面的代码就不会出现上面的异常的。但是要注意的是，如果 C 语言使用期间对应的内存被 Go 运行时释放了，将会导致更严重的崩溃问题。cgocheck 默认的值是 1，对应一个简化版本的检测，如果需要完整的检测功能可以将 cgocheck 设置为 2。

关于 cgo 运行时指针检测的功能详细说明可以参考 Go 语言的官方文档。
