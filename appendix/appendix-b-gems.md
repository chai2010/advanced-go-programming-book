# 附录B：有趣的代码片段

这里收集一些比较有意思的Go程序片段。

## 自重写程序

UNIX/Go语言之父 Ken Thompson 在1983年的图灵奖演讲 Reflections on Trusting Trust 就给出了一个C语言的自重写程序。

最短的C语言自重写程序是 Vlad Taeerov 和 Rashit Fakhreyev 的版本：

```c
main(a){printf(a="main(a){printf(a=%c%s%c,34,a,34);}",34,a,34);}
```

下面的Go语言版本自重写程序是 [rsc](https://research.swtch.com/zip) 提供的：

```go
/* Go quine */
package main

import "fmt"

func main() {
	fmt.Printf("%s%c%s%c\n", q, 0x60, q, 0x60)
}

var q = `/* Go quine */
package main

import "fmt"

func main() {
	fmt.Printf("%s%c%s%c\n", q, 0x60, q, 0x60)
}

var q = `
```

在 golang-nuts 中还有很多版本：

```go
package main;func main(){c:="package main;func main(){c:=%q;print(c,c)}";print(c,c)}
```

```go
package main;func main(){print(c+"\x60"+c+"\x60")};var c=`package main;func main(){print(c+"\x60"+c+"\x60")};var c=`
```

如果有更短的版本欢迎告诉我们。

## 三元表达式

```go
func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

a, b := 2, 3
max := If(a > b, a, b).(int)
println(max)
```

## 禁止 main 函数退出的方法

```go
func main() {
	defer func() { for {} }()
}

func main() {
	defer func() { select {} }()
}

func main() {
	defer func() { <-make(chan bool) }()
}
```

## 基于管道的随机数生成器

随机数的一个特点是不好预测。如果一个随机数的输出是可以简单预测的，那么一般会称为伪随机数。

```go
func main() {
	for i := range random(100) {
		fmt.Println(i)
	}
}

func random(n int) <-chan int {
	c := make(chan int)
	go func() {
		defer close(c)
		for i := 0; i < n; i++ {
			select {
			case c <- 0:
			case c <- 1:
			}
		}
	}()
	return c
}
```

基于select语言特性构造的随机数生成器。

## Assert测试断言

```go
type testing_TBHelper interface {
	Helper()
}

func Assert(tb testing.TB, condition bool, args ...interface{}) {
	if x, ok := tb.(testing_TBHelper); ok {
		x.Helper() // Go1.9+
	}
	if !condition {
		if msg := fmt.Sprint(args...); msg != "" {
			tb.Fatalf("Assert failed, %s", msg)
		} else {
			tb.Fatalf("Assert failed")
		}
	}
}

func Assertf(tb testing.TB, condition bool, format string, a ...interface{}) {
	if x, ok := tb.(testing_TBHelper); ok {
		x.Helper() // Go1.9+
	}
	if !condition {
		if msg := fmt.Sprintf(format, a...); msg != "" {
			tb.Fatalf("Assertf failed, %s", msg)
		} else {
			tb.Fatalf("Assertf failed")
		}
	}
}

func AssertFunc(tb testing.TB, fn func() error) {
	if x, ok := tb.(testing_TBHelper); ok {
		x.Helper() // Go1.9+
	}
	if err := fn(); err != nil {
		tb.Fatalf("AssertFunc failed, %v", err)
	}
}
```
