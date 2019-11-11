# 附录A：Go语言常见坑

这里列举的Go语言常见坑都是符合Go语言语法的，可以正常的编译，但是可能是运行结果错误，或者是有资源泄漏的风险。

## 可变参数是空接口类型

当参数的可变参数是空接口类型时，传入空接口的切片时需要注意参数展开的问题。

```go
func main() {
	var a = []interface{}{1, 2, 3}

	fmt.Println(a)
	fmt.Println(a...)
}
```

不管是否展开，编译器都无法发现错误，但是输出是不同的：

```
[1 2 3]
1 2 3
```

## 数组是值传递

在函数调用参数中，数组是值传递，无法通过修改数组类型的参数返回结果。

```go
func main() {
	x := [3]int{1, 2, 3}

	func(arr [3]int) {
		arr[0] = 7
		fmt.Println(arr)
	}(x)

	fmt.Println(x)
}
```

必要时需要使用切片。

## map遍历是顺序不固定

map是一种hash表实现，每次遍历的顺序都可能不一样。

```go
func main() {
	m := map[string]string{
		"1": "1",
		"2": "2",
		"3": "3",
	}

	for k, v := range m {
		println(k, v)
	}
}
```

## 返回值被屏蔽

在局部作用域中，命名的返回值内同名的局部变量屏蔽：

```go
func Foo() (err error) {
	if err := Bar(); err != nil {
		return
	}
	return
}
```

## recover必须在defer函数中运行

recover捕获的是祖父级调用时的异常，直接调用时无效：

```go
func main() {
	recover()
	panic(1)
}
```

直接defer调用也是无效：

```go
func main() {
	defer recover()
	panic(1)
}
```

defer调用时多层嵌套依然无效：

```go
func main() {
	defer func() {
		func() { recover() }()
	}()
	panic(1)
}
```

必须在defer函数中直接调用才有效：

```go
func main() {
	defer func() {
		recover()
	}()
	panic(1)
}
```

## main函数提前退出

后台Goroutine无法保证完成任务。

```go
func main() {
	go println("hello")
}
```

## 通过Sleep来回避并发中的问题

休眠并不能保证输出完整的字符串：

```go
func main() {
	go println("hello")
	time.Sleep(time.Second)
}
```

类似的还有通过插入调度语句：

```go
func main() {
	go println("hello")
	runtime.Gosched()
}
```

## 独占CPU导致其它Goroutine饿死

Goroutine是协作式抢占调度，Goroutine本身不会主动放弃CPU：

```go
func main() {
	runtime.GOMAXPROCS(1)

	go func() {
		for i := 0; i < 10; i++ {
			fmt.Println(i)
		}
	}()

	for {} // 占用CPU
}
```

解决的方法是在for循环加入runtime.Gosched()调度函数：

```go
func main() {
	runtime.GOMAXPROCS(1)

	go func() {
		for i := 0; i < 10; i++ {
			fmt.Println(i)
		}
	}()

	for {
		runtime.Gosched()
	}
}
```

或者是通过阻塞的方式避免CPU占用：

```go
func main() {
	runtime.GOMAXPROCS(1)

	go func() {
		for i := 0; i < 10; i++ {
			fmt.Println(i)
		}
		os.Exit(0)
	}()

	select{}
}
```

## 不同Goroutine之间不满足顺序一致性内存模型

因为在不同的Goroutine，main函数中无法保证能打印出`hello, world`:

```go
var msg string
var done bool

func setup() {
	msg = "hello, world"
	done = true
}

func main() {
	go setup()
	for !done {
	}
	println(msg)
}
```

解决的办法是用显式同步：

```go
var msg string
var done = make(chan bool)

func setup() {
	msg = "hello, world"
	done <- true
}

func main() {
	go setup()
	<-done
	println(msg)
}
```
msg的写入是在channel发送之前，所以能保证打印`hello, world`

## 闭包错误引用同一个变量

```go
func main() {
	for i := 0; i < 5; i++ {
		defer func() {
			println(i)
		}()
	}
}
```

改进的方法是在每轮迭代中生成一个局部变量：

```go
func main() {
	for i := 0; i < 5; i++ {
		i := i
		defer func() {
			println(i)
		}()
	}
}
```

或者是通过函数参数传入：

```go
func main() {
	for i := 0; i < 5; i++ {
		defer func(i int) {
			println(i)
		}(i)
	}
}
```

## 在循环内部执行defer语句

defer在函数退出时才能执行，在for执行defer会导致资源延迟释放：

```go
func main() {
	for i := 0; i < 5; i++ {
		f, err := os.Open("/path/to/file")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
	}
}
```

解决的方法可以在for中构造一个局部函数，在局部函数内部执行defer：

```go
func main() {
	for i := 0; i < 5; i++ {
		func() {
			f, err := os.Open("/path/to/file")
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
		}()
	}
}
```

## 切片会导致整个底层数组被锁定

切片会导致整个底层数组被锁定，底层数组无法释放内存。如果底层数组较大会对内存产生很大的压力。

```go
func main() {
	headerMap := make(map[string][]byte)

	for i := 0; i < 5; i++ {
		name := "/path/to/file"
		data, err := ioutil.ReadFile(name)
		if err != nil {
			log.Fatal(err)
		}
		headerMap[name] = data[:1]
	}

	// do some thing
}
```

解决的方法是将结果克隆一份，这样可以释放底层的数组：

```go
func main() {
	headerMap := make(map[string][]byte)

	for i := 0; i < 5; i++ {
		name := "/path/to/file"
		data, err := ioutil.ReadFile(name)
		if err != nil {
			log.Fatal(err)
		}
		headerMap[name] = append([]byte{}, data[:1]...)
	}

	// do some thing
}
```

## 空指针和空接口不等价

比如返回了一个错误指针，但是并不是空的error接口：

```go
func returnsError() error {
	var p *MyError = nil
	if bad() {
		p = ErrBad
	}
	return p // Will always return a non-nil error.
}
```

## 内存地址会变化

Go语言中对象的地址可能发生变化，因此指针不能从其它非指针类型的值生成：

```go
func main() {
	var x int = 42
	var p uintptr = uintptr(unsafe.Pointer(&x))

	runtime.GC()
	var px *int = (*int)(unsafe.Pointer(p))
	println(*px)
}
```

当内存发送变化的时候，相关的指针会同步更新，但是非指针类型的uintptr不会做同步更新。

同理CGO中也不能保存Go对象地址。

## Goroutine泄露

Go语言是带内存自动回收的特性，因此内存一般不会泄漏。但是Goroutine确存在泄漏的情况，同时泄漏的Goroutine引用的内存同样无法被回收。

```go
func main() {
	ch := func() <-chan int {
		ch := make(chan int)
		go func() {
			for i := 0; ; i++ {
				ch <- i
			}
		} ()
		return ch
	}()

	for v := range ch {
		fmt.Println(v)
		if v == 5 {
			break
		}
	}
}
```

上面的程序中后台Goroutine向管道输入自然数序列，main函数中输出序列。但是当break跳出for循环的时候，后台Goroutine就处于无法被回收的状态了。

我们可以通过context包来避免这个问题：


```go
func main() {
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
			cancel()
			break
		}
	}
}
```

当main函数在break跳出循环时，通过调用`cancel()`来通知后台Goroutine退出，这样就避免了Goroutine的泄漏。
