# Appendix A: Go language common pit

The Go pits listed here are all in accordance with the Go language syntax and can be compiled normally, but may be the result of running errors or the risk of resource leaks.

## Variable parameter is an empty interface type

When the variable parameter of the parameter is a null interface type, you need to pay attention to the parameter expansion when you import the slice of the empty interface.

```go
Func main() {
Var a = []interface{}{1, 2, 3}

fmt.Println(a)
fmt.Println(a...)
}
```

The compiler can't find the error, regardless of whether it is expanded, but the output is different:

```
[1 2 3]
1 2 3
```

## Array is value passing

In the function call argument, the array is a value passed, and the result cannot be returned by modifying the argument of the array type.

```go
Func main() {
x := [3]int{1, 2, 3}

Func(arr [3]int) {
Arr[0] = 7
fmt.Println(arr)
}(x)

fmt.Println(x)
}
```

Slices are required if necessary.

## map traversal is not fixed in order

Map is a hash table implementation, and the order of each traversal may be different.

```go
Func main() {
m := map[string]string{
"1": "1",
		"twenty two",
"3": "3",
}

For k, v := range m {
Println(k, v)
}
}
```

## Return value is blocked

In the local scope, the local variable of the same name within the named return value is masked:

```go
Func Foo() (err error) {
If err := Bar(); err != nil {
Return
}
Return
}
```

## recover must be run in the defer function

Recover captures the exception when the grandfather calls, invalid when called directly:

```go
Func main() {
Recover()
Panic(1)
}
```

Direct defer calls are also invalid:

```go
Func main() {
Defer recover()
Panic(1)
}
```

Multi-level nesting is still invalid when defer is called:

```go
Func main() {
Defer func() {
Func() { recover() }()
}()
Panic(1)
}
```

Must be called directly in the defer function to be valid:

```go
Func main() {
Defer func() {
Recover()
}()
Panic(1)
}
```

## main function exit early

Goroutine is not guaranteed to complete the task.

```go
Func main() {
Go println("hello")
}
```

## Avoiding problems in concurrency through Sleep

Hibernate does not guarantee the output of a complete string:

```go
Func main() {
Go println("hello")
time.Sleep(time.Second)
}
```

Similarly, by inserting a dispatch statement:

```go
Func main() {
Go println("hello")
runtime.Gosched()
}
```

## Exclusive CPU causes other Goroutine to starve

Goroutine is a collaborative preemptive schedule, and Goroutine itself does not actively give up the CPU:

```go
Func main() {
runtime.GOMAXPROCS(1)

Go func() {
For i := 0; i < 10; i++ {
fmt.Println(i)
}
}()

For {} // occupy CPU
}
```

The solution is to add the runtime.Gosched() dispatch function to the for loop:

```go
Func main() {
runtime.GOMAXPROCS(1)

Go func() {
For i := 0; i < 10; i++ {
fmt.Println(i)
}
}()

For {
runtime.Gosched()
}
}
```

Or avoid CPU usage by blocking:

```go
Func main() {
runtime.GOMAXPROCS(1)

Go func() {
For i := 0; i < 10; i++ {
fmt.Println(i)
}
os.Exit(0)
}()

Select{}
}
```

## Different Goroutine does not satisfy the orderly consistent memory model

Because in different Goroutine, the main function cannot guarantee to print out `hello, world`:

```go
Var msg string
Var done bool

Func setup() {
Msg = "hello, world"
Done = true
}

Func main() {
Go setup()
For !done {
}
Println(msg)
}
```

The solution is to use explicit synchronization:

```go
Var msg string
Var done = make(chan bool)

Func setup() {
Msg = "hello, world"
Done <- true
}

Func main() {
Go setup()
<-done
Println(msg)
}
```
The msg write is before the channel is sent, so it can guarantee to print `hello, world`

## Closure error references the same variable

```go
Func main() {
For i := 0; i < 5; i++ {
Defer func() {
Println(i)
}()
}
}
```

The improved method is to generate a local variable in each iteration:

```go
Func main() {
For i := 0; i < 5; i++ {
i := i
Defer func() {
Println(i)
}()
}
}
```

Or pass in via function argument:

```go
Func main() {
For i := 0; i < 5; i++ {
Defer func(i int) {
Println(i)
}(i)
}
}
```

## Executing a defer statement inside a loop

Defer can only be executed when the function exits, and executing defer in for will cause the resource to be delayed:

```go
Func main() {
For i := 0; i < 5; i++ {
f, err := os.Open("/path/to/file")
If err != nil {
log.Fatal(err)
}
Defer f.Close()
}
}
```

The solution can be to construct a local function in for, and execute defer inside the local function:

```go
Func main() {
For i := 0; i < 5; i++ {
Func() {
f, err := os.Open("/path/to/file")
If err != nil {
log.Fatal(err)
}
Defer f.Close()
}()
}
}
```

## Slices cause the entire underlying array to be locked

Slices cause the entire underlying array to be locked, and the underlying array cannot free memory. If the underlying array is large, it will put a lot of pressure on the memory.

```go
Func main() {
headerMap := make(map[string][]byte)

For i := 0; i < 5; i++ {
Name := "/path/to/file"
Data, err := ioutil.ReadFile(name)
If err != nil {
log.Fatal(err)
}
headerMap[name] = data[:1]
}

// do some thing
}
```

The solution is to clone the result so that the underlying array can be freed:

```go
Func main() {
headerMap := make(map[string][]byte)

For i := 0; i < 5; i++ {
Name := "/path/to/file"
Data, err := ioutil.ReadFile(name)
If err != nil {
log.Fatal(err)
}
headerMap[name] = append([]byte{}, data[:1]...)
}

// do some thing
}
```

## Empty pointer and empty interface are not equivalent

For example, it returns an error pointer, but it is not an empty error interface:

```go
Func returnsError() error {
Var p *MyError = nil
If bad() {
p = ErrBad
}
Return p // Will always return a non-nil error.
}
```

## Memory address will change

The address of an object in the Go language may change, so the pointer cannot be generated from other non-pointer types:

```go
Func main() {
Var x int = 42
Var p uintptr = uintptr(unsafe.Pointer(&x))

runtime.GC()
Var px *int = (*int)(unsafe.Pointer(p))
Println(*px)
}
```

When the memory is sent, the associated pointer will be updated synchronously, but the uintptr of the non-pointer type will not be updated synchronously.

Similarly, the Go object address cannot be saved in CGO.

## Goroutine leaked

The Go language is characterized by automatic memory reclamation, so memory is generally not leaking. However, Goroutine does have a leak, and the memory referenced by the leaked Goroutine cannot be recycled.

```go
Func main() {
Ch := func() <-chan int {
Ch := make(chan int)
Go func() {
For i := 0; ; i++ {
Ch <- i
}
} ()
Return ch
}()

For v := range ch {
fmt.Println(v)
If v == 5 {
Break
}
}
}
```

In the above program, the background Goroutine inputs the natural number sequence to the pipeline, and the main function outputs the sequence. But when the break jumps out of the for loop, the background Goroutine is in a state where it cannot be recycled.

We can avoid this problem through the context package:


```go
Func main() {
ctx, cancel := context.WithCancel(context.Background())

ch := func(ctx context.Context) <-chan int {
ch := make(chan int)
go func() {
for i := 0; ; i++ {
select {
case <- ctx.Done():
return
case ch <- i:
}
}
} ()
return ch
}(ctx)

for v := range ch {
fmt.Println(v)
if v == 5 {
Cancel()
Break
}
}
}
```

When the main function jumps out of the loop in break, it calls the `cancel()` to notify the background Goroutine to exit, thus avoiding the Goroutine leak.