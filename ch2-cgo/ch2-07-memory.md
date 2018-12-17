# 2.7 CGO内存模型

CGO是架接Go语言和C语言的桥梁，它使二者在二进制接口层面实现了互通，但是我们要注意因两种语言的内存模型的差异而可能引起的问题。如果在CGO处理的跨语言函数调用时涉及到了指针的传递，则可能会出现Go语言和C语言共享某一段内存的场景。我们知道C语言的内存在分配之后就是稳定的，但是Go语言因为函数栈的动态伸缩可能导致栈中内存地址的移动(这是Go和C内存模型的最大差异)。如果C语言持有的是移动之前的Go指针，那么以旧指针访问Go对象时会导致程序崩溃。

## 2.7.1 Go访问C内存

C语言空间的内存是稳定的，只要不是被人为提前释放，那么在Go语言空间可以放心大胆地使用。在Go语言访问C语言内存是最简单的情形，我们在之前的例子中已经见过多次。

因为Go语言实现的限制，我们无法在Go语言中创建大于2GB内存的切片（具体请参考makeslice实现代码）。不过借助cgo技术，我们可以在C语言环境创建大于2GB的内存，然后转为Go语言的切片使用：

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

func makeByteSlize(n int) []byte {
	p := C.makeslice(C.size_t(n))
	return ((*[1 << 31]byte)(p))[0:n:n]
}

func freeByteSlice(p []byte) {
	C.free(unsafe.Pointer(&p[0]))
}

func main() {
	s := makeByteSlize(1<<32+1)
	s[len[s]-1] = 1234
	print(s[len[s]-1])
	freeByteSlice(p)
}
```

例子中我们通过makeByteSlize来创建大于4G内存大小的切片，从而绕过了Go语言实现的限制（需要代码验证）。而freeByteSlice辅助函数则用于释放从C语言函数创建的切片。

因为C语言内存空间是稳定的，基于C语言内存构造的切片也是绝对稳定的，不会因为Go语言栈的变化而被移动。

## 2.7.2 C临时访问传入的Go内存

cgo之所以存在的一大因素是为了方便在Go语言中接纳吸收过去几十年来使用C/C++语言软件构建的大量的软件资源。C/C++很多库都是需要通过指针直接处理传入的内存数据的，因此cgo中也有很多需要将Go内存传入C语言函数的应用场景。

假设一个极端场景：我们将一块位于某goroutinue的栈上的Go语言内存传入了C语言函数后，在此C语言函数执行期间，此goroutinue的栈因为空间不足的原因发生了扩展，也就是导致了原来的Go语言内存被移动到了新的位置。但是此时此刻C语言函数并不知道该Go语言内存已经移动了位置，仍然用之前的地址来操作该内存——这将将导致内存越界。以上是一个推论（真实情况有些差异），也就是说C访问传入的Go内存可能是不安全的！

当然有RPC远程过程调用的经验的用户可能会考虑通过完全传值的方式处理：借助C语言内存稳定的特性，在C语言空间先开辟同样大小的内存，然后将Go的内存填充到C的内存空间；返回的内存也是如此处理。下面的例子是这种思路的具体实现：

```go
package main

/*
void printString(const char* s) {
	printf("%s", s);
}
*/
import "C"

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

在需要将Go的字符串传入C语言时，先通过`C.CString`将Go语言字符串对应的内存数据复制到新创建的C语言内存空间上。上面例子的处理思路虽然是安全的，但是效率极其低下（因为要多次分配内存并逐个复制元素），同时也极其繁琐。

为了简化并高效处理此种向C语言传入Go语言内存的问题，cgo针对该场景定义了专门的规则：在CGO调用的C语言函数返回前，cgo保证传入的Go语言内存在此期间不会发生移动，C语言函数可以大胆地使用Go语言的内存！

根据新的规则我们可以直接传入Go字符串的内存：

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

任何完美的技术都有被滥用的时候，CGO的这种看似完美的规则也是存在隐患的。我们假设调用的C语言函数需要长时间运行，那么将会导致被他引用的Go语言内存在C语言返回前不能被移动，从而可能间接地导致这个Go内存栈对应的goroutine不能动态伸缩栈内存，也就是可能导致这个goroutine被阻塞。因此，在需要长时间运行的C语言函数（特别是在纯CPU运算之外，还可能因为需要等待其它的资源而需要不确定时间才能完成的函数），需要谨慎处理传入的Go语言内存。

不过需要小心的是在取得Go内存后需要马上传入C语言函数，不能保存到临时变量后再间接传入C语言函数。因为CGO只能保证在C函数调用之后被传入的Go语言内存不会发生移动，它并不能保证在传入C函数之前内存不发生变化。

以下代码是错误的：

```go
// 错误的代码
tmp := uintptr(unsafe.Pointer(&x))
pb := (*int16)(unsafe.Pointer(tmp))
*pb = 42
```

因为tmp并不是指针类型，在它获取到Go对象地址之后x对象可能会被移动，但是因为不是指针类型，所以不会被Go语言运行时更新成新内存的地址。在非指针类型的tmp保持Go对象的地址，和在C语言环境保持Go对象的地址的效果是一样的：如果原始的Go对象内存发生了移动，Go语言运行时并不会同步更新它们。

## 2.7.3 C长期持有Go指针对象

作为一个Go程序员在使用CGO时潜意识会认为总是Go调用C函数。其实CGO中，C语言函数也可以回调Go语言实现的函数。特别是我们可以用Go语言写一个动态库，导出C语言规范的接口给其它用户调用。当C语言函数调用Go语言函数的时候，C语言函数就成了程序的调用方，Go语言函数返回的Go对象内存的生命周期也就自然超出了Go语言运行时的管理。简言之，我们不能在C语言函数中直接使用Go语言对象的内存。

虽然Go语言禁止在C语言函数中长期持有Go指针对象，但是这种需求是切实存在的。如果需要在C语言中访问Go语言内存对象，我们可以将Go语言内存对象在Go语言空间映射为一个int类型的id，然后通过此id来间接访问和控制Go语言对象。

以下代码用于将Go对象映射为整数类型的ObjectId，用完之后需要手工调用free方法释放该对象ID：

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

我们通过一个map来管理Go语言对象和id对象的映射关系。其中NewObjectId用于创建一个和对象绑定的id，而id对象的方法可用于解码出原始的Go对象，也可以用于结束id和原始Go对象的绑定。

下面一组函数以C接口规范导出，可以被C语言函数调用：

```go
package main

/*
extern char* NewGoString(char* );
extern void FreeGoString(char* );
extern void PrintGoString(char* );

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

在printString函数中，我们通过NewGoString创建一个对应的Go字符串对象，返回的其实是一个id，不能直接使用。我们借助PrintGoString函数将id解析为Go语言字符串后打印。该字符串在C语言函数中完全跨越了Go语言的内存管理，在PrintGoString调用前即使发生了栈伸缩导致的Go字符串地址发生变化也依然可以正常工作，因为该字符串对应的id是稳定的，在Go语言空间通过id解码得到的字符串也就是有效的。

## 2.7.4 导出C函数不能返回Go内存

在Go语言中，Go是从一个固定的虚拟地址空间分配内存。而C语言分配的内存则不能使用Go语言保留的虚拟内存空间。在CGO环境，Go语言运行时默认会检查导出返回的内存是否是由Go语言分配的，如果是则会抛出运行时异常。

下面是CGO运行时异常的例子：

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

其中getGoPtr返回的虽然是C语言类型的指针，但是内存本身是从Go语言的new函数分配，也就是由Go语言运行时统一管理的内存。然后我们在C语言的Main函数中调用了getGoPtr函数，此时默认将发送运行时异常：

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

异常说明cgo函数返回的结果中含有Go语言分配的指针。指针的检查操作发生在C语言版的getGoPtr函数中，它是由cgo生成的桥接C语言和Go语言的函数。

下面是cgo生成的C语言版本getGoPtr函数的具体细节（在cgo生成的`_cgo_export.c`文件定义）：

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

其中`_cgo_tsan_acquire`是从LLVM项目移植过来的内存指针扫描函数，它会检查cgo函数返回的结果是否包含Go指针。

需要说明的是，cgo默认对返回结果的指针的检查是有代价的，特别是cgo函数返回的结果是一个复杂的数据结构时将花费更多的时间。如果已经确保了cgo函数返回的结果是安全的话，可以通过设置环境变量`GODEBUG=cgocheck=0`来关闭指针检查行为。

```
$ GODEBUG=cgocheck=0 go run main.go
```

关闭cgocheck功能后再运行上面的代码就不会出现上面的异常的。但是要注意的是，如果C语言使用期间对应的内存被Go运行时释放了，将会导致更严重的崩溃问题。cgocheck默认的值是1，对应一个简化版本的检测，如果需要完整的检测功能可以将cgocheck设置为2。

关于cgo运行时指针检测的功能详细说明可以参考Go语言的官方文档。
