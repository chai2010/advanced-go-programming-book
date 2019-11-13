# 2.7 CGO Memory Model

CGO is a bridge between Go and C. It enables interoperability at the binary interface level, but we should pay attention to the problems that may arise due to the difference in memory models between the two languages. If the transfer of pointers is involved in the cross-language function call handled by CGO, there may be scenes in which Go language and C language share a certain segment of memory. We know that the C language's memory is stable after allocation, but the Go language may cause the memory address in the stack to move due to the dynamic scaling of the function stack (this is the biggest difference between the Go and C memory models). If the C language holds the Go pointer before the move, accessing the Go object with the old pointer will cause the program to crash.

## 2.7.1 Go access C memory

The memory of the C language space is stable, as long as it is not released by humans in advance, then the Go language space can be used with confidence. Accessing C memory in Go is the simplest case, and we've seen it many times in previous examples.

Because of the limitations of Go implementation, we can't create slices larger than 2GB of memory in Go (see the makeice implementation code for details). But with the help of cgo technology, we can create more than 2GB of memory in the C language environment, and then switch to the Go language slice:

```go
Package main

/*
#include <stdlib.h>

Void* makeslice(size_t memsize) {
Return malloc(memsize);
}
*/
Import "C"
Import "unsafe"

Func makeByteSlize(n int) []byte {
p := C.makeslice(C.size_t(n))
Return ((*[1 << 31]byte)(p))[0:n:n]
}

Func freeByteSlice(p []byte) {
C.free(unsafe.Pointer(&p[0]))
}

Func main() {
s := makeByteSlize(1<<32+1)
s[len(s)-1] = 255
Print(s[len(s)-1])
freeByteSlice(s)
}
```

In the example, we use makeByteSlize to create a slice larger than 4G memory size, thus bypassing the Go language implementation limitation (requires code validation). The freeByteSlice helper function is used to release the slice created from the C language function.

Because the C language memory space is stable, slices based on C language memory constructs are also absolutely stable and will not be moved due to changes in the Go language stack.

## 2.7.2 C Temporary access to incoming Go memory

A major factor in the existence of cgo is to facilitate the adoption of a large amount of software resources built in the Go language using C/C++ language software over the past few decades. Many libraries in C/C++ need to directly process incoming memory data through pointers. Therefore, there are many application scenarios in Cgo that need to pass Go memory into C language functions.

Suppose an extreme scenario: after we pass a Go language function on a stack of a goroutinue, we pass the C language function. During the execution of this C language function, the stack of this goroutinue is expanded due to insufficient space, which leads to The original Go language memory was moved to a new location. But at this point in the C language function does not know that the Go language memory has moved the location, still use the previous address to operate the memory - this will lead to memory out of bounds. The above is a corollary (there are some differences in the real situation), which means that C access to the incoming Go memory may be unsafe!

Of course, users with experience with RPC remote procedure calls may consider processing by completely passing values: with the C language memory stable feature, first open the same amount of memory in the C language space, and then fill Go memory into C memory. Space; the memory returned is treated as such. The following example is a concrete implementation of this idea:

```go
Package main

/*
Void printString(const char* s) {
Printf("%s", s);
}
*/
Import "C"

Func printString(s string) {
Cs := C.CString(s)
Defer C.free(unsafe.Pointer(cs))

C.printString(cs)
}

Func main() {
s := "hello"
printString(s)
}
```

When you need to pass Go's string into C language, first copy the memory data corresponding to the Go language string to the newly created C language memory space by `C.CString`. Although the above example is safe, it is extremely inefficient (because it has to allocate memory multiple times and copy elements one by one), and it is extremely cumbersome.

In order to simplify and efficiently handle this problem of passing Go language memory into C language, cgo defines a special rule for this scenario: before the C language function called by CGO returns, cgo guarantees that the incoming Go language memory does not exist during this period. Moves occur, C language functions can boldly use Go language memory!

According to the new rules we can directly pass in the memory of the Go string:

```go
Package main

/*
#include<stdio.h>

Void printString(const char* s, int n) {
Int i;
For(i = 0; i < n; i++) {
Putchar(s[i]);
}
Putchar('\n');
}
*/
Import "C"

Func printString(s string) {
p := (*reflect.StringHeader)(unsafe.Pointer(&s))
C.printString((*C.char)(unsafe.Pointer(p.Data)), C.int(len(s)))
}

Func main() {
s := "hello"
printString(s)
}
```

The current processing is more straightforward and avoids allocating extra memory. The perfect solution!

When any perfect technology is abused, CGO's seemingly perfect rules are also hidden. We assume that the called C language function needs to run for a long time, which will cause the Go language referenced by him to be unable to be moved before the C language returns, which may indirectly cause the goroutine corresponding to the Go memory stack to not dynamically scale the stack memory. That is, it may cause this goroutine to be blocked. Therefore, in C language functions that need to run for a long time (especially in pure CPU operations, but also because of the need to wait for other resources and need uncertain time to complete the function), you need to be careful to handle the incoming Go language memory.

However, you need to be careful to pass the C language function immediately after getting the Go memory. You cannot save the temporary variable and then pass the C language function indirectly. Because CGO can only guarantee that the Go language memory that is passed in after the C function call does not move, it does not guarantee that the memory will not change before the C function is passed.

The following code is wrong:

```go
// wrong code
Tmp := uintptr(unsafe.Pointer(&x))
Pb := (*int16)(unsafe.Pointer(tmp))
*pb = 42
```

Because tmp is not a pointer type, the x object may be moved after it gets the Go object address, but because it is not a pointer type, it will not be updated to the address of the new memory when the Go language is run. Keeping the Go object's address in a non-pointer type of tmp has the same effect as keeping the Go object's address in the C locale: if the original Go object's memory has moved, the Go language runtime will not update them synchronously.

## 2.7.3 C long-term holding Go pointer object

As a Go programmer, when using CGO, the subconscious will always think that Go calls C functions. In fact, in CGO, C language functions can also call back functions implemented by Go. In particular, we can write a dynamic library in Go language, and export the interface of the C language specification to other users. When the C language function calls the Go language function, the C language function becomes the caller of the program, and the life cycle of the Go object memory returned by the Go language function is naturally beyond the management of the Go language runtime. In short, we can't use the memory of the Go language object directly in the C language function.

Although the Go language prohibits long-term holding of Go pointer objects in C language functions, this requirement is tangible. If you need to access Go language memory objects in C language, we can map Go language memory objects in Go language space to an int type id, and then indirectly access and control Go language objects through this id.

The following code is used to map the Go object to an ObjectId of the integer type. After using it, you need to manually call the free method to release the object ID:

```go
Package main

Import "sync"

Type ObjectId int32

Var refs struct {
sync.Mutex
Objs map[ObjectId]interface{}
Next ObjectId
}

Func init() {
refs.Lock()
Defer refs.Unlock()

Refs.objs = make(map[ObjectId]interface{})
Refs.next = 1000
}

Func NewObjectId(obj interface{}) ObjectId {
refs.Lock()
Defer refs.Unlock()

Id := refs.next
Refs.next++

Refs.objs[id] = obj
Return id
}

Func (id ObjectId) IsNil() bool {
Return id == 0
}

Func (id ObjectId) Get() interface{} {
refs.Lock()
Defer refs.Unlock()

Return refs.objs[id]
}

Func (id *ObjectId) Free() interface{} {
refs.Lock()
Defer refs.Unlock()

Obj := refs.objs[*id]
Delete(refs.objs, *id)
*id = 0

Return obj
}
```

We use a map to manage the mapping between Go language objects and id objects. The NewObjectId is used to create an id bound to the object, and the id object's method can be used to decode the original Go object, and can also be used to end the binding of the id and the original Go object.

The following set of functions are exported as C interface specifications and can be called by C language functions:

```go
Package main

/*
Extern char* NewGoString(char* );
Extern void FreeGoString(char* );
Extern void PrintGoString(char* );

Static void printString(const char* s) {
Char* gs = NewGoString(s);
PrintGoString(gs);
FreeGoString(gs);
}
*/
Import "C"

//export NewGoString
Func NewGoString(s *C.char) *C.char {
Gs := C.GoString(s)
Id := NewObjectId(gs)
Return (*C.char)(unsafe.Pointer(uintptr(id)))
}

//export FreeGoString
Func FreeGoString(p *C.char) {
Id := ObjectId(uintptr(unsafe.Pointer(p)))
id.Free()
}

//export PrintGoString
Func PrintGoString(s *C.char) {
Id := ObjectId(uintptr(unsafe.Pointer(p)))
Gs := id.Get().(string)
Print(gs)
}

Func main() {
C.printString("hello")
}
```

In the printString function, we create a corresponding Go string object through NewGoString, the return is actually an id, can not be used directly. We use the PrintGoString function to parse the id into a Go language string. This string completely spans the Go language in the C language function.Memory management, even if the Go string address changes caused by the stack scaling before the PrintGoString call, it can still work normally, because the id corresponding to the string is stable, and the string obtained by id decoding in the Go language space is Effective.

## 2.7.4 Exporting C functions cannot return Go memory

In Go, Go allocates memory from a fixed virtual address space. The memory allocated by the C language cannot use the virtual memory space reserved by the Go language. In the CGO environment, the Go language runtime checks by default whether the memory returned by the export is allocated by the Go language, and if so, a runtime exception is thrown.

The following is an example of a CGO runtime exception:

```go
/*
Extern int* getGoPtr();

Static void Main() {
Int* p = getGoPtr();
*p = 42;
}
*/
Import "C"

Func main() {
C.Main()
}

//export getGoPtr
Func getGoPtr() *C.int {
Return new(C.int)
}
```

The getGoPtr returns a pointer of the C language type, but the memory itself is allocated from the Go function of the Go language, which is the memory managed by the Go language runtime. Then we call the getGoPtr function in the C main function, and the runtime exception will be sent by default:

```
$ go run main.go
Panic: runtime error: cgo result has Go pointer

Goroutine 1 [running]:
main._cgoexpwrap_cfb3840e3af2_getGoPtr.func1(0xc420051dc0)
  Command-line-arguments/_obj/_cgo_gotypes.go:60 +0x3a
main._cgoexpwrap_cfb3840e3af2_getGoPtr(0xc420016078)
  Command-line-arguments/_obj/_cgo_gotypes.go:62 +0x67
main._Cfunc_Main()
  Command-line-arguments/_obj/_cgo_gotypes.go:43 +0x41
Main.main()
  /Users/chai/go/src/github.com/chai2010 \
  /advanced-go-programming-book/examples/ch2-xx \
  /return-go-ptr/main.go:17 +0x20
Exit status 2
```

The exception indicates that the result returned by the cgo function contains a pointer to the Go language assignment. The check operation of the pointer occurs in the C language version of the getGoPtr function, which is a function of the C language and Go language generated by cgo.

The following is the specific details of the C language version of the getGoPtr function generated by cgo (defined in the `_cgo_export.c` file generated by cgo):

```c
Int* getGoPtr()
{
__SIZE_TYPE__ _cgo_ctxt = _cgo_wait_runtime_init_done();
Struct {
Int* r0;
} __attribute__((__packed__)) a;
_cgo_tsan_release();
Crosscall2(_cgoexp_95d42b8e6230_getGoPtr, &a, 8, _cgo_ctxt);
_cgo_tsan_acquire();
_cgo_release_context(_cgo_ctxt);
Return a.r0;
}
```

Where `_cgo_tsan_acquire` is a memory pointer scan function ported from the LLVM project, which checks if the result returned by the cgo function contains a Go pointer.

It should be noted that cgo's default check of the pointer to the returned result is costly, especially if the result returned by the cgo function is a complex data structure, it will take more time. If you have ensured that the result returned by the cgo function is safe, you can turn off the pointer checking behavior by setting the environment variable `GODEBUG=cgocheck=0`.

```
$ GODEBUG=cgocheck=0 go run main.go
```

After closing the cgocheck function and then running the above code, the above exception will not occur. However, it should be noted that if the corresponding memory during the C language is released by the Go runtime, it will cause a more serious crash. The default value of cgocheck is 1, which corresponds to the detection of a simplified version. If you need the full detection function, you can set cgocheck to 2.

For detailed descriptions of the functions of the cgo runtime pointer detection, refer to the official documentation of the Go language.