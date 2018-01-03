# 2.4. CGO内存模型(Doing)

CGO是架接Go语言和C语言的桥梁，它不仅仅在二进制接口层面实现互通，同时要考虑两种语言的内存模型的差异。如果在CGO处理的跨语言函数调用时涉及指针的传递，则可能会出现Go语言和C语言共享某一段内存的场景。我们知道C语言的内存在分配之后就是稳定的，但是Go语言因为函数栈的动态伸缩可能导致栈中内存地址的移动。如果C语言持有的是移动之前的Go指针，那么以旧指针访问Go对象时会导致程序崩溃。这是Go和C内存模型的最大差异。

## Go访问C内存

在Go语言访问C语言内存是最简单的情形，我们在之前的例子中已经见过多次。因此C语言空间的内存是稳定的，只要不是被人为提前释放，那么在Go语言空间可以放心大胆地使用。

因为Go语言实现的现在，我们无法在Go语言中创建大于2GB内存的切片（具体请参考makeslice实现代码）。不过借助cgo技术，我们可以在C语言环境创建大于2GB的内存，然后转为Go语言的切片使用：

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

例子中我们通过makeByteSlize来创建大于4G内存大小的切片，从而绕过了Go语言实现的限制（需要代码验证）。而freeByteSlice辅助函数用于释放从C语言函数创建的切片。

因为C语言内存空间是稳定的，基于C语言内存构造的切片也是绝对稳定的，不会因为Go语言栈的变化而被移动。


## C临时访问传入的Go内存

## C长期持有Go指针对象

TODO
