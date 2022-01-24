# 3.8 例子：Goroutine ID

在操作系统中，每个进程都会有一个唯一的进程编号，每个线程也有自己唯一的线程编号。同样在 Go 语言中，每个 Goroutine 也有自己唯一的 Go 程编号，这个编号在 panic 等场景下经常遇到。虽然 Goroutine 有内在的编号，但是 Go 语言却刻意没有提供获取该编号的接口。本节我们尝试通过 Go 汇编语言获取 Goroutine ID。

## 3.8.1 故意设计没有 goid

根据官方的相关资料显示，Go 语言刻意没有提供 goid 的原因是为了避免被滥用。因为大部分用户在轻松拿到 goid 之后，在之后的编程中会不自觉地编写出强依赖 goid 的代码。强依赖 goid 将导致这些代码不好移植，同时也会导致并发模型复杂化。同时，Go 语言中可能同时存在海量的 Goroutine，但是每个 Goroutine 何时被销毁并不好实时监控，这也会导致依赖 goid 的资源无法很好地自动回收（需要手工回收）。不过如果你是 Go 汇编语言用户，则完全可以忽略这些借口。

## 3.8.2 纯 Go 方式获取 goid

为了便于理解，我们先尝试用纯 Go 的方式获取 goid。使用纯 Go 的方式获取 goid 的方式虽然性能较低，但是代码有着很好的移植性，同时也可以用于测试验证其它方式获取的 goid 是否正确。

每个 Go 语言用户应该都知道 panic 函数。调用 panic 函数将导致 Goroutine 异常，如果 panic 在传递到 Goroutine 的根函数还没有被 recover 函数处理掉，那么运行时将打印相关的异常和栈信息并退出 Goroutine。

下面我们构造一个简单的例子，通过 panic 来输出 goid：

```go
package main

func main() {
	panic("goid")
}
```

运行后将输出以下信息：

```
panic: goid

goroutine 1 [running]:
main.main()
	/path/to/main.go:4 +0x40
```

我们可以猜测 Panic 输出信息 `goroutine 1 [running]` 中的 1 就是 goid。但是如何才能在程序中获取 panic 的输出信息呢？其实上述信息只是当前函数调用栈帧的文字化描述，runtime.Stack 函数提供了获取该信息的功能。

我们基于 runtime.Stack 函数重新构造一个例子，通过输出当前栈帧的信息来输出 goid：

```go
package main

import "runtime"

func main() {
	var buf = make([]byte, 64)
	var stk = buf[:runtime.Stack(buf, false)]
	print(string(stk))
}
```

运行后将输出以下信息：

```
goroutine 1 [running]:
main.main()
	/path/to/main.g
```

因此从 runtime.Stack 获取的字符串中就可以很容易解析出 goid 信息：

```go
func GetGoid() int64 {
	var (
		buf [64]byte
		n   = runtime.Stack(buf[:], false)
		stk = strings.TrimPrefix(string(buf[:n]), "goroutine")
	)

	idField := strings.Fields(stk)[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Errorf("can not get goroutine id: %v", err))
	}

	return int64(id)
}
```

GetGoid 函数的细节我们不再赘述。需要补充说明的是 `runtime.Stack` 函数不仅仅可以获取当前 Goroutine 的栈信息，还可以获取全部 Goroutine 的栈信息（通过第二个参数控制）。同时在 Go 语言内部的 [net/http2.curGoroutineID](https://github.com/golang/net/blob/master/http2/gotrack.go) 函数正是采用类似方式获取的 goid。


## 3.8.3 从 g 结构体获取 goid

根据官方的 Go 汇编语言文档，每个运行的 Goroutine 结构的 g 指针保存在当前运行 Goroutine 的系统线程的局部存储 TLS 中。可以先获取 TLS 线程局部存储，然后再从 TLS 中获取 g 结构的指针，最后从 g 结构中取出 goid。

下面是参考 runtime 包中定义的 get_tls 宏获取 g 指针：

```
get_tls(CX)
MOVQ g(CX), AX     // Move g into AX.
```

其中 get_tls 是一个宏函数，在 [runtime/go_tls.h](https://github.com/golang/go/blob/master/src/runtime/go_tls.h) 头文件中定义。

对于 AMD64 平台，get_tls 宏函数定义如下：

```
#ifdef GOARCH_amd64
#define	get_tls(r)	MOVQ TLS, r
#define	g(r)	0(r)(TLS*1)
#endif
```

将 get_tls 宏函数展开之后，获取 g 指针的代码如下：

```
MOVQ TLS, CX
MOVQ 0(CX)(TLS*1), AX
```

其实 TLS 类似线程局部存储的地址，地址对应的内存里的数据才是 g 指针。我们还可以更直接一点:

```
MOVQ (TLS), AX
```

基于上述方法可以包装一个 getg 函数，用于获取 g 指针：

```
// func getg() unsafe.Pointer
TEXT ·getg(SB), NOSPLIT, $0-8
	MOVQ (TLS), AX
	MOVQ AX, ret+0(FP)
	RET
```

然后在 Go 代码中通过 goid 成员在 g 结构体中的偏移量来获取 goid 的值：

```go
const g_goid_offset = 152 // Go1.10

func GetGroutineId() int64 {
	g := getg()
	p := (*int64)(unsafe.Pointer(uintptr(g) + g_goid_offset))
	return *p
}
```

其中 `g_goid_offset` 是 goid 成员的偏移量，g 结构参考 [runtime/runtime2.go](https://github.com/golang/go/blob/master/src/runtime/runtime2.go)。

在 Go1.10 版本，goid 的偏移量是 152 字节。因此上述代码只能正确运行在 goid 偏移量也是 152 字节的 Go 版本中。根据汤普森大神的神谕，枚举和暴力穷举是解决一切疑难杂症的万金油。我们也可以将 goid 的偏移保存到表格中，然后根据 Go 版本号查询 goid 的偏移量。

下面是改进后的代码：

```go
var offsetDictMap = map[string]int64{
	"go1.10": 152,
	"go1.9":  152,
	"go1.8":  192,
}

var g_goid_offset = func() int64 {
	goversion := runtime.Version()
	for key, off := range offsetDictMap {
		if goversion == key || strings.HasPrefix(goversion, key) {
			return off
		}
	}
	panic("unsupported go version:"+goversion)
}()
```

现在的 goid 偏移量已经终于可以自动适配已经发布的 Go 语言版本。


## 3.8.4 获取 g 结构体对应的接口对象

枚举和暴力穷举虽然够直接，但是对于正在开发中的未发布的 Go 版本支持并不好，我们无法提前知晓开发中的某个版本的 goid 成员的偏移量。

如果是在 runtime 包内部，我们可以通过 `unsafe.OffsetOf(g.goid)` 直接获取成员的偏移量。也可以通过反射获取 g 结构体的类型，然后通过类型查询某个成员的偏移量。因为 g 结构体是一个内部类型，Go 代码无法从外部包获取 g 结构体的类型信息。但是在 Go 汇编语言中，我们是可以看到全部的符号的，因此理论上我们也可以获取 g 结构体的类型信息。

在任意的类型被定义之后，Go 语言都会为该类型生成对应的类型信息。比如 g 结构体会生成一个 `type·runtime·g` 标识符表示 g 结构体的值类型信息，同时还有一个 `type·*runtime·g` 标识符表示指针类型的信息。如果 g 结构体带有方法，那么同时还会生成 `go.itab.runtime.g` 和 `go.itab.*runtime.g` 类型信息，用于表示带方法的类型信息。

如果我们能够拿到表示 g 结构体类型的 `type·runtime·g` 和 g 指针，那么就可以构造 g 对象的接口。下面是改进的 getg 函数，返回 g 指针对象的接口：

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
	MOVQ 16(SP), AX
	MOVQ 24(SP), BX

	// return interface{}
	MOVQ AX, ret+0(FP)
	MOVQ BX, ret+8(FP)
	RET
```

其中 AX 寄存器对应 g 指针，BX 寄存器对应 g 结构体的类型。然后通过 runtime·convT2E 函数将类型转为接口。因为我们使用的不是 g 结构体指针类型，因此返回的接口表示的 g 结构体值类型。理论上我们也可以构造 g 指针类型的接口，但是因为 Go 汇编语言的限制，我们无法使用 `type·*runtime·g` 标识符。

基于 g 返回的接口，就可以容易获取 goid 了：

```go
func GetGoid() int64 {
	g := getg()
	gid := reflect.ValueOf(g).FieldByName("goid").Int()
	return goid
}
```

上述代码通过反射直接获取 goid，理论上只要反射的接口和 goid 成员的名字不发生变化，代码都可以正常运行。经过实际测试，以上的代码可以在 Go1.8、Go1.9 和 Go1.10 版本中正确运行。乐观推测，如果 g 结构体类型的名字不发生变化，Go 语言反射的机制也不发生变化，那么未来 Go 语言版本应该也是可以运行的。

反射虽然具备一定的灵活性，但是反射的性能一直是被大家诟病的地方。一个改进的思路是通过反射获取 goid 的偏移量，然后通过 g 指针和偏移量获取 goid，这样反射只需要在初始化阶段执行一次。

下面是 g_goid_offset 变量的初始化代码：

```go
var g_goid_offset uintptr = func() uintptr {
	g := GetGroutine()
	if f, ok := reflect.TypeOf(g).FieldByName("goid"); ok {
		return f.Offset
	}
	panic("can not find g.goid field")
}()
```

有了正确的 goid 偏移量之后，采用前面讲过的方式获取 goid：


```go
func GetGroutineId() int64 {
	g := getg()
	p := (*int64)(unsafe.Pointer(uintptr(g) + g_goid_offset))
	return *p
}
```

至此我们获取 goid 的实现思路已经足够完善了，不过汇编的代码依然有严重的安全隐患。

虽然 getg 函数是用 NOSPLIT 标志声明的禁止栈分裂的函数类型，但是 getg 内部又调用了更为复杂的 runtime·convT2E 函数。runtime·convT2E 函数如果遇到栈空间不足，可能触发栈分裂的操作。而栈分裂时，GC 将要挪动栈上所有函数的参数和返回值和局部变量中的栈指针。但是我们的 getg 函数并没有提供局部变量的指针信息。

下面是改进后的 getg 函数的完整实现：

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
	MOVQ 16(SP), AX
	MOVQ 24(SP), BX

	// return interface{}
	MOVQ AX, ret_type+0(FP)
	MOVQ BX, ret_data+8(FP)
	RET
```

其中 NO_LOCAL_POINTERS 表示函数没有局部指针变量。同时对返回的接口进行零值初始化，初始化完成后通过 GO_RESULTS_INITIALIZED 告知 GC。这样可以在保证栈分裂时，GC 能够正确处理返回值和局部变量中的指针。


## 3.8.5 goid 的应用: 局部存储

有了 goid 之后，构造 Goroutine 局部存储就非常容易了。我们可以定义一个 gls 包提供 goid 的特性：

```go
package gls

var gls struct {
	m map[int64]map[interface{}]interface{}
	sync.Mutex
}

func init() {
	gls.m = make(map[int64]map[interface{}]interface{})
}
```

gls 包变量简单包装了 map，同时通过 `sync.Mutex` 互斥量支持并发访问。

然后定义一个 getMap 内部函数，用于获取每个 Goroutine 字节的 map：

```go
func getMap() map[interface{}]interface{} {
	gls.Lock()
	defer gls.Unlock()

	goid := GetGoid()
	if m, _ := gls.m[goid]; m != nil {
		return m
	}

	m := make(map[interface{}]interface{})
	gls.m[goid] = m
	return m
}
```

获取到 Goroutine 私有的 map 之后，就是正常的增、删、改操作接口了：

```go
func Get(key interface{}) interface{} {
	return getMap()[key]
}
func Put(key interface{}, v interface{}) {
	getMap()[key] = v
}
func Delete(key interface{}) {
	delete(getMap(), key)
}
```

最后我们再提供一个 Clean 函数，用于释放 Goroutine 对应的 map 资源：

```go
func Clean() {
	gls.Lock()
	defer gls.Unlock()

	delete(gls.m, GetGoid())
}
```

这样一个极简的 Goroutine 局部存储 gls 对象就完成了。

下面是使用局部存储简单的例子：

```go
import (
	gls "path/to/gls"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer gls.Clean()

			defer func() {
				fmt.Printf("%d: number = %d\n", idx, gls.Get("number"))
			}()
			gls.Put("number", idx+100)
		}(i)
	}
	wg.Wait()
}
```

通过 Goroutine 局部存储，不同层次函数之间可以共享存储资源。同时为了避免资源泄漏，需要在 Goroutine 的根函数中，通过 defer 语句调用 gls.Clean() 函数释放资源。

