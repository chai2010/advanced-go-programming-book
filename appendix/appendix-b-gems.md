# Appendix B: Interesting code snippets

Here are some interesting Go program snippets.

## Self-rewriting program

The father of the UNIX/Go language Ken Thompson gave a self-rewriting program for the C language in the 1983 Turing Awards Reflections on Trusting Trust.

The shortest C self-rewriting program is the version of Vlad Taeerov and Rashit Fakhreyev:

```c
Main(a){printf(a="main(a){printf(a=%c%s%c,34,a,34);}",34,a,34);}
```

The following Go language version self-rewriting program is provided by [rsc](https://research.swtch.com/zip):

```go
/* Go quine */
Package main

Import "fmt"

Func main() {
fmt.Printf("%s%c%s%c\n", q, 0x60, q, 0x60)
}

Var q = `/* Go quine */
Package main

Import "fmt"

Func main() {
fmt.Printf("%s%c%s%c\n", q, 0x60, q, 0x60)
}

Var q = `
```

There are many more versions in golang-nuts:

```go
Package main;func main(){c:="package main;func main(){c:=%q;print(c,c)}";print(c,c)}
```

```go
Package main;func main(){print(c+"\x60"+c+"\x60")};var c=`package main;func main(){print(c+"\x60"+c+"\x60") };var c=`
```

if there is a shorter version, please let us know.

## Ternary expression

```go
func if(condition bool, trueVal, falseVal interface{}) interface{} {
if condition {
return trueVal
}
return falseVal
}

a, b := 2, 3
Max := if(a > b, a, b).(int)
Println(max)
```

## Prohibit the method of exiting the main function

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

## Pipe-based random number generator

One feature of random numbers is poor prediction. if the output of a random number is simple predictable, it is generally called a pseudo-random number.

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

A random number generator constructed based on the characteristics of the select language.

## Assert Test Assertion

```go
Type testing_TBHelper interface {
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