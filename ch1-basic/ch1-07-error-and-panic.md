# 1.7 错误和异常

错误处理是每个编程语言都要考虑的一个重要话题。在 Go 语言的错误处理中，错误是软件包 API 和应用程序用户界面的一个重要组成部分。

在程序中总有一部分函数总是要求必须能够成功的运行。比如 `strconv.Itoa` 将整数转换为字符串，从数组或切片中读写元素，从 `map` 读取已经存在的元素等。这类操作在运行时几乎不会失败，除非程序中有 BUG，或遇到灾难性的、不可预料的情况，比如运行时的内存溢出。如果真的遇到真正异常情况，我们只要简单终止程序就可以了。

排除异常的情况，如果程序运行失败仅被认为是几个预期的结果之一。对于那些将运行失败看作是预期结果的函数，它们会返回一个额外的返回值，通常是最后一个来传递错误信息。如果导致失败的原因只有一个，额外的返回值可以是一个布尔值，通常被命名为 ok。比如，当从一个 `map` 查询一个结果时，可以通过额外的布尔值判断是否成功：

```go
if v, ok := m["key"]; ok {
	return v
}
```

但是导致失败的原因通常不止一种，很多时候用户希望了解更多的错误信息。如果只是用简单的布尔类型的状态值将不能满足这个要求。在 C 语言中，默认采用一个整数类型的 `errno` 来表达错误，这样就可以根据需要定义多种错误类型。在 Go 语言中，`syscall.Errno` 就是对应 C 语言中 `errno` 类型的错误。在 `syscall` 包中的接口，如果有返回错误的话，底层也是 `syscall.Errno` 错误类型。

比如我们通过 `syscall` 包的接口来修改文件的模式时，如果遇到错误我们可以通过将 `err` 强制断言为 `syscall.Errno` 错误类型来处理：

```go
err := syscall.Chmod(":invalid path:", 0666)
if err != nil {
	log.Fatal(err.(syscall.Errno))
}
```

我们还可以进一步地通过类型查询或类型断言来获取底层真实的错误类型，这样就可以获取更详细的错误信息。不过一般情况下我们并不关心错误在底层的表达方式，我们只需要知道它是一个错误就可以了。当返回的错误值不是 `nil` 时，我们可以通过调用 `error` 接口类型的 `Error` 方法来获得字符串类型的错误信息。

在 Go 语言中，错误被认为是一种可以预期的结果；而异常则是一种非预期的结果，发生异常可能表示程序中存在 BUG 或发生了其它不可控的问题。Go 语言推荐使用 `recover` 函数将内部异常转为错误处理，这使得用户可以真正的关心业务相关的错误处理。

如果某个接口简单地将所有普通的错误当做异常抛出，将会使错误信息杂乱且没有价值。就像在 `main` 函数中直接捕获全部一样，是没有意义的：

```go
func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	...
}
```

捕获异常不是最终的目的。如果异常不可预测，直接输出异常信息是最好的处理方式。

## 1.7.1 错误处理策略

让我们演示一个文件复制的例子：函数需要打开两个文件，然后将其中一个文件的内容复制到另一个文件：

```go
func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}

	dst, err := os.Create(dstName)
	if err != nil {
		return
	}

	written, err = io.Copy(dst, src)
	dst.Close()
	src.Close()
	return
}
```

上面的代码虽然能够工作，但是隐藏一个 bug。如果第一个 `os.Open` 调用成功，但是第二个 `os.Create` 调用失败，那么会在没有释放 `src` 文件资源的情况下返回。虽然我们可以通过在第二个返回语句前添加 `src.Close()` 调用来修复这个 BUG；但是当代码变得复杂时，类似的问题将很难被发现和修复。我们可以通过 `defer` 语句来确保每个被正常打开的文件都能被正常关闭：

```go
func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(dstName)
	if err != nil {
		return
	}
	defer dst.Close()

	return io.Copy(dst, src)
}
```

`defer`语句可以让我们在打开文件时马上思考如何关闭文件。不管函数如何返回，文件关闭语句始终会被执行。同时 `defer` 语句可以保证，即使 `io.Copy` 发生了异常，文件依然可以安全地关闭。

前文我们说到，Go 语言中的导出函数一般不抛出异常，一个未受控的异常可以看作是程序的 BUG。但是对于那些提供类似 Web 服务的框架而言；它们经常需要接入第三方的中间件。因为第三方的中间件是否存在 BUG 是否会抛出异常，Web 框架本身是不能确定的。为了提高系统的稳定性，Web 框架一般会通过 `recover` 来防御性地捕获所有处理流程中可能产生的异常，然后将异常转为普通的错误返回。

让我们以 JSON 解析器为例，说明 recover 的使用场景。考虑到 JSON 解析器的复杂性，即使某个语言解析器目前工作正常，也无法肯定它没有漏洞。因此，当某个异常出现时，我们不会选择让解析器崩溃，而是会将 panic 异常当作普通的解析错误，并附加额外信息提醒用户报告此错误。

```go
func ParseJSON(input string) (s *Syntax, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("JSON: internal error: %v", p)
		}
	}()
	// ...parser...
}
```

标准库中的 `json` 包，在内部递归解析 JSON 数据的时候如果遇到错误，会通过抛出异常的方式来快速跳出深度嵌套的函数调用，然后由最外一级的接口通过 `recover` 捕获 `panic`，然后返回相应的错误信息。

Go 语言库的实现习惯: 即使在包内部使用了 `panic`，但是在导出函数时会被转化为明确的错误值。

## 1.7.2 获取错误的上下文

有时候为了方便上层用户理解；底层实现者会将底层的错误重新包装为新的错误类型返回给用户：

```go
if _, err := html.Parse(resp.Body); err != nil {
	return nil, fmt.Errorf("parsing %s as HTML: %v", url,err)
}
```

上层用户在遇到错误时，可以很容易从业务层面理解错误发生的原因。但是鱼和熊掌总是很难兼得，在上层用户获得新的错误的同时，我们也丢失了底层最原始的错误类型（只剩下错误描述信息了）。

为了记录这种错误类型在包装的变迁过程中的信息，我们一般会定义一个辅助的 `WrapError` 函数，用于包装原始的错误，同时保留完整的原始错误类型。为了问题定位的方便，同时也为了能记录错误发生时的函数调用状态，我们很多时候希望在出现致命错误的时候保存完整的函数调用信息。同时，为了支持 RPC 等跨网络的传输，我们可能要需要将错误序列化为类似 JSON 格式的数据，然后再从这些数据中将错误解码恢复出来。

为此，我们可以定义自己的 `github.com/chai2010/errors` 包，里面是以下的错误类型：

```go

type Error interface {
	Caller() []CallerInfo
	Wraped() []error
	Code() int
	error

	private()
}

type CallerInfo struct {
	FuncName string
	FileName string
	FileLine int
}
```

其中 `Error` 为接口类型，是 `error` 接口类型的扩展，用于给错误增加调用栈信息，同时支持错误的多级嵌套包装，支持错误码格式。为了使用方便，我们可以定义以下的辅助函数：

```go
func New(msg string) error
func NewWithCode(code int, msg string) error

func Wrap(err error, msg string) error
func WrapWithCode(code int, err error, msg string) error

func FromJson(json string) (Error, error)
func ToJson(err error) string
```

`New`用于构建新的错误类型，和标准库中 `errors.New` 功能类似，但是增加了出错时的函数调用栈信息。`FromJson` 用于从 JSON 字符串编码的错误中恢复错误对象。`NewWithCode` 则是构造一个带错误码的错误，同时也包含出错时的函数调用栈信息。`Wrap` 和 `WrapWithCode` 则是错误二次包装函数，用于将底层的错误包装为新的错误，但是保留的原始的底层错误信息。这里返回的错误对象都可以直接调用 `json.Marshal` 将错误编码为 JSON 字符串。

我们可以这样使用包装函数:

```go
import (
	"github.com/chai2010/errors"
)

func loadConfig() error {
	_, err := ioutil.ReadFile("/path/to/file")
	if err != nil {
		return errors.Wrap(err, "read failed")
	}

	// ...
}

func setup() error {
	err := loadConfig()
	if err != nil {
		return errors.Wrap(err, "invalid config")
	}

	// ...
}

func main() {
	if err := setup(); err != nil {
		log.Fatal(err)
	}

	// ...
}
```

上面的例子中，错误被进行了 2 层包装。我们可以这样遍历原始错误经历了哪些包装流程：

```go
	for i, e := range err.(errors.Error).Wraped() {
		fmt.Printf("wrapped(%d): %v\n", i, e)
	}
```

同时也可以获取每个包装错误的函数调用堆栈信息：

```go
	for i, x := range err.(errors.Error).Caller() {
		fmt.Printf("caller:%d: %s\n", i, x.FuncName)
	}
```

如果需要将错误通过网络传输，可以用 `errors.ToJson(err)` 编码为 JSON 字符串：

```go
// 以 JSON 字符串方式发送错误
func sendError(ch chan<- string, err error) {
	ch <- errors.ToJson(err)
}

// 接收 JSON 字符串格式的错误
func recvError(ch <-chan string) error {
	p, err := errors.FromJson(<-ch)
	if err != nil {
		log.Fatal(err)
	}
	return p
}
```

对于基于 http 协议的网络服务，我们还可以给错误绑定一个对应的 http 状态码：

```go
err := errors.NewWithCode(404, "http error code")

fmt.Println(err)
fmt.Println(err.(errors.Error).Code())
```

在 Go 语言中，错误处理也有一套独特的编码风格。检查某个子函数是否失败后，我们通常将处理失败的逻辑代码放在处理成功的代码之前。如果某个错误会导致函数返回，那么成功时的逻辑代码不应放在 `else` 语句块中，而应直接放在函数体中。

```go
f, err := os.Open("filename.ext")
if err != nil {
	// 失败的情形, 马上返回错误
}

// 正常的处理流程
```

Go 语言中大部分函数的代码结构几乎相同，首先是一系列的初始检查，用于防止错误发生，之后是函数的实际逻辑。


## 1.7.3 错误的错误返回

Go 语言中的错误是一种接口类型。接口信息中包含了原始类型和原始的值。只有当接口的类型和原始的值都为空的时候，接口的值才对应 `nil`。其实当接口中类型为空的时候，原始值必然也是空的；反之，当接口对应的原始值为空的时候，接口对应的原始类型并不一定为空的。

在下面的例子中，试图返回自定义的错误类型，当没有错误的时候返回 `nil`：

```go
func returnsError() error {
	var p *MyError = nil
	if bad() {
		p = ErrBad
	}
	return p // Will always return a non-nil error.
}
```

但是，最终返回的结果其实并非是 `nil`：是一个正常的错误，错误的值是一个 `MyError` 类型的空指针。下面是改进的 `returnsError`：

```go
func returnsError() error {
	if bad() {
		return (*MyError)(err)
	}
	return nil
}
```

因此，在处理错误返回值的时候，没有错误的返回值最好直接写为 `nil`。

Go 语言作为一个强类型语言，不同类型之间必须要显式的转换（而且必须有相同的基础类型）。但是，Go 语言中 `interface` 是一个例外：非接口类型到接口类型，或者是接口类型之间的转换都是隐式的。这是为了支持鸭子类型，当然会牺牲一定的安全性。

## 1.7.4 剖析异常

`panic`支持抛出任意类型的异常（而不仅仅是 `error` 类型的错误），`recover` 函数调用的返回值和 `panic` 函数的输入参数类型一致，它们的函数签名如下：

```go
func panic(interface{})
func recover() interface{}
```

Go 语言函数调用的正常流程是函数执行返回语句返回结果，在这个流程中是没有异常的，因此在这个流程中执行 `recover` 异常捕获函数始终是返回 `nil`。另一种是异常流程: 当函数调用 `panic` 抛出异常，函数将停止执行后续的普通语句，但是之前注册的 `defer` 函数调用仍然保证会被正常执行，然后再返回到调用者。对于当前函数的调用者，因为处理异常状态还没有被捕获，和直接调用 `panic` 函数的行为类似。在异常发生时，如果在 `defer` 中执行 `recover` 调用，它可以捕获触发 `panic` 时的参数，并且恢复到正常的执行流程。

在非 `defer` 语句中执行 `recover` 调用是初学者常犯的错误:

```go
func main() {
	if r := recover(); r != nil {
		log.Fatal(r)
	}

	panic(123)

	if r := recover(); r != nil {
		log.Fatal(r)
	}
}
```

上面程序中两个 `recover` 调用都不能捕获任何异常。在第一个 `recover` 调用执行时，函数必然是在正常的非异常执行流程中，这时候 `recover` 调用将返回 `nil`。发生异常时，第二个 `recover` 调用将没有机会被执行到，因为 `panic` 调用会导致函数马上执行已经注册 `defer` 的函数后返回。

其实 `recover` 函数调用有着更严格的要求：我们必须在 `defer` 函数中直接调用 `recover`。如果 `defer` 中调用的是 `recover` 函数的包装函数的话，异常的捕获工作将失败！比如，有时候我们可能希望包装自己的 `MyRecover` 函数，在内部增加必要的日志信息然后再调用 `recover`，这是错误的做法：

```go
func main() {
	defer func() {
		// 无法捕获异常
		if r := MyRecover(); r != nil {
			fmt.Println(r)
		}
	}()
	panic(1)
}

func MyRecover() interface{} {
	log.Println("trace...")
	return recover()
}
```

同样，如果是在嵌套的 `defer` 函数中调用 `recover` 也将导致无法捕获异常：

```go
func main() {
	defer func() {
		defer func() {
			// 无法捕获异常
			if r := recover(); r != nil {
				fmt.Println(r)
			}
		}()
	}()
	panic(1)
}
```

2 层嵌套的 `defer` 函数中直接调用 `recover` 和 1 层 `defer` 函数中调用包装的 `MyRecover` 函数一样，都是经过了 2 个函数帧才到达真正的 `recover` 函数，这个时候 Goroutine 的对应上一级栈帧中已经没有异常信息。

如果我们直接在 `defer` 语句中调用 `MyRecover` 函数又可以正常工作了：

```go
func MyRecover() interface{} {
	return recover()
}

func main() {
	// 可以正常捕获异常
	defer MyRecover()
	panic(1)
}
```

但是，如果 `defer` 语句直接调用 `recover` 函数，依然不能正常捕获异常：

```go
func main() {
	// 无法捕获异常
	defer recover()
	panic(1)
}
```

必须要和有异常的栈帧只隔一个栈帧，`recover` 函数才能正常捕获异常。换言之，`recover` 函数捕获的是祖父一级调用函数栈帧的异常（刚好可以跨越一层 `defer` 函数）！

当然，为了避免 `recover` 调用者不能识别捕获到的异常, 应该避免用 `nil` 为参数抛出异常:

```go
func main() {
	defer func() {
		if r := recover(); r != nil { ... }
		// 虽然总是返回 nil, 但是可以恢复异常状态
	}()

	// 警告: 用 nil 为参数抛出异常
	panic(nil)
}
```

当希望将捕获到的异常转为错误时，如果希望忠实返回原始的信息，需要针对不同的类型分别处理：

```go
func foo() (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("Unknown panic: %v", r)
			}
		}
	}()

	panic("TODO")
}
```

基于这个代码模板，我们甚至可以模拟出不同类型的异常。通过为定义不同类型的保护接口，我们就可以区分异常的类型了：

```go
func main {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case runtime.Error:
				// 这是运行时错误类型异常
			case error:
				// 普通错误类型异常
			default:
				// 其他类型异常
			}
		}
	}()

	// ...
}
```

不过这样做和 Go 语言简单直接的编程哲学背道而驰了。
