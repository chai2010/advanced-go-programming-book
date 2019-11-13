# 1.7 Errors and Exceptions

Error handling is an important topic to consider in every programming language. In Go error handling, errors are an important part of the package API and the application's user interface.

There are always some functions in the program that always require a successful run. For example, `strconv.Itoa` converts an integer to a string, reads and writes elements from an array or slice, and reads existing elements from `map`. Such operations will hardly fail at runtime unless there are bugs in the program or catastrophic, unpredictable situations, such as memory leaks at runtime. If you really encounter a real abnormal situation, we can simply terminate the program.

Excluding an anomaly, if the program fails, it is only considered to be one of several expected results. For functions that treat run failures as expected, they return an extra return value, usually the last one to pass the error message. If there is only one reason for the failure, the extra return value can be a Boolean value, usually named ok. For example, when querying a result from a `map`, you can use the extra boolean value to determine whether it is successful:

```go
If v, ok := m["key"]; ok {
Return v
}
```

But there are usually more than one cause of failure, and many times users want to know more about the error. If you only use a simple boolean state value, you will not be able to meet this requirement. In C, the default is an integer type of `errno` to express errors, so you can define multiple error types as needed. In the Go language, `syscall.Errno` is the error corresponding to the `errno` type in the C language. The interface in the `syscall` package, if there is a return error, the underlying is also the `syscall.Errno` error type.

For example, when we modify the file mode through the interface of the `syscall` package, if we encounter an error, we can handle it by forcing the `err` to assert the `syscall.Errno` error type:

```go
Err := syscall.Chmod(":invalid path:", 0666)
If err != nil {
log.Fatal(err.(syscall.Errno))
}
```

We can further obtain the underlying true error type through type queries or type assertions, so that we can get more detailed error information. However, in general, we don't care about the way the error is expressed at the bottom. We just need to know that it is a mistake. When the returned error value is not `nil`, we can get the string type error message by calling the `Error` method of the `error` interface type.

In the Go language, errors are considered to be a predictable result; exceptions are an unintended result, and an exception may indicate a bug in the program or other uncontrollable problems. The Go language recommends using the `recover` function to turn internal exceptions into error handling, which allows users to really care about business-related error handling.

If an interface simply throws all normal errors as exceptions, it will make the error message cluttered and worthless. Just like capturing all of them directly in the `main` function, it makes no sense:

```go
Func main() {
Defer func() {
If r := recover(); r != nil {
log.Fatal(r)
}
}()

...
}
```

Capturing an exception is not the ultimate goal. If the exception is unpredictable, directly outputting the exception information is the best way to handle it.

## 1.7.1 Error Handling Strategy

Let's demonstrate an example of file copying: the function needs to open two files and then copy the contents of one of the files to another file:

```go
Func CopyFile(dstName, srcName string) (written int64, err error) {
Src, err := os.Open(srcName)
If err != nil {
Return
}

Dst, err := os.Create(dstName)
If err != nil {
Return
}

Written, err = io.Copy(dst, src)
dst.Close()
src.Close()
Return
}
```

The above code works, but hides a bug. If the first `os.Open` call succeeds, but the second `os.Create` call fails, it will be returned without releasing the `src` file resource. Although we can fix this bug by adding a `src.Close()` call before the second return statement, similar problems will be hard to find and fix when the code becomes complex. We can use the `defer` statement to ensure that every file that is normally opened can be closed normally:

```go
Func CopyFile(dstName, srcName string) (written int64, err error) {
Src, err := os.Open(srcName)
If err != nil {
Return
}
Defer src.Close()

Dst, err := os.Create(dstName)
If err != nil {
Return
}
Defer dst.Close()

Return io.Copy(dst, src)
}
```

The `defer` statement allows us to think about how to close a file as soon as we open the file. Regardless of how the function returns, the file close statement will always be executed. At the same time, the `defer` statement guarantees that the file can be safely closed even if an exception occurs in `io.Copy`.

As we mentioned earlier, the exported function in the Go language generally does not throw an exception, and an uncontrolled exception can be regarded as a bug in the program. But for those frameworks that offer similar Web services; they often need to access third-party middleware. Because the third-party middleware has a bug, whether the exception will throw an exception, the Web framework itself is not certain. In order to improve the stability of the system, the Web framework generally uses ‘recover` to defensively capture all possible exceptions in the processing flow, and then turn the exception into a normal error return.

Let us take the JSON parser as an example to illustrate the use scenario of recover. Given the complexity of the JSON parser, even if a language parser is working properly, it is not certain that it has no loopholes. Therefore, when an exception occurs, we won't choose to crash the parser. Instead, we will treat the panic exception as a normal parsing error and attach additional information to alert the user to report the error.

```go
Func ParseJSON(input string) (s *Syntax, err error) {
Defer func() {
If p := recover(); p != nil {
Err = fmt.Errorf("JSON: internal error: %v", p)
}
}()
// ...parser...
}
```

The `json` package in the standard library, if it encounters an error when recursively parsing JSON data internally, it will quickly jump out of the deeply nested function call by throwing an exception, and then pass the 'recover' by the interface of the outermost level. `Capture `panic` and return the corresponding error message.

Go language library implementation habits: Even if `panic` is used inside the package, it will be converted to an explicit error value when the function is exported.

## 1.7.2 Getting the wrong context

Sometimes it is easy for the upper level user to understand; the underlying implementer will repack the underlying error as a new error type and return it to the user:

```go
If _, err := html.Parse(resp.Body); err != nil {
Return nil, fmt.Errorf("parsing %s as HTML: %v", url, err)
}
```

When an upper user encounters an error, it is easy to understand the cause of the error from the business level. But the fish and the bear's paw are always difficult to get both. When the upper user gets a new mistake, we also lose the underlying most primitive error type (only the error description information is left).

In order to record information about the type of error in the package's transition, we generally define an auxiliary `WrapError` function that wraps the original error while preserving the full original error type. In order to facilitate the positioning of the problem, and in order to record the state of the function call when the error occurs, we often want to save the complete function call information when a fatal error occurs. At the same time, in order to support cross-network transmission such as RPC, we may need to serialize the error into data similar to JSON format, and then recover the error decoding from the data.

To do this, we can define our own `github.com/chai2010/errors` package, which is the following error type:

```go

Type Error interface {
Caller() []CallerInfo
Wraped() []error
Code() int
Error

Private()
}

Type CallerInfo struct {
FuncName string
FileName string
FileLine int
}
```

Among them, `Error` is an interface type, which is an extension of the `error` interface type. It is used to add call stack information to errors, and supports error multi-level nested wrappers and supports error code formats. For ease of use, we can define the following helper functions:

```go
Func New(msg string) error
Func NewWithCode(code int, msg string) error

Func Wrap(err error, msg string) error
Func WrapWithCode(code int, err error, msg string) error

Func FromJson(json string) (Error, error)
Func ToJson(err error) string
```

`New` is used to build new error types, similar to the `errors.New` function in the standard library, but with the addition of function call stack information when an error occurs. `FromJson` is used to recover the wrong object from the JSON string encoded error. `NewWithCode` is to construct an error with an error code, and also contains the function call stack information when the error occurs. `Wrap` and `WrapWithCode` are error secondary wrappers that wrap the underlying error as a new error, but retain the original underlying error message. The error object returned here can directly call `json.Marshal` to encode the error as a JSON string.

We can use the wrapper function like this:

```go
Import (
"github.com/chai2010/errors"
)

Func loadConfig() error {
_, err := ioutil.ReadFile("/path/to/file")
If err != nil {
Return errors.Wrap(err, "read failed")
}

// ...
}

Func setup() error {
Err := loadConfig()
If err != nil {
Return errors.Wrap(err, "invalid config")
}

// ...
}

Func main() {
If err := setup(); err != nil {
log.Fatal(err)
}

// ...
}
```

In the above example, the error was wrapped in 2 layers. We can traverse the packaging process that the original error has gone through:

```go
For i, e := range err.(errors.Error).Wraped() {
fmt.Printf("wraped(%d): %v\n", i, e)
}
```

You can also get the function call stack information for each wrapper error:

```go
For i, x := range err.(errors.Error).Caller() {
fmt.Printf("caller:%d: %s\n", i, x.FuncName)
}
```

If you need to pass the error over the network, you can encode it as a JSON string with `errors.ToJson(err)`:

```go
// Send error as a JSON string
Func sendError(ch chan<- string, err error) {
Ch <-errors.ToJson(err)
}

// Receive error in JSON string format
Func recvError(ch <-chan string) error {
p, err := errors.FromJson(<-ch)
If err != nil {
log.Fatal(err)
}
Return p
}
```

For web services based on the http protocol, we can also bind a corresponding http status code to the error:

```go
Err := errors.NewWithCode(404, "http error code")

fmt.Println(err)
fmt.Println(err.(errors.Error).Code())
```

In the Go language, error handling also has a unique coding style. After checking if a subfunction has failed, we usually put the logic code that failed to process before the code that processed it successfully. If an error causes the function to return, then the logic code for success should not be placed in the `else` statement block, but should be placed directly in the function body.

```go
f, err := os.Open("filename.ext")
If err != nil {
// In case of failure, return error immediately
}

// normal processing flow
```

The code structure of most functions in the Go language is almost the same, starting with a series of initial checks to prevent errors from occurring, followed by the actual logic of the function.


## 1.7.3 Error returning error

The error in the Go language is an interface type. The interface information contains the original type and the original value. The value of the interface corresponds to `nil` only if the type of the interface and the original value are both empty. In fact, when the type of the interface is empty, the original value must be empty. Conversely, when the original value of the interface is empty, the original type corresponding to the interface is not necessarily empty.

In the following example, I tried to return a custom error type and return `nil` when there are no errors:

```go
Func returnsError() error {
Var p *MyError = nil
If bad() {
p = ErrBad
}
Return p // Will always return a non-nil error.
}
```

However, the result of the final return is actually not `nil`: a normal error, the wrong value is a null pointer of type `MyError`. Here's the improved `returnsError`:

```go
Func returnsError() error {
If bad() {
Return (*MyError)(err)
}
Return nil
}
```

Therefore, when processing the error return value, the error return value is preferably written directly as `nil`.

The Go language is a strongly typed language, and explicit conversions must be made between different types (and must have the same underlying type). However, the `interface` in the Go language is an exception: non-interface types to interface types, or conversions between interface types are implicit. This is to support the duck type, of course, will sacrifice a certain degree of security.

## 1.7.4 Parsing Exceptions

`panic` supports throwing arbitrary types of exceptions (not just `error` type errors). The return value of the `recover` function call is the same as the input parameter type of the `panic` function. Their function signatures are as follows:

```go
Func panic(interface{})
Func recover() interface{}
```

The normal flow of the Go language function call is the result returned by the function execution return statement. There is no exception in this process, so executing the `recover` exception catching function in this process always returns `nil`. The other is the exception flow: When the function calls `panic` to throw an exception, the function will stop executing the subsequent normal statement, but the previously registered `defer` function call will still be executed normally and then returned to the caller. For the caller of the current function, because the exception handling state has not been caught, it is similar to the behavior of calling the `panic` function directly. When an exception occurs, if a `recover` call is made in `defer`, it can capture the parameters when the `panic` is triggered and return to the normal execution flow.

Executing a `recover` call in a non-defer` statement is a common mistake for beginners:

```go
Func main() {
If r := recover(); r != nil {
log.Fatal(r)
}

Panic(123)

If r := recover(); r != nil {
log.Fatal(r)
}
}
```

Both `recover` calls in the above program cannot catch any exceptions. When the first `recover` call is executed, the function must be in the normal non-exception execution flow, at which point the `recover` call will return `nil`. When an exception occurs, the second `recover` call will have no chance to be executed, because the `panic` call will cause the function to return immediately after executing the function that has registered the `defer`.

In fact, the `recover` function call has stricter requirements: we must call `recover` directly in the `defer` function. If the wrapper function of the `recover` function is called in `defer`, the catchup of the exception will fail! For example, sometimes we might want to wrap our own `MyRecover` function, add the necessary log information internally and then call `recover`, which is the wrong approach:

```go
Func main() {
Defer func() {
// Unable to catch exception
If r := MyRecover(); r != nil {
fmt.Println(r)
}
}()
Panic(1)
}

Func MyRecover() interface{} {
log.Println("trace...")
Return recover()
}
```

Similarly, if you call `recover` in a nested `defer` function, it will also fail to catch the exception:

```go
Func main() {
Defer func() {
Defer func() {
// Unable to catch exception
If r := recover(); r != nil {
fmt.Println(r)
}
}()
}()
Panic(1)
}
```

The 2-layer nested `defer` function directly calls `recover` and the 1-layer `defer` function to call the wrapped `MyRecover` function. After two function frames, the real `recover` function is reached. At the time, Goroutine has no abnormal information in the corresponding upper stack frame.

If we call the `MyRecover` function directly in the `defer` statement, it works fine:

```go
Func MyRecover() interface{} {
Return recover()
}

Func main() {
// can catch exceptions normally
Defer MyRecover()
Panic(1)
}
```

However, if the `defer` statement directly calls the `recover` function, the exception will still not be caught properly:

```go
Func main() {
// Unable to catch exception
Defer recover()
Panic(1)
}
```

It must be separated from the stack frame with an exception by a stack frame, and the `recover` function can catch the exception normally. In other words, the `recover` function captures the exception of the grandfather's first call to the function stack frame (just a layer of the `defer` function)!

Of course, in order to avoid the `recover` caller not recognizing the caught exception, you should avoid throwing an exception with `nil` for the argument:

```go
Func main() {
Defer func() {
If r := recover(); r != nil { ... }
// Although it always returns nil, it can restore the abnormal state
}()

// Warning: throw an exception with 'nil` as a parameter
Panic(nil)
}
```

When you want to turn the caught exception into an error, if you want to faithfully return the original information, you need to handle it separately for different types:

```go
Func foo() (err error) {
Defer func() {
If r := recover(); r != nil {
Switch x := r.(type) {
Case string:
Err = errors.New(x)
Case error:
Err = x
Default:
Err = fmt.Errorf("Unknown panic: %v", r)
}
}
}()

Panic("TODO")
}
```

Based on this code template, we can even simulate different types of exceptions. By defining different types of protection interfaces, we can distinguish the types of exceptions:

```go
Func main {
Defer func() {
If r := recover(); r != nil {
Switch x := r.(type) {
Case runtime.Error:
// This is a runtime error type exception
Case error:
// Ordinary error type exception
Default:
// other types of exceptions
}
}
}()

// ...
}
```

But doing so runs counter to the simple and straightforward programming philosophy of Go.