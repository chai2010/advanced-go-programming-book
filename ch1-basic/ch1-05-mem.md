# 1.5 Concurrent-oriented memory model

In the early days, the CPU executed machine instructions in the form of a single core. The ancestor C language of the Go language is the representative of this sequential programming language. The order in a sequential programming language means that all instructions are executed in a serial manner, and there is only one CPU executing instructions of the program at the same time.

With the development of processor technology, the single-core era has encountered bottlenecks in ways to increase processor frequency to improve operational efficiency. Currently, various mainstream CPU frequencies are basically locked in the vicinity of 3GHZ. The stagnation of the development of single-core CPUs has brought opportunities for the development of multi-core CPUs. Accordingly, programming languages ​​are beginning to evolve in the direction of parallelization. The Go language is a native programming language that supports concurrency in the context of multi-core and networking.

Common parallel programming has a variety of models, mainly multi-threading, messaging, and so on. In theory, multithreading and message-based concurrent programming are equivalent. Since the multi-threaded concurrency model can naturally correspond to multi-core processors, the mainstream operating systems therefore provide system-level multi-threading support, and conceptually multi-threading seems to be more intuitive, so the multi-threaded programming model is gradually absorbed. Go to mainstream programming language features or language extension libraries. The mainstream programming language supports less of the message-based concurrent programming model. The Erlang language supports the representative of the message-based concurrent programming model, and its concurrency does not share memory. The Go language is a collection of message-concurrency models. It integrates CSP-based concurrency programming into the language. It is easy to start a Goroutine with a go keyword. Unlike Erlang, Gor language is shared between Go languages. Memory.

## 1.5.1 Goroutine and system threads

Goroutine is a unique concurrency of the Go language. It is a lightweight thread that is launched by the go keyword. In real Go implementations, goroutines and system threads are not equivalent. Although the difference between the two is actually only a quantitative difference, it is this quantitative change that has led to a qualitative leap in Go language programming.

First, each system-level thread will have a fixed-size stack (generally the default may be 2MB). This stack is mainly used to store parameters and local variables when the function is recursively called. Fixed the size of the stack leads to two problems: one is a huge waste for many threads that only need a small stack space, and the other is the risk of stack overflow for a few threads that need huge stack space. . The solution to these two problems is to either reduce the fixed stack size and increase the space utilization; or increase the size of the stack to allow deeper function recursive calls, but the two cannot be combined at the same time. Instead, a Goroutine will start with a small stack (possibly 2KB or 4KB), and when it encounters deep recursion, the current stack space is insufficient, Goroutine will dynamically scale the stack as needed (the maximum stack size in the mainstream implementation) Can reach 1GB). Because the cost of booting is small, we can easily launch thousands of Goroutines.

Go's runtime also includes its own scheduler, which uses a number of techniques to multiplex m Goroutines on n operating system threads. The work of the Go scheduler is similar to the scheduling of the kernel, but this scheduler only focuses on Goroutine in a separate Go program. Goroutine uses semi-preemptive cooperative scheduling, which only causes scheduling when the current Goroutine is blocked. At the same time, it occurs in the user mode. The scheduler only saves the necessary registers according to the specific function, and the switching cost is lower than the system thread. many. The runtime has a `runtime.GOMAXPROCS` variable that controls the number of system threads that are currently running a normal non-blocking Goroutine.

Starting a Goroutine in Go is not only as simple as calling a function, but also the cost of scheduling between Goroutines. These factors greatly contribute to the popularity and development of concurrent programming.

## 1.5.2 Atomic operation

The so-called atomic operation is the "minimum and non-parallelizable" operation in concurrent programming. In general, if multiple concurrent operations on the same shared resource are atomic, then at most one concurrent entity can operate on the resource at the same time. From a thread perspective, other threads cannot access the resource while the current thread is modifying the shared resource. Atomic operations For multi-threaded concurrent programming models, there is no surprise that is different from single-threading, and the integrity of shared resources can be guaranteed.

In general, atomic operations are guaranteed by "mutually exclusive" access, usually protected by special CPU instructions. Of course, if you just want to simulate a coarse-grained atomic operation, we can do so with the help of `sync.Mutex`:

```go
Import (
"sync"
)

Var total struct {
sync.Mutex
Value int
}

Func worker(wg *sync.WaitGroup) {
Defer wg.Done()

For i := 0; i <= 100; i++ {
total.Lock()
Total.value += i
total.Unlock()
}
}

Func main() {
Var wg sync.WaitGroup
wg.Add(2)
Go worker(&wg)
Go worker(&wg)
wg.Wait()

fmt.Println(total.value)
}
```

In the `worker` loop, in order to guarantee the atomicity of `total.value += i`, we lock and unlock by `sync.Mutex` to ensure that the statement is only accessed by one thread at a time. For programs with a multi-threaded model, it is necessary to lock and unlock before and after entering and leaving the critical section. Without the protection of locks, the final value of `total` will probably be incorrect due to competition between multiple threads.

Using a mutex to protect a numeric shared resource is cumbersome and inefficient. The standard library's `sync/atomic` package provides rich support for atomic operations. We can reimplement the above example:

```go
Import (
"sync"
"sync/atomic"
)

Var total uint64

Func worker(wg *sync.WaitGroup) {
Defer wg.Done()

Var i uint64
For i = 0; i <= 100; i++ {
atomic.AddUint64(&total, i)
}
}

Func main() {
Var wg sync.WaitGroup
wg.Add(2)

Go worker(&wg)
Go worker(&wg)
wg.Wait()
}
```

The `atomic.AddUint64` function call guarantees that reading, updating, and saving of `total` is an atomic operation, so accessing in multiple threads is also safe.

Atomic operation with a mutex can achieve a very efficient single-piece mode. The cost of a mutex is much higher than that of a normal integer. You can add a numeric flag in a performance-sensitive place to improve performance by reducing the number of mutex locks by atomic detection.

```go
Type singleton struct {}

Var (
Instance *singleton
Initialized uint32
Mu sync.Mutex
)

Func Instance() *singleton {
If atomic.LoadUint32(&initialized) == 1 {
Return instance
}

mu.Lock()
Defer mu.Unlock()

If instance == nil {
Defer atomic.StoreUint32(&initialized, 1)
Instance = &singleton{}
}
Return instance
}
```

We can extract the generic code and become the implementation of `sync.Once` in the standard library:

```go
Type Once struct {
m Mutex
Done uint32
}

Func (o *Once) Do(f func()) {
If atomic.LoadUint32(&o.done) == 1 {
Return
}

o.m.Lock()
Defer o.m.Unlock()

If o.done == 0 {
Defer atomic.StoreUint32(&o.done, 1)
f()
}
}
```

Reimplementing the singleton mode based on `sync.Once`:

```go
Var (
Instance *singleton
Once sync.Once
)

Func Instance() *singleton {
once.Do(func() {
Instance = &singleton{}
})
Return instance
}
```

The `sync/atomic` package provides atomic operations support for basic numeric types and reading and writing complex objects. The `atomic.Value` atomic object provides two atomic methods `Load` and `Store` for loading and saving data. The return value and parameters are both `interface{}` types, so they can be used for any customization. Complex type.

```go
Var config atomic.Value // save the current configuration information

/ / Initialize the configuration information
config.Store(loadConfig())

/ / Start a background thread, load the updated configuration information
Go func() {
For {
time.Sleep(time.Second)
config.Store(loadConfig())
}
}()

// The worker thread used to process the request always uses the latest configuration information
For i := 0; i < 10; i++ {
Go func() {
For r := range requests() {
c := config.Load()
// ...
}
}()
}
```

This is a simplified producer consumer model: the background thread generates the latest configuration information; the front-end multiple worker threads get the latest configuration information. All threads share configuration information resources.

## 1.5.3 Sequential Consistent Memory Model

If you just want to synchronize data between threads, atomic operations have provided some synchronization guarantees for programmers. However, this guarantee has a premise: a sequential consistency memory model. To understand the order consistency, let's take a look at a simple example:

```go
Var a string
Var done bool

Func setup() {
a = "hello, world"
Done = true
}

Func main() {
Go setup()
For !done {}
Print(a)
}
```

We created the `setup` thread for initializing the string `a`. After the initialization is complete, set the `done` flag to `true`. In the main thread where the `main` function is located, when `for !done {}` detects that `done` becomes `true`, the string initialization is considered complete, and then the string is printed.

However, the Go language does not guarantee that the write operation to `done` observed in the `main` function occurs after the write operation to the string `a`, so the program is likely to print an empty string. Worse, because there is no synchronization event between the two threads, the `setup` thread's write operation to `done` can't even be seen by the `main` thread, and the `main` function may fall into an infinite loop.

In Go, the same sequential memory model is guaranteed inside the same Goroutine thread. However, between different Goroutines, the sequential consistency memory model is not satisfied, and a well-defined synchronization event is needed as a reference for synchronization. If two events are not sortable, then the two events are said to be concurrent. In order to maximize parallelism, the Go language compiler and processor may reorder the execution statements without affecting the above rules (the CPU will also perform some instructions out of order).

Therefore, if you execute `a = 1; b = 2;` two statements sequentially in a Goroutine, although the current Goroutine can be considered as `a = 1;` statement before the `b = 2;` statement, but In another Goroutine, the `b = 2;` statement may be executed before the `a = 1;` statement, and even in another Goroutine it is not visible (possibly always in the register). That is to say, in another Goroutine's view, `a = 1; b = 2;` The execution order of the two statements is undefined. If a concurrent program cannot determine the order relationship of events, then The results of the sequence often have uncertain results. For example, the following program:


```go
Func main() {
Go println("Hello, World")
}
```

According to the Go language specification, the program ends when the `main` function exits, and does not wait for any background threads. Because the execution of Goroutine and the return event of the `main` function are concurrent, anyone can happen first, so when to print, whether it can be printed is unknown.

Using the previous atomic operation does not solve the problem because we cannot determine the order between the two atomic operations. The solution to the problem is to explicitly sort the two events by synchronizing the primitives:

```go
Func main() {
Done := make(chan int)

Go func(){
Println("Hello, World")
Done <- 1
}()

<-done
}
```

When `<-done` is executed, it is necessary to require `done <- 1` to be executed. According to the same Gorouine still meets the order consistency rules, we can judge that when `done <- 1` is executed, the `println("hello, world")` statement must have been executed. Therefore, the current program ensures that the results can be printed normally.

Of course, synchronization can also be achieved with the `sync.Mutex` mutex:

```go
Func main() {
Var mu sync.Mutex

mu.Lock()
Go func(){
Println("Hello, World")
mu.Unlock()
}()

mu.Lock()
}
```

It can be determined that the `mu.Unlock()` of the background thread must occur after the completion of `println("hello, world") (the same thread satisfies the order consistency), the second `mu.Lock of the `main` function. () must occur after the background thread's `mu.Unlock()` (`sync.Mutex` guarantee), at this point the background thread's print job has been successfully completed.

## 1.5.4 Initialization sequence

In the previous function chapter, we have briefly introduced the initialization sequence of the program, which is the basic specification of the Go language-oriented concurrent memory model.

The initialization and execution of the Go program always starts with the `main.main` function. However, if other packages are imported into the `main` package, they will be included in the `main` package in order. The import order here depends on the implementation. It may be imported in the string order of the file name or package path name. ). If a package is imported multiple times, it will only be imported once during execution. When a package is imported, if it also imports other packages, it first includes the other packages, then creates and initializes the constants and variables of the package. Then is to call the `init` function in the package. If a package has multiple `init` functions, the implementation may be called in the order of file names. Multiple `init`s in the same file are in the order in which they appear. Call (`init` is not a normal function, you can define more than one, so it can't be called by other functions). Finally, all package constants, package variables in the `main` package are created and initialized, and the `init` function is executed before the `main.main` function is entered and the program starts executing normally. The following figure is a schematic diagram of the startup sequence of the Go program function:

![](../images/ch1-12-init.ditaa.png)

*Figure 1-12 Package Initialization Process*

It should be noted that all code runs in the same Goroutine before the main.main` function is executed, and is also running in the main system thread of the program. If a `init` function internally starts a new Goroutine with the go keyword, the new Goroutine and `main.main` functions are executed concurrently.

Because all of the `init` and `main` functions are done in the main thread, they also satisfy the order consistency model.

## 1.5.5 Goroutine creation

The `go` statement will create a new Goroutine before the current Goroutine counterpart returns. For example:


```go
Var a string

Func f() {
Print(a)
}

Func hello() {
a = "hello, world"
Go f()
}
```

Executing the `go f()` statement creates Goroutine and `hello` functions that are executed in the same Goroutine. Depending on the order in which the statements are written, it can be determined that Goroutine creation occurs before the `hello` function returns, but the newly created Goroutine corresponds to `f The execution events of () and the events returned by the `hello` function are not sortable, that is, concurrent. Calling `hello` may print `"hello, world"` at some point in the future, or it may be printed after the `hello` function has finished executing.

## 1.5.6 Channel-based communication

Channel communication is the primary method of synchronizing between Goroutines. Each send operation on an unbuffered Channel is paired with its corresponding receive operation. Send and receive operations usually occur on different Goroutines (two operations on the same Goroutine can easily lead to deadlocks). **Send operations on uncached channels always occur before the corresponding receive operation is completed.**


```go
Var done = make(chan bool)
Var msg string

Func aGoroutine() {
Msg = "Hello, the world"
Done <- true
}

Func main() {
Go aGoroutine()
<-done
Println(msg)
}
```

Guaranteed to print out "hello, world". The program first writes to `msg`, then sends a sync signal on the `done` pipeline, then receives the corresponding sync signal from `done`, and finally executes the `println` function.

If the data is received from the channel after the channel is closed, the receiver will receive the zero value returned by the channel. So in this example, closing the pipe with `close(c)` instead of `done <- false` still guarantees that the program will behave the same.

```go
Var done = make(chan bool)
Var msg string

Func aGoroutine() {
Msg = "Hello, the world"
Close(done)
}

Func main() {
Go aGoroutine()
<-done
Println(msg)
}
```

** For reception from an unbuffered channel, occurs before the transmission to the Channel is completed. **

Based on the above rules, it is possible to exchange receive and send operations in two Goroutines (but it is dangerous):

```go
Var done = make(chan bool)
Var msg string

Func aGoroutine() {
Msg = "hello, world"
<-done
}
Func main() {
Go aGoroutine()
Done <- true
Println(msg)
}
```

It is also guaranteed to print "hello, world". Because the `done <- true` in the `main` thread is sent, the background thread `<-done` has already started, which guarantees that `msg = "hello, world"` is executed, so after `println(msg)` The msg has been assigned. In short, the background thread first writes to `msg`, then receives the signal from `done`, then the `main` thread sends the corresponding signal to `done`, and finally executes the `println` function. However, if the Channel is buffered (for example, `done = make(chan bool, 1)`), the `done <- true` receive operation of the `main` thread will not be ``ddone` by the background thread. The receiving operation is blocked and the program cannot guarantee to print out "hello, world".

For buffered Channels, the `k` receive completion operations for the Channel occur before the completion of the `K+C' transmit operations, where `C` is the buffer size of the Channel. ** If `C` is set to 0, it corresponds to the unbuffered Channel, even if the Kth reception is completed before the Kth transmission is completed. Since the unbuffered Channel can only be sent synchronously, it is reduced to the rule of the previously unbuffered Channel: ** For the reception from the unbuffered Channel, before the transmission to the Channel is completed. **

We can control the maximum number of concurrent Goroutines based on the size of the control channel's cache, for example:

```go
Var limit = make(chan int, 3)

Func main() {
For _, w := range work {
Go func() {
Limit <- 1
w()
<-limit
}()
}
Select{}
}
```

The last sentence `select{}` is an empty pipe select statement that causes the `main` thread to block, preventing the program from exiting prematurely. There are also many methods such as `for{}`, `<-make(chan int)`, which can achieve similar effects. Because the `main` thread is blocked, it can be implemented by calling `os.Exit(0)` if the program needs to exit normally.

## 1.5.7 Unreliable synchronization

As we have analyzed before, the following code does not guarantee normal print results. The actual running effect is also a large probability that the result cannot be output normally.

```go
Func main() {
Go println("Hello, World")
}
```

Just contact Go, you may want to ensure normal output by adding a random sleep time:

```go
Func main() {
Go println("hello, world")
time.Sleep(time.Second)
}
```

Because the main thread sleeps for 1 second, this program has a high probability that it can output the result normally. Therefore, many people will feel that this program is no longer a problem. But this program is not stable and there is still the possibility of failure. Let us first assume that the program can stabilize the output. Because the start of the Go thread is non-blocking, the `main` thread explicitly sleeps for 1 second and the program ends. We can approximate that the program has executed more than 1 second. Now suppose that the internal sleep of the `println` function is longer than the sleep time of the `main` thread, which will lead to contradiction: since the background thread finishes printing before the `main` thread, the execution time is definitely less than the `main` thread execution time. of. Of course this is impossible.

The correctness of a rigorous concurrent program should not be dependent on unreliable factors such as CPU execution speed and sleep time. Strict concurrency should also be able to statically derive results: according to the order consistency within the thread, combined with the sortability of the Channel or `sync` synchronization events to derive, and finally complete the ordering of the partial order of each thread of each thread. If two events cannot be sorted according to this rule, then they are concurrent, that is, the execution order is not reliable.

The idea to solve the synchronization problem is the same: use explicit synchronization.