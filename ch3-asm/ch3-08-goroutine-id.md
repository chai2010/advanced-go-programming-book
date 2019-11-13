# 3.8 Example: Goroutine ID

In the operating system, each process has a unique process number, and each thread also has its own unique thread number. Also in the Go language, each Goroutine also has its own unique Go-number, which is often encountered in panic and other scenarios. Although Goroutine has an intrinsic number, Go does not deliberately provide an interface to get the number. In this section we try to get the Goroutine ID through the Go assembly language.

## 3.8.1 Intentional design without goid

According to official information, the reason why Go does not provide a goal is to avoid being abused. Because most users can easily get the goid code, in the later programming will unconsciously write code that depends on the godid. Strong dependency on the goid will result in poor porting of these code and will also complicate the concurrency model. At the same time, there may be a large amount of Goroutine in the Go language, but when each Goroutine is destroyed, it is not well monitored in real time, which also causes the resources that depend on the godid to not be automatically reclaimed automatically (requires manual recycling). However, if you are a Go assembly language user, you can ignore these excuses.

## 3.8.2 Pure Go way to get goid

In order to facilitate understanding, we first try to get the goid in pure Go way. Although the performance of using the pure Go method to obtain the goid is low, the code has good portability, and can also be used to test and verify whether the goid obtained by other methods is correct.

Every Go user should know the panic function. Calling the panic function will cause a Goroutine exception. If the root function passed to Goroutine has not been processed by the recover function, the runtime will print the relevant exception and stack information and exit Goroutine.

Let's construct a simple example to output the goid via panic:

```go
Package main

Func main() {
Panic("goid")
}
```

After running, the following information will be output:

```
Panic: goid

Goroutine 1 [running]:
Main.main()
/path/to/main.go:4 +0x40
```

We can guess that the 1 in the Panic output information `goroutine 1 [running]` is the godd. But how can I get the output of panic in the program? In fact, the above information is only a textual description of the current function call stack frame, and the runtime.Stack function provides the function of obtaining the information.

We reconstruct an example based on the runtime.Stack function and output the goid by outputting the information of the current stack frame:

```go
Package main

Import "runtime"

Func main() {
Var buf = make([]byte, 64)
Var stk = buf[:runtime.Stack(buf, false)]
Print(string(stk))
}
```

After running, the following information will be output:

```
Goroutine 1 [running]:
Main.main()
/path/to/main.g
```

Therefore, the name information can be easily parsed from the string obtained from runtime.Stack:

```go
Func GetGoid() int64 {
Var (
Buf [64]byte
n = runtime.Stack(buf[:], false)
Stk = strings.TrimPrefix(string(buf[:n]), "goroutine ")
)

idField := strings.Fields(stk)[0]
Id, err := strconv.Atoi(idField)
If err != nil {
Panic(fmt.Errorf("can not get goroutine id: %v", err))
}

Return int64(id)
}
```

The details of the GetGoid function are not repeated here. It should be added that the `runtime.Stack` function can not only get the stack information of the current Goroutine, but also get the stack information of all Goroutine (controlled by the second parameter). At the same time, the [net/http2.curGoroutineID](https://github.com/golang/net/blob/master/http2/gotrack.go) function inside the Go language is the goid obtained in a similar way.


## 3.8.3 Getting the god from the g structure

According to the official Go assembly language documentation, the g pointer for each running Goroutine structure is stored in the local storage TLS of the system thread currently running Goroutine. You can get the TLS thread local storage first, then get the pointer of the g structure from TLS, and finally take the goid from the g structure.

The following is the g_pointer obtained by referring to the get_tls macro defined in the runtime package:

```
Get_tls(CX)
MOVQ g(CX), AX // Move g into AX.
```

Where get_tls is a macro function defined in the [runtime/go_tls.h] (https://github.com/golang/go/blob/master/src/runtime/go_tls.h) header file.

For the AMD64 platform, the get_tls macro function is defined as follows:

```
#ifdef GOARCH_amd64
#define get_tls(r) MOVQ TLS, r
#define g(r) 0(r)(TLS*1)
#endif
```

After expanding the get_tls macro function, the code to get the g pointer is as follows:

```
MOVQ TLS, CX
MOVQ 0 (CX) (TLS*1), AX
```

In fact, TLS is similar to the address stored locally by the thread, and the data in the memory corresponding to the address is the g pointer. We can also be more straightforward:

```
MOVQ (TLS), AX
```

Based on the above method, you can wrap a getg function to get the g pointer:

```
// func getg() unsafe.Pointer
TEXT · getg(SB), NOSPLIT, $0-8
MOVQ (TLS), AX
MOVQ AX, ret+0(FP)
RET
```

Then get the value of goid in the Go code by the offset of the goid member in the g structure:

```go
Const g_goid_offset = 152 // Go1.10

Func GetGroutineId() int64 {
g := getg()
p := (*int64)(unsafe.Pointer(uintptr(g) + g_goid_offset))
Return *p
}
```

Where `g_goid_offset` is the offset of the goid member, g structure reference [runtime/runtime2.go] (https://github.com/golang/go/blob/master/src/runtime/runtime2.go).

In Go1.10, the offset of the godd is 152 bytes. Therefore the above code can only be run correctly in the Go version with a target offset of 152 bytes. According to the god of Thompson, the enumeration and violent exhaustion are the essential oils to solve all intractable diseases. We can also save the offset of the goid to the table and then query the offset of the goid according to the Go version number.

Here's the improved code:

```go
Var offsetDictMap = map[string]int64{
"go1.10": 152,
"go1.9": 152,
"go1.8": 192,
}

Var g_goid_offset = func() int64 {
Goversion := runtime.Version()
For key, off := range offsetDictMap {
If goversion == key || strings.HasPrefix(goversion, key) {
Return off
}
}
Panic("unsupport go verion:"+goversion)
}()
```

The current god offset has finally been automatically adapted to the already released version of the Go language.


## 3.8.4 Obtaining the interface object corresponding to the g structure

Enumeration and violent exhaustion are straightforward, but support for the unreleased version of Go being developed is not good, and we can't know in advance the offset of a member of a certain version of the development.

If it is inside the runtime package, we can get the offset of the member directly through `unsafe.OffsetOf(g.goid)`. You can also get the type of g structure by reflection, and then query the offset of a member by type. Because the g structure is an internal type, the Go code cannot get the type information of the g structure from the outer package. But in the Go assembly language, we can see all the symbols, so in theory we can also get the type information of the g structure.

After any type is defined, the Go language will generate corresponding type information for that type. For example, the g structure will generate a `type·runtime·g` identifier to represent the value type information of the g structure, and a `type·*runtime·g` identifier to indicate the pointer type information. If the g structure has methods, it also generates `go.itab.runtime.g` and `go.itab.*runtime.g` type information to indicate the type information of the method.

If we can get the `type·runtime·g` and g pointers representing the g structure type, then we can construct the interface of the g object. The following is an improved getg function that returns the interface of the g pointer object:

```
// func getg() interface{}
TEXT ·getg(SB), NOSPLIT, $32-16
// get runtime.g
MOVQ (TLS), AX
// get runtime.g type
MOVQ $type·runtime·g(SB), BX

// convert (*g) to interface{}
MOVQ AX, 8(SP)
MOVQ BX, 0(SP)
CALL runtime·convT2E(SB)
MOVQ 16 (SP), AX
MOVQ 24 (SP), BX

// return interface{}
MOVQ AX, ret+0(FP)
MOVQ BX, ret+8(FP)
RET
```

The AX register corresponds to the g pointer, and the BX register corresponds to the type of the g structure. Then convert the type to an interface via the runtime·convT2E function. Because we are not using the g structure pointer type, the returned interface represents the g structure value type. In theory we can also construct an interface of the g pointer type, but due to the limitations of the Go assembly language, we cannot use the `type·*runtime·g` identifier.

Based on the interface returned by g, you can easily get the godd:

```go
Func GetGoid() int64 {
g := getg()
Gid := reflect.ValueOf(g).FieldByName("goid").Int()
Return goid
}
```

The above code directly obtains the goid through reflection. In theory, as long as the reflected interface and the name of the goid member do not change, the code can run normally. After actual testing, the above code works correctly in Go1.8, Go1.9 and Go1.10 versions. Optimistic speculation, if the name of the g structure type does not change, the mechanism of Go language reflection does not change, then the future Go language version should also be able to run.

Although the reflection has a certain flexibility, the performance of reflection has always been a place of criticism. An improved idea is to get the offset of the godd by reflection, and then get the goid through the g pointer and the offset, so that the reflection only needs to be executed once during the initialization phase.

Here is the initialization code for the g_goid_offset variable:

```go
Var g_goid_offset uintptr = func() uintptr {
g := GetGroutine()
If f, ok := reflect.TypeOf(g).FieldByName("goid"); ok {
Returnf.Offset
}
Panic("can not find g.goid field")
}()
```

With the correct god offset, get the goid in the same way as before:


```go
Func GetGroutineId() int64 {
g := getg()
p := (*int64)(unsafe.Pointer(uintptr(g) + g_goid_offset))
Return *p
}
```

At this point, we have achieved a complete idea of ​​getting the goid, but the compiled code still has serious security risks.

Although the getg function is a function type that prohibits stack splitting with the NOSPLIT flag, getg internally calls the more complex runtime·convT2E function. The runtime·convT2E function may trigger a stack split operation if it encounters insufficient stack space. When the stack splits, the GC will move the parameters and return values ​​of all functions on the stack and the stack pointer in the local variables. But our getg function does not provide pointer information for local variables.

The following is a complete implementation of the improved getg function:

```
// func getg() interface{}
TEXT ·getg(SB), NOSPLIT, $32-16
NO_LOCAL_POINTERS

MOVQ $0, ret_type+0(FP)
MOVQ $0, ret_data+8(FP)
GO_RESULTS_INITIALIZED

// get runtime.g
MOVQ (TLS), AX

// get runtime.g type
MOVQ $type·runtime·g(SB), BX

// convert (*g) to interface{}
MOVQ AX, 8(SP)
MOVQ BX, 0(SP)
CALL runtime·convT2E(SB)
MOVQ 16 (SP), AX
MOVQ 24 (SP), BX

// return interface{}
MOVQ AX, ret_type+0(FP)
MOVQ BX, ret_data+8(FP)
RET
```

Where NO_LOCAL_POINTERS indicates that the function does not have a local pointer variable. At the same time, the returned interface is initialized with zero value, and the GC is notified by GO_RESULTS_INITIALIZED after the initialization is completed. This allows the GC to correctly handle pointers in return values ​​and local variables when the stack is guaranteed to split.


## 3.8.5 Application of goid: Local storage

With the god, it's very easy to construct Goroutine local storage. We can define a gls package to provide the characteristics of the god:

```go
Package gls

Var gls struct {
m map[int64]map[interface{}]interface{}
sync.Mutex
}

Func init() {
Gls.m = make(map[int64]map[interface{}]interface{})
}
```

The gls package variable simply wraps the map and supports concurrent access via the `sync.Mutex` mutex.

Then define a getMap internal function that gets the map for each Goroutine byte:

```go
Func getMap() map[interface{}]interface{} {
gls.Lock()
Defer gls.Unlock()

Goid := GetGoid()
If m, _ := gls.m[goid]; m != nil {
Return m
}

m := make(map[interface{}]interface{})
Gls.m[goid] = m
Return m
}
```

After getting the private map of Goroutine, it is the normal interface for adding, deleting, and changing operations:

```go
Func Get(key interface{}) interface{} {
Return getMap()[key]
}
Func Put(key interface{}, v interface{}) {
getMap()[key] = v
}
Func Delete(key interface{}) {
Delete(getMap(), key)
}
```

Finally, we provide a Clean function to release the map resource corresponding to Goroutine:

```go
Func Clean() {
gls.Lock()
Defer gls.Unlock()

Delete(gls.m, GetGoid())
}
```

Such a minimalist Goroutine local storage gls object is complete.

Here's a simple example of using local storage:

```go
Import (
Gls "path/to/gls"
)

Func main() {
Var wg sync.WaitGroup
For i := 0; i < 5; i++ {
wg.Add(1)
Go func(idx int) {
Defer wg.Done()
Defer gls.Clean()

Defer func() {
fmt.Printf("%d: number = %d\n", idx, gls.Get("number"))
}()
gls.Put("number", idx+100)
}(i)
}
wg.Wait()
}
```

With Goroutine local storage, storage resources can be shared between different levels of functions. At the same time, in order to avoid resource leakage, you need to call the gls.Clean() function to release resources in the root function of Goroutine.