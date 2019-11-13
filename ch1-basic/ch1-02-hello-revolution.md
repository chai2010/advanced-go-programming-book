# 1.2 Hello, World's revolution

In the Genesis chapter, we briefly introduced the evolutionary gene family of Go language, and highlighted the unique concurrent programming genes from Bell Labs. Finally, the Go version of the "Hello, World" program was introduced. In fact, the "Hello, World" program is the best example of showing various language features, and is a window to the language. In this section, we will briefly review the timeline of the evolution of each programming language, and briefly review how the "Hello, World" program evolved into the current Go language form and finally completed its revolutionary mission.

![](../images/ch1-4-go-history.png)

*Figure 1-4 Go language concurrent evolution history*

## 1.2.1 B Language - Ken Thompson, 1972

The first is the B language, which is a general-purpose programming language developed by Ken Thompson of Bell Labs, the father of the Go language, designed to aid in the development of UNIX systems. However, the lack of a flexible type system in the B language makes it difficult to use. Later, Ken Thompson's colleague Dennis Ritchie developed the C language based on the B language. The C language provides a rich variety, which greatly increases the language's expressive ability. It is still one of the most commonly used programming languages ​​in the world so far. Since the B language has been replaced by it, it has only existed in various documents and has become history.

The B language version of "Hello World" currently seen is generally considered to be from the B language introduction tutorial written by Brian W. Kernighan (the first committer name in the Go core code base is Brian W. Kernighan), the program as follows:

```c
Main() {
Extrn a, b, c;
Putchar(a); putchar(b); putchar(c);
Putchar('!*n');
}
a 'hell';
b 'o, w';
c 'orld';
```

Due to the lack of flexible data types in the B language, the contents to be output can only be defined by the `a/b/c` global variables, and the length of each variable must be aligned to 4 bytes (there is a feeling of writing assembly language). ). Then output the character by calling the `putchar` function multiple times. The last `'!*n'` means to output a newline.

In general, the B language is simple and the function is relatively simple.

## 1.2.2 C - Dennis Ritchie, 1974 ~ 1989

The C language was developed by Dennis Ritchie on the basis of the B language, which adds rich data types and ultimately achieves the great goal of rewriting UNIX with it. C language can be said to be the most important software cornerstone of the modern IT industry. At present, almost all mainstream operating systems are developed by C language, and many basic system software is also developed by C language. The programming languages ​​of the C-series have dominated for decades and have remained vibrant for more than half a century.

In the C language introductory tutorial written by Brian W. Kernighan around 1974, the first C language version of the "Hello World" program appeared. This provided the convention for the first program with "Hello World" for most of the later programming language tutorials. The first C language version of the "Hello World" program is as follows:

```c
Main()
{
Printf("hello, world");
}
```

There are a few points to note about this program: first, the `main` function returns the `int` type by default because it does not explicitly return a value type. Secondly, the `printf` function can be used by default without importing a function declaration; finally `main` There is no explicit return statement, but the default returns a value of 0. When this program appeared, the C language was far from standardized. What we saw was the C language syntax of the ancient times: the function does not need to write the return value, the function parameters can be ignored, and the printf does not need to include the header file.

This example also appeared in the first edition of the C Programming Language published in 1978 by Brian W. Kernighan and Dennis M. Ritchie (K&R). A newline output has been added to the end of "Hello World" in the book:

```c
Main()
{
Printf("hello, world\n");
}
```

This example adds a newline at the end of the string. The ``n` line feed in C language looks a bit cleaner than the B's ``!*n'` line feed.

In 1988, 10 years after the introduction of K&R's tutorial, the second edition of C Programming Language was finally published. At this point, the standardization of the ANSI C language has been initially completed, but the official version of the document has not yet been released. However, the "Hello World" program in the book adds a `#include <stdio.h>` header file containing statement for the declaration containing the `printf` function according to the new specification (in the new C89 standard, only for `printf In terms of functions, you can still use them without declaring functions.

```c
#include <stdio.h>

Main()
{
Printf("hello, world\n");
}
```

Then in 1989, the first international standard for ANSI C was released, commonly referred to as C89. The C89 is the most popular C language standard and is still widely used. The second edition of "C Programming Language" also reprinted the new version, and for the newly released C89 specification, added a `void` input parameter description to the parameter of the `main` function, indicating that there is no input parameter.

```c
#include <stdio.h>

Main(void)
{
Printf("hello, world\n");
}
```

At this point, the evolution of the C language itself is basically completed. The latter C92/C99/C11 is only perfect for some language details. Because of various historical factors, C89 is still the most widely used standard.


## 1.2.3 Newsqueak - Rob Pike, 1989

Newsqueak is the second generation of the mouse language invented by Rob Pike, the battlefield he used to practice the CSP concurrent programming model. Newsqueak is the meaning of the new squeak language, where squeak is the sound of a rat, or it can be seen as a sound similar to a mouse click. Squeak is a programming language that provides mouse and keyboard event handling. The Squeak language pipeline is statically created. The improved version of the Newsqueak language provides syntax similar to C language statements and expressions and derivation syntax similar to the Pascal language. Newsqueak is a purely functional language with automatic garbage collection that is once again managed for keyboard, mouse and window events. However, in the Newsqueak language, pipes are dynamically created and belong to the first class of values, so they can be saved to variables.

Newsqueak is similar to the scripting language, with a `print` function built in, and its "Hello World" program doesn't see anything special:

```go
Print("Hello,", "World", "\n");
```

From the above program, in addition to guessing that the `print` function can support multiple parameters, it is difficult to see the Newsqueak language-related features. Because the Newsqueak language and Go language related features are mainly concurrency and pipeline. Therefore, we take a look at the features of the Newsqueak language through a concurrent version of the "prime sieve" algorithm. The principle of "prime sieve" is as follows:

![](../images/ch1-5-prime-sieve.png)

*Figure 1-5 Prime Screen*

The "prime sieve" program for the concurrent version of the Newsqueak language is as follows:

```go
// Output a sequence of natural numbers starting at 2 to the pipeline
Counter := prog(c:chan of int) {
i := 2;
For(;;) {
c <-= i++;
}
};

// Filter the number of multiples of prime for the sequence obtained by the listen pipeline
// new sequence output to the send pipeline
Filter := prog(prime:int, listen, send:chan of int) {
i:int;
For(;;) {
If((i = <-listen)%prime) {
Send <-= i;
}
}
};

// main function
// The first outflow of each pipe must be a prime number
// then build a new prime filter based on this new prime number
Sieve := prog() of chan of int {
c := mk(chan of int);
Begin counter(c);
Prime := mk(chan of int);
Begin prog(){
p:int;
Newc:chan of int;
For(;;){
Prime <-= p =<- c;
Newc = mk();
Begin filter(p, c, newc);
c = newc;
}
}();
Become prime;
};

// Start the prime screen
Prime := sieve();
```

The `counter` function is used to output the original sequence of natural numbers to the pipeline. Each `filter` function object corresponds to each new prime filter pipeline. These prime filter pipelines filter the number of columns flowing into the input pipeline according to the current prime sieve. Output to the output pipe. `mk(chan of int)` is used to create a pipeline, similar to the Go language's `make(chan int)` statement; `begin filter(p, c, newc)` keyword to start a prime number of concurrency, similar to Go `go filter(p, c, newc)` statement; `become` is used to return the result of the function, similar to the `return` statement.

The syntax of the concurrency and pipeline in the Newsqueak language is quite similar to the Go language. The post-type declaration is similar to the Go language.

## 1.2.4 Alef - Phil Winterbottom, 1993

Before the emergence of the Go language, the Alef language was a perfect concurrency language in the author's mind, and the Alef syntax and runtime were basically seamlessly compatible with the C language. The Alef language provides support for threads and process concurrency, where `proc receive(c)` is used to start a process, `task receive(c)` is used to start a thread, and they are passed through a pipe `c `Communicate. However, due to the lack of automatic memory recovery mechanism of Alef, the memory resource management of the concurrent body is extremely complicated. Moreover, the Alef language only provides short-term support in the Plan9 system, and other operating systems do not have an actual Alef development environment. Moreover, the Alef language has only two public documents, the Alef Language Specification and the Alef Programming Wizard. Therefore, there is not much discussion about the Alef language outside Bell Labs.

Since the Alef language supports both process and thread concurrency, and more concurrents can be started again in the concurrency, the concurrent state of Alef is extremely complicated. At the same time, Alef does not have an automatic garbage collection mechanism (Alef has a flexible pointer feature for the reserved C language, which also makes the automatic garbage collection mechanism difficult to implement). Various resources are flooded between different threads and processes, resulting in concurrent memory resources. Management is extremely complicated. The Alef language inherits the syntax of the C language and can be considered as the C language that enhances the concurrent syntax. The following image is a possible concurrency state shown in the Alef language documentation:

![](../images/ch1-6-alef.png)

*Figure 1-6 Alef Concurrency Model*

The "Hello World" program for the concurrent version of the Alef language is as follows:

```c
#include <alef.h>

Void receive(chan(byte*) c) {
Byte *s;
s = <- c;
Print("%s\n", s);
Terminate(nil);
}

Void main(void) {
Chan(byte*) c;
Alloc c;
Proc receive(c);
Task receive(c);
c <- = "hello proc or task";
c <- = "hello proc or task";
Print("done\n");
Terminate(nil);
}
```

The `#include <alef.h>` statement at the beginning of the program is used to include the runtime library for the Alef language. `receive` is a normal function that is used as the entry function for each concurrency in the program; the `alloc c` statement in the `main` function first creates a `chan(byte*)` type of pipeline, similar to the Go language. Make(chan []byte)` statement; then start the `receive` function in process and thread mode respectively; after starting the concurrency, the `main` function sends two string data to the `c` pipe; and the process and thread The `receive` function of the state run willThe indeterminate order successively receives the data from the pipeline and then prints the strings separately; finally each concurrency ends itself by calling `terminate(nil)`.

Alef's grammar is basically the same as C language. It can be considered as a C++ language based on the grammar of C language. It can be regarded as another dimension of C++ language.

## 1.2.5 Limbo - Sean Dorward, Phil Winterbottom, Rob Pike, 1995

Limbo (Hell) is a programming language for developing distributed applications running on small computers. It supports modular programming, strong type checking at compile time and runtime, in-process communication pipeline based on type, atomic garbage collection. And simple abstract data types. Limbo is designed to operate safely even on small devices without hardware memory protection. The Limbo language runs primarily on the Inferno system.

The Limbo language version of the "Hello World" program is as follows:

```go
Implement Hello;

Include "sys.m"; sys: Sys;
Include "draw.m";

Hello: module
{
Init: fn(ctxt: ref Draw->Context, args: list of string);
};

Init(ctxt: ref Draw->Context, args: list of string)
{
Sys = load Sys Sys->PATH;
Sys->print("hello, world\n");
}
```

From this version of the "Hello World" program, we can already find a lot of prototypes of Go language features. The first sentence `implement Hello;` basically corresponds to the Go language's `package Hello` package declaration statement. Then the `include "sys.m"; sys: Sys;` and `include "draw.m";` statements are used to import other modules, similar to Go's `import "sys"` and `import "draw". Statement. Then the Hello package module also provides the module initialization function `init`, and the type of the function's parameters is also postfixed, but the Go language initialization function has no parameters.

## 1.2.6 Go Language - 2007~2009

Bell Labs later experienced several turmoil, and the original team of the Plan9 project, including Ken Thompson, eventually joined Google. After inventing the predecessor language such as Limbo more than 10 years later, at the end of 2007, the three original authors of the Go language gathered together to fight C++ because of accidental factors (the legend is that C++ language evangelists advocated C++11 everywhere in Google). The horrific features completely annoyed them), they finally took 20% of the free time to create the Go language. The original Go language specification was written in March 2008, and the original Go program was compiled directly into C and then recompiled into machine code. In May 2008, Google’s leaders finally discovered the great potential of the Go language and began to fully support the project (Google’s founders even contributed the `func` keyword) so they could put all their work time into action. Go to the design and development of the Go language. After the first version of the Go language specification is completed, the Go language compiler can finally generate machine code directly.

### 1.2.6.1 hello.go - June 2008

```go
Package main

Func main() int {
Print "hello, world\n";
Return 0;
}
```

This is the version that the initial Go language program officially started testing. The built-in `print` statement for debugging already exists, but it is used as a command. The entry `main` function also returns the value of the `int` type like the `main` function in C, and requires `return` to explicitly return the value. The semicolon at the end of each statement also exists.

### 1.2.6.2 hello.go - June 27, 2008

```go
Package main

Func main() {
Print "hello, world\n";
}
```

The entry function `main` has removed the return value, and the program returns by default by implicitly calling `exit(0)`. The Go language evolved in a simple direction.

### 1.2.6.3 hello.go - August 11, 2008

```go
Package main

Func main() {
Print("hello, world\n");
}
```

The built-in `print` for debugging is changed from the start command to the normal built-in function, making the syntax simpler and more consistent.

### 1.2.6.4 hello.go - October 24, 2008

```go
Package main

Import "fmt"

Func main() {
Fmt.printf("hello, world\n");
}
```

The `printf` formatting function, which is a signboard in the C language, has been ported to the Go language, and the function is placed in the `fmt` package (`fmt` is the abbreviation for the format word `format`). However, the beginning of the `printf` function name is still lowercase, with uppercase letters indicating that the exported feature has not yet appeared.

### 1.2.6.5 hello.go - January 15, 2009

```go
Package main

Import "fmt"

Func main() {
fmt.Printf("hello, world\n");
}
```

The Go language starts with whether the first letter of the capital is used to distinguish whether the symbol can be exported. Uppercase letters begin with the exported public symbol, and lowercase letters begin with the private symbol inside the package. Domestic users need to pay attention to the fact that there is no concept of uppercase and lowercase letters in Chinese characters. Therefore, symbols beginning with Chinese characters cannot be exported at present (for Chinese users who have already given relevant suggestions for problems, after Go2, they may adjust the export rules for Chinese characters) .

### 1.2.6.7 hello.go - December 11, 2009

```go
Package main

Import "fmt"

Func main() {
fmt.Printf("hello, world\n")
}
```

The Go language finally removed the semicolon at the end of the statement. This is the first important grammar improvement after Go officially opened source on November 10, 2009. From the rule of semicolon segmentation introduced in the first edition of the C language tutorial in 1978, the authors of the Go language have finally removed the semicolon at the end of the sentence for a full 32 years. In the course of this 32-year evolution, it is bound to be full of various gossip stories. I think this must be the result of deliberation by Go language designers. Now new languages ​​such as Swift also ignore semicolons by default. importance).


## 1.2.7 Hello, the world! - V2.0

After half a century of Nirvana rebirth, Go language not only prints the Unicode version of "Hello, World", but also provides print services to users around the world. The following version prints the Chinese "Hello, World!" and current time information to each client accessed via the `http` service.

```go
Package main

Import (
"fmt"
"log"
"net/http"
"time"
)

Func main() {
fmt.Println("Please visit http://127.0.0.1:12345/")
http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
s := fmt.Sprintf("Hello, World! -- Time: %s", time.Now().String())
fmt.Fprintf(w, "%v\n", s)
log.Printf("%v\n", s)
})
If err := http.ListenAndServe(":12345", nil); err != nil {
log.Fatal("ListenAndServe: ", err)
}
}
```

We constructed a stand-alone http service through the `net/http` package that comes with the Go language standard library. Where `http.HandleFunc("/", ...)` is registered with the response handler for the `/` root path request. In the response handler, we still use the `fmt.Fprintf` formatted output function to print the formatted string to the requesting client via the http protocol, and also print the relevant string on the server side through the standard library's log package. . Finally, the http service is started by the `http.ListenAndServe` function call.

At this point, the Go language has finally completed the transformation from the C language of the single-core single-core era to the general-purpose programming language of the multi-core environment of the Internet era in the 21st century.