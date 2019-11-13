# 1.6 Common Concurrency Mode

The most appealing aspect of the Go language is its built-in concurrency support. The theory of the Go language concurrency system is the CSP (Communicating Sequential Process) proposed by C.A.R Hoare in 1978. CSP has a precise mathematical model and is actually applied to the T9000 general purpose computer that Hoare is involved in. From NewSqueak, Alef, Limbo to the current Go language, Rob Pike, who has more than 20 years of practical experience with CSP, is more concerned with the potential of applying CSP to a general-purpose programming language. There is only one core concept of CSP theory as the core of Go concurrent programming: synchronous communication. The topic of synchronous communication has already been covered in the previous section. In this section we will briefly introduce the concurrency patterns common in the Go language.

The first thing to be clear is a concept: concurrency is not parallel. Concurrency is more concerned with the design level of the program. Concurrent programs can be executed sequentially, and only on real multi-core CPUs can actually run at the same time. Parallelism is more concerned with the running level of the program. Parallelism is generally a simple and large number of repetitions. For example, there are a large number of parallel operations on image processing in the GPU. In order to better write concurrent programs, from the beginning of the design, Go language focuses on how to design a simple, safe and efficient abstract model at the programming language level, allowing programmers to focus on solving problems and combining solutions without being managed by threads and signals. Respond to these cumbersome operations.

In concurrent programming, proper access to shared resources requires precise control. In most current languages, this difficult problem is solved by a thread synchronization scheme such as locking, and the Go language has a different approach, which will share The value is passed through the Channel (in fact, multiple independently executing threads seldom actively share resources). At any given moment, it is best to have only one Goroutine to own the resource. Data competition has been eliminated from the design level. In order to promote this way of thinking, the Go language philosophizes its concurrent programming into a slogan:

> Do not communicate by sharing memory; instead, share memory by communicating.

> Do not communicate through shared memory, but share memory through communication.

This is a higher level of concurrent programming philosophy (passing values ​​through the pipeline is recommended by Go). While simple concurrency issues like reference counting are well implemented by atomic operations or mutex locks, controlling access through Channels allows you to write cleaner and more correct programs.

## 1.6.1 Concurrent version of Hello world

We first output "Hello world" in a new Goroutine, `main` waits for the background thread output to finish exiting, so a simple concurrent program is warmed up.

The core concept of concurrent programming is synchronous communication, but there are many ways to synchronize. We first implement synchronous communication with the familiar mutex `sync.Mutex`. According to the documentation, we can't directly unlock an unlocked `sync.Mutex`, which causes runtime exceptions. The following way does not guarantee normal work:

```go
Func main() {
Var mu sync.Mutex

Go func(){
fmt.Println("Hello, World")
mu.Lock()
}()

mu.Unlock()
}
```

Because `mu.Lock()` and `mu.Unlock()` are not in the same Goroutine, they do not satisfy the sequential consistency memory model. At the same time, they have no other synchronization events to refer to. These two events can not be ordered, that is, they can be concurrent. Because it may be a concurrent event, `mu.Unlock()` in the `main` function is likely to occur first, and the `mu` mutex is still in an unlocked state at this time, which causes a runtime exception. .

Here is the repaired code:

```go
Func main() {
Var mu sync.Mutex

mu.Lock()
Go func(){
fmt.Println("Hello, World")
mu.Unlock()
}()

mu.Lock()
}
```

The way to fix it is to execute `mu.Lock()` twice in the thread where the `main` function is located. When the second lock is applied, it will be blocked because the lock is already occupied (not a recursive lock), and the `main` function is blocked. The state-driven background thread continues to execute forward. When the background thread is unlocked when it executes to `mu.Unlock()`, the print job is completed. Unlocking will cause the second `mu.Lock()` blocking state in the `main` function to be canceled. There is no other synchronization event reference with the main thread, and the events they exit will be concurrent: when the `main` function exits causing the program to exit, the background thread may have exited or may not exit. Although it is not possible to determine when two threads will exit, the print job can be done correctly.

Synchronization with `sync.Mutex` is a relatively low-level approach. We are now using a non-cached pipeline to achieve synchronization:

```go
Func main() {
Done := make(chan int)

Go func(){
fmt.Println("Hello, World")
<-done
}()

Done <- 1
}
```

According to the Go language memory model specification, for reception from an unbuffered channel, before the transmission to the Channel is completed. Therefore, after the background thread `<-done` receives the operation, the `done <- 1` send operation of the `main` thread is completed (thus exiting the main, exiting the program), and the print job is now complete.

Although the above code can be properly synchronized, it is too sensitive to the size of the pipeline's cache: if the pipeline has a cache, there is no guarantee that the background thread will print properly before the main exit. A better approach is to swap the send and receive directions of the pipeline to avoid synchronization events being affected by the size of the pipeline cache:

```go
Func main() {
Done := make(chan int, 1) // pipe with cache

Go func(){
fmt.Println("Hello, World")
Done <- 1
}()

<-done
}
```

For a buffered Channel, the Kth receive completion operation for the Channel occurs before the K+Cth transmit operation is completed, where C is the buffer size of the Channel. Although the pipeline is cached, the completion of the `main` thread is the time when the background thread sends the start but has not yet completed, and the print job is completed.

Based on the pipeline with cache, we can easily extend the print thread to N. The following example is to open 10 background threads to print separately:

```go
Func main() {
Done := make(chan int, 10) // with 10 caches

// Open N spool threads
For i := 0; i < cap(done); i++ {
Go func(){
fmt.Println("Hello, World")
Done <- 1
}()
}

/ / Wait for N background threads to complete
For i := 0; i < cap(done); i++ {
<-done
}
}
```

One simple way to do this is to wait for N threads to complete before proceeding to the next synchronization, using `sync.WaitGroup` to wait for a set of events:

```go
Func main() {
Var wg sync.WaitGroup

// Open N spool threads
For i := 0; i < 10; i++ {
wg.Add(1)

Go func() {
fmt.Println("Hello, World")
wg.Done()
}()
}

/ / Wait for N background threads to complete
wg.Wait()
}
```

The `wg.Add(1)` is used to increase the number of wait events, and must be executed before the background thread starts (if it is executed in the background thread, it is not guaranteed to be executed normally). When the background thread finishes printing, call `wg.Done()` to complete an event. The `wg.Wait()` of the `main` function waits for all events to complete.

## 1.6.2 Producer Consumer Model

The most common example of concurrent programming is the Producer Consumer model, which increases the overall processing speed of the program by balancing the working power of the production and consumption threads. Simply put, the producer produces some data and then puts it in the results queue, while the consumer takes the data from the results queue. This makes production and consumption become two processes of asynchronous. When there is no data in the results queue, the consumer enters the wait for hunger; when the data in the results queue is full, the producer faces the problem of the laid-off of the CPU being deprived due to product squeeze.

Go language implementation producer consumer concurrency is very simple:

```go
// Producer: Generate a sequence of factor multiples of factor
Func Producer(factor int, out chan<- int) {
For i := 0; ; i++ {
Out <- i*factor
}
}

// consumer
Func Consumer(in <-chan int) {
For v := range in {
fmt.Println(v)
}
}
Func main() {
Ch := make(chan int, 64) // result queue

Go Producer(3, ch) // Generate a sequence of multiples of 3
Go Producer(5, ch) // Generate a sequence of multiples of 5
Go Consumer(ch) // consumption generated queue

// Exit after running for a certain period of time
time.Sleep(5 * time.Second)
}
```

We have opened two `Producer` production lines for generating a sequence of multiples of 3 and 5. Then open a `Consumer` consumer thread and print the result. We let producers and consumers work for a certain amount of time by sleeping in the `main` function for a certain amount of time. As mentioned in the previous section, this sleep mode does not guarantee a stable output.

We can let the `main` function save the blocking state without exiting, only when the user enters `Ctrl-C` to actually exit the program:

```go
Func main() {
Ch := make(chan int, 64) // result queue

Go Producer(3, ch) // Generate a sequence of multiples of 3
Go Producer(5, ch) // Generate a sequence of multiples of 5
Go Consumer(ch) // consumption generated queue

// Ctrl+C to exit
Sig := make(chan os.Signal, 1)
signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
fmt.Printf("quit (%v)\n", <-sig)
}
```

There are 2 producers in our example, and there are no synchronization events between the two producers. They are concurrent. Therefore, the order of the output sequence of the consumer output is uncertain, and this is no problem, and producers and consumers can still work together.

## 1.6.3 Publishing a subscription model

The publish-and-subscribe model is usually abbreviated as a pub/sub model. In this model, the message producer becomes the publisher, and the message consumer becomes the subscriber, and the producer and consumer are M:N relationships. In the traditional producer and consumer model, messages are sent to a queue, and the publish-subscription model publishes messages to a topic.

To do this, we built a publish subscription model support package called `pubsub`:

```go
// Package pubsub implements a simple multi-topic pub-sub library.
Package pubsub

Import (
"sync"
"time"
)

Type (
Subscriber chan interface{} // subscriber is a pipe
topicFunc func(v interface{}) bool // The theme is a filter
)

// publisher object
Type Publisher struct {
m sync.RWMutex // read-write lock
Buffer int // the cache size of the subscription queue
Timeout time.Duration// Publish timeout
Subscribers map[subscriber]topicFunc // subscriber information
}

// Build a publisher object, you can set the publish timeout and the length of the cache queue
Func NewPublisher(publishTimeout time.Duration, buffer int) *Publisher {
Return &Publisher{
Buffer: buffer,
Timeout: publishTimeout,
Subscribers: make(map[subscriber]topicFunc),
}
}

// Add a new subscriber, subscribe to all topics
Func (p *Publisher) Subscribe() chan interface{} {
Return p.SubscribeTopic(nil)
}

// Add a new subscriber, subscribe to the filter filtered subject
Func (p *Publisher) SubscribeTopic(topic topicFunc) chan interface{} {
Ch := make(chan interface{}, p.buffer)
p.m.Lock()
P.subscribers[ch] = topic
p.m.Unlock()
Return ch
}

// quit the subscription
Func (p *Publisher) Evict(sub chan interface{}) {
p.m.Lock()
Defer p.m.Unlock()

Delete(p.subscribers, sub)
Close(sub)
}

// Post a topic
Func (p *Publisher) Publish(v interface{}) {
p.m.RLock()
Defer p.m.RUnlock()

Var wg sync.WaitGroup
For sub, topic := range p.subscribers {
wg.Add(1)
Go p.sendTopic(sub, topic, v, &wg)
}
wg.Wait()
}

// Close the publisher object and close all subscriber pipes.
Func (p *Publisher) Close() {
p.m.Lock()
Defer p.m.Unlock()

For sub := range p.subscribers {
Delete(p.subscribers, sub)
Close(sub)
}
}

/ / Send a topic, can tolerate a certain timeout
Func (p *Publisher) sendTopic(
Sub subscriber, topic topicFunc, v interface{}, wg *sync.WaitGroup,
) {
Defer wg.Done()
If topic != nil && !topic(v) {
Return
}

Select {
Case sub <- v:
Case <-time.After(p.timeout):
}
}
```

In the following example, two subscribers subscribed to all topics and topics with "golang":

```go
Import "path/to/pubsub"

Func main() {
p := pubsub.NewPublisher(100*time.Millisecond, 10)
Defer p.Close()

All := p.Subscribe()
Golang := p.SubscribeTopic(func(v interface{}) bool {
If s, ok := v.(string); ok {
Return strings.Contains(s, "golang")
}
Return false
})

p.Publish("hello, world!")
p.Publish("hello, golang!")

Go func() {
For msg := range all {
fmt.Println("all:", msg)
}
} ()

Go func() {
For msg := range golang {
fmt.Println("golang:", msg)
}
} ()

// Exit after running for a certain period of time
time.Sleep(3 * time.Second)
}
```

In the publish subscription model, each message is delivered to multiple subscribers. Publishers usually don't know or care which subscriber is receiving the subject message. Subscribers and publishers can be dynamically added at runtime, a loosely coupled relationship that allows the complexity of the system to grow over time. In real life, applications like weather forecasts can apply this concurrency pattern.

## 1.6.4 Controlling the number of concurrent

Many users tend to write the most concurrent programs after adapting to the powerful concurrency features of the Go language, as this seems to provide maximum performance. In reality, we are in a hurry, but sometimes we need to slow down and enjoy life. The same procedure is used: sometimes we need to control the degree of concurrency appropriately, because it can not only give up other applications/tasks/ Reserve a certain amount of CPU resources, you can also reduce the power consumption to ease the pressure on the battery.

In the godoc program implementation of Go language, there is a `vfs` package corresponding to the virtual file system. There is a sub-package of `gatefs` under the `vfs` package. The purpose of the `gatefs` sub-package is to control access. The maximum number of concurrency of the virtual file system. The application of the `gatefs` package is simple:

```go
Import (
"golang.org/x/tools/godoc/vfs"
"golang.org/x/tools/godoc/vfs/gatefs"
)

Func main() {
Fs := gatefs.New(vfs.OS("/path"), make(chan bool, 8))
// ...
}
```

Where `vfs.OS("/path")` constructs a virtual filesystem based on the local filesystem, and then `gatefs.New` constructs a concurrently controlled virtual filesystem based on the existing virtual filesystem. The principle of concurrency control has been discussed in the previous section, which is to achieve maximum concurrent blocking by sending and receiving rules with cache pipes:

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


However, `gatefs` does an abstract type `gate` for this, adding the `enter` and `leave` methods to the entry and exit of the concurrent code, respectively. When the number of concurrency limits is exceeded, the `enter` method blocks until the number of concurrency drops.

```go
Type gate chan bool

Func (g gate) enter() { g <- true }
Func (g gate) leave() { <-g }
```

The new virtual filesystem wrapped by `gatefs` adds the `enter` and `leave` calls to methods that need to control concurrency:


```go
Type gatefs struct {
Fs vfs.FileSystem
Gate
}

Func (fs gatefs) Lstat(p string) (os.FileInfo, error) {
Fs.enter()
Defer fs.leave()
Return fs.fs.Lstat(p)
}
```

We can not only control the maximum number of concurrency, but also determine the concurrency rate of the program running by using the usage and maximum capacity ratio of the cached channel. When the pipeline is empty, it can be considered as idle state. When the pipeline is full, the task is busy. This is of reference value for the operation of some low-level tasks in the background.


## 1.6.5 Winner is king

There are many motivations for concurrent programming: concurrent programming can simplify problems. For example, one type of problem is simpler for a processing thread. Concurrent programming can also improve performance. Opening two threads on a multi-core CPU is generally faster than opening one thread. some. In fact, in terms of improving performance, the program is not simply running fast, indicating that the user experience is good; in many cases, the program can respond to user requests quickly, which is the most important. When there is no user request to process, it is appropriate to handle some low priority. Level background tasks.

Suppose we want to quickly search for "golang" related topics, we may open multiple search engines such as Bing, Google or Baidu. When a search returns results first, you can close other search pages. Because of the influence of the network environment and search engine algorithms, some search engines may return search results quickly, and some search engines may wait until their company fails and does not complete the search. We can use a similar strategy to write this program:

```go
Func main() {
Ch := make(chan string, 32)

Go func() {
Ch <- searchByBing("golang")
}()
Go func() {
Ch <- searchByGoogle("golang")
}()
Go func() {
Ch <- searchByBaidu("golang")
}()

fmt.Println(<-ch)
}
```

First, we created a pipeline with a cache that is large enough to ensure that there is no unnecessary blocking due to the size of the cache. Then we opened multiple background threads and submitted search requests to different search engines. When any search engine has the first result, it will immediately send the result to the pipeline (because the pipeline has enough cache, this process will not block). But in the end we only take the first result from the pipeline, which is the result of the first return.

By properly opening some redundant threads, try to solve the same problem in different ways, and finally improve the corresponding performance of the program by winning the winner.


## 1.6.6 Prime screen

In the "Revolution of the Hello World" section, we demonstrate the implementation of the concurrent version of the prime number in order to demonstrate the concurrency of Newsqueak. The concurrent version of Prime Screen is a classic concurrency example that gives us a deeper understanding of the concurrency features of Go. The principle of "prime sieve" is as follows:

![](../images/ch1-13-prime-sieve.png)

*Figure 1-13 Prime Screen*


We need Mr. to form the original `2, 3, 4, ...` natural number sequence (without the beginning 0, 1):

```go
// Return the pipeline that generated the sequence of natural numbers: 2, 3, 4, ...
Func GenerateNatural() chan int {
Ch := make(chan int)
Go func() {
For i := 2; ; i++ {
Ch <- i
}
}()
Return ch
}
```

The `GenerateNatural` function internally starts a Goroutine production sequence and returns the corresponding pipeline.

Then construct a sieve for each prime number: propose a number that is a multiple of the prime number in the input sequence, and return a new sequence, which is a new pipe.

```go
// Pipeline filter: Delete the number that can be divisible by the prime number
Func PrimeFilter(in <-chan int, prime int) chan int {
Out := make(chan int)
Go func() {
For {
If i := <-in; i%prime != 0 {
Out <- i
}
}
}()
Return out
}
```

The `PrimeFilter` function also internally launches a Goroutine production sequence, returning the pipeline corresponding to the filtered sequence.

Now we can drive this concurrent prime number filter in the `main` function:

```go
Func main() {
Ch := GenerateNatural() // NaturalNumber sequence: 2, 3, 4, ...
For i := 0; i < 100; i++ {
Prime := <-ch // new prime number
fmt.Printf("%v: %v\n", i+1, prime)
Ch = PrimeFilter(ch, prime) // Filter based on new primes
}
}
```

We first call `GenerateNatural()` to generate the most primitive sequence of natural numbers starting with 2. Then start a cycle of 100 iterations, hoping to generate 100 prime numbers. At the beginning of each loop iteration, the first number in the pipeline must be a prime number. We read and print the prime number first. Then based on the remaining series in the pipeline, and filtering the subsequent prime numbers with the currently extracted primes as a sieve. The pipes corresponding to different prime screens are connected in series.

Prime screens show an elegant concurrent program structure. However, because the task granularity of each concurrent body processing is too small, the overall performance of the program is not ideal. For fine-grained concurrent programs, the inherent message passing in the CSP model is too expensive (multithreaded concurrent models also face the cost of thread startup).

## 1.6.7 Concurrent security exit

Sometimes we need to tell goroutine to stop what it is doing, especially when it is working in the wrong direction. The Go language does not provide a way to terminate Goroutine directly, as this will cause the shared variable between goroutines to be in an undefined state. But what if we want to quit two or more Goroutines?

Different Goroutines in the Go language rely on pipes for communication and synchronization. To handle the sending or receiving of multiple pipes at the same time, we need to use the `select` keyword (this keyword behaves similarly to the `select` function in network programming). When `select` has multiple branches, it will randomly select an available pipe branch. If there is no pipe branch available, select the `default` branch, otherwise the blocking state will be saved.

The timeout judgment of the pipeline based on `select`:

```go
Select {
Case v := <-in:
fmt.Println(v)
Case <-time.After(time.Second):
Return // timeout
}
```

Non-blocking pipe send or receive operations through the `default` branch of `select`:

```go
Select {
Case v := <-in:
fmt.Println(v)
Default:
	// no data
}
```

Prevent `main` function from exiting by `select`:

```go
Func main() {
// do some thins
Select{}
}
```

When multiple pipes are operational, `select` will randomly select a pipe. Based on this feature, we can use `select` to implement a program that generates a sequence of random numbers:

```go
Func main() {
Ch := make(chan int)
Go func() {
For {
Select {
Case ch <- 0:
Case ch <- 1:
}
}
}()

For v := range ch {
fmt.Println(v)
}
}
```

We can easily implement a Goroutine exit control through the `select` and `default` branches:

```go
Func worker(cannel chan bool) {
For {
Select {
Default:
fmt.Println("hello")
			// normal work
Case <-cannel:
			// drop out
}
}
}

Func main() {
Cannel := make(chan bool)
Go worker(cannel)

time.Sleep(time.Second)
Cannel <- true
}
```

However, the sending and receiving operations of the pipeline are one-to-one. If you want to stop multiple Goroutines, you may need to create the same number of pipes. This is too expensive. In fact, we can close the pipeline through `close` to achieve the effect of the broadcast, all operations received from the closed pipeline will receive a zero value and an optional failure flag.

```go
Func worker(cannel chan bool) {
For {
Select {
Default:
fmt.Println("hello")
			// normal work
Case <-cannel:
			// drop out
}
}
}

Func main() {
Cancel := make(chan bool)

For i := 0; i < 10; i++ {
Go worker(cancel)
}

time.Sleep(time.Second)
Close(cancel)
}
```

We use `close` to close the `cancel` pipeline to broadcast instructions to multiple Goroutine exits. However, this program is still not robust enough: when each Goroutine receives an exit command, it will generally perform some cleanup work, but the cleanup work of the exit is not guaranteed to be completed, because the `main` thread does not wait for the work Goroutine to exit the work. Mechanisms. We can improve with `sync.WaitGroup`:

```go
Func worker(wg *sync.WaitGroup, cannel chan bool) {
Defer wg.Done()

For {
Select {
Default:
fmt.Println("hello")
Case <-cannel:
Return
}
}
}

Func main() {
Cancel := make(chan bool)

Var wg sync.WaitGroup
For i := 0; i < 10; i++ {
wg.Add(1)
Go worker(&wg, cancel)
}

time.Sleep(time.Second)
Close(cancel)
wg.Wait()
}
```

Now the creation, operation, pause, and exit of each worker's concurrency are under the security control of the `main` function.


## 1.6.8 context package

At the time of the release of Go1.7, the standard library added a `context` package to simplify the operation of data, timeouts, and exits between multiple Goroutines and request domains for processing a single request. Introduction. We can use the `context` package to re-implement the previous thread-safe exit or timeout control:

```go
Func worker(ctx context.Context, wg *sync.WaitGroup) error {
Defer wg.Done()

For {
Select {
Default:
fmt.Println("hello")
Case <-ctx.Done():
Return ctx.Err()
}
}
}

Func main() {
Ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

Var wg sync.WaitGroup
For i := 0; i < 10; i++ {
wg.Add(1)
Go worker(ctx, &wg)
}

time.Sleep(time.Second)
Cancel()

wg.Wait()
}
```

When the concurrent body times out or `main` actively stops the worker Goroutine, each worker can safely quit.

The Go language is automatically reclaimed with memory, so memory is generally not leaking. In the previous prime-screen example, the `GenerateNatural` and `PrimeFilter` functions internally start a new Goroutine, and the back-end Goroutine is at risk of leaking when the `main` function no longer uses the pipeline. We can avoid this problem with the `context` package. Here is the improved prime screen implementation:

```go
// Return the pipeline that generated the sequence of natural numbers: 2, 3, 4, ...
Func GenerateNatural(ctx context.Context) chan int {
Ch := make(chan int)
Go func() {
For i := 2; ; i++ {
Select {
Case <- ctx.Done():
Return
Case ch <- i:
}
}
}()
Return ch
}

// Pipeline filter: Delete the number that can be divisible by the prime number
Func PrimeFilter(ctx context.Context, in <-chan int, prime int) chan int {
Out := make(chan int)
Go func() {
For {
If i := <-in; i%prime != 0 {
Select {
Case <- ctx.Done():
Return
Case out <- i:
}
}
}
}()
Return out
}

Func main() {
/ / Control the background Goroutine state through the Context
Ctx, cancel := context.WithCancel(context.Background())

Ch := GenerateNatural(ctx) // Sequence of natural numbers: 2, 3, 4, ...
For i := 0; i < 100; i++ {
Prime := <-ch // new prime number
fmt.Printf("%v: %v\n", i+1, prime)
Ch = PrimeFilter(ctx, ch, prime) // Filter based on new primes
}

Cancel()
}
```

Before the main function finishes working, the background Goroutine is notified to exit by calling `cancel()`, thus avoiding the Goroutine leak.

Concurrency is a very big topic, and here we are just showing a few examples of very basic concurrent programming. The official documentation also has a lot of discussion about concurrent programming, and there are books in the country that specifically discuss Go language concurrent programming. Readers can refer to relevant documents according to their needs.