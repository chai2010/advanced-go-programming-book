# 1.1 Go Language Genesis

The Go language was originally designed and invented by Google's Robert Griesemer, Ken Thompson and Rob Pike in 2007. The initial effort to design a new language comes from the touting report on the super-complex C++11 features. Defiance, the ultimate goal is to design the C language in the network and multi-core era. By the middle of 2008, most of the language's feature design had been completed, and began to implement the compiler and runtime, about the same year Russ Cox as the main developer. By 2009, the Go language has gradually stabilized. In September of the same year, the Go language was officially released and open sourced.

The Go language is often described as "C-like language" or "C language in the 21st century." From various angles, the Go language inherits a lot of programming ideas from the C language, such as expression syntax, control flow structure, basic data types, call parameter values, pointers, etc., and completely inherits and promotes the C language. Direct violence programming philosophy, etc. Figure 1-1 shows the genetic map of the Go language given in the Go Language Bible. We can see which programming languages ​​have influenced the Go language.

![](../images/ch1-1-go-family-tree.png)

_Figure 1-1 Go language gene family spectrum_

First look at the left side of the genetic map. It can be clearly seen that the concurrency of the Go language evolved from the CSP theory published by Bell Labs' Hoare in 1978. Later, the CSP concurrency model was gradually refined and applied to practical applications in programming languages ​​such as Squeak/NewSqueak and Alef. These design experiences were eventually absorbed and absorbed into the Go language. The concurrent programming model of the familiar Erlang programming language is another implementation of CSP theory.

Look at the middle of the genetic map. The middle one mainly contains the evolution of object-oriented and package features in the Go language. The Go language package and interface and object-oriented features are inherited from the Pascal language designed by Niklaus Wirth and the related programming languages ​​derived from it. The syntax of the package concept, package import and declaration is mainly from the Modula-2 programming language. The declaration syntax of the methods provided by the object-oriented features comes from the Oberon programming language. In the end, the Go language evolved its own unique features such as the implicit interface that supports the duck object-oriented model.

Finally, the right side of the gene map, this is a tribute to the C language. Go language is the most thorough abandonment of C language, not only the grammar and C language have many differences, the most important thing is to abandon the flexible but dangerous pointer operation in C language. Moreover, the Go language has redesigned the priority of some of the less reasonable operators in the C language, and has done the necessary polishing and changes in many subtle places. Of course, the lesser and more direct violent programming philosophy in the C language is further developed by the Go language (the Go language has only 25 keywords, and the sepc language specification is less than 50 pages).

Some of the other features of the Go language come from other programming languages. For example, the iota syntax is borrowed from the APL language. Features such as lexical scope and nested functions come from the Scheme language (and many other programming languages). There are also many designs in the Go language that invented and innovated. For example, the Go language slice provides efficient random access performance for lightweight dynamic arrays, which may be reminiscent of the underlying sharing mechanism of the linked list. There is also a new defer statement in the Go language (Ken invention) is also a pen.

## 1.1.1 Genes from Bell Labs

The iconic concurrent programming feature of the Go language comes from the little-known basic literature on concurrency research published by Tony Hoare of Bell Labs in 1978: the commutative sequential processes (CSP). In the original CSP paper, the program was just a set of parallel running processes with no intermediate shared state, using pipes for communication and control synchronization. Tony Hoare's CSP concurrency model is just a description language for describing the basic concepts of concurrency. It is not a general-purpose programming language for writing executable programs.

The most classic practical application of the CSP concurrency model is the Erlang programming language invented by Ericsson. However, while Erlang used CSP theory as a concurrent programming model, Rob Pike, who also came from Bell Labs, and his colleagues were constantly trying to introduce the CSP concurrency model into the newly invented programming language of the time. The first time they tried to introduce the CSP concurrency feature, the programming language called Squeak, was a programming language for providing mouse and keyboard event processing in which pipes were created statically. Then there is the improved version of the Newsqueak language (the new version of the mouse's voice), new syntax similar to C language statements and expressions, and derivation syntax similar to the Pascal language. Newsqueak is a purely functional language with garbage collection that is once again managed for keyboard, mouse and window events. However, in the Newsqueak language, the pipeline is already dynamically created. The pipeline belongs to the first type of value and can be saved to the variable. Then there is the Alef programming language (Alef is also the favorite programming language of Ritchie, the father of C language). The Alef language tries to transform the Newsqueak language into a system programming language, but it is painful to have concurrent programming due to the lack of garbage collection mechanism (this is also the inheritance of C language). The cost of manually managing memory). There is also a programming language called Limbo after the Aelf language (meaning hell), which is a scripting language that runs in a virtual machine. The Limbo language is the closest ancestor of the Go language, and it has the closest grammar to the Go language. By designing the Go language, Rob Pike has accumulated decades of experience in the practice of CSP concurrent programming models. The characteristics of concurrent programming in Go language are completely handy, and the arrival of new programming languages ​​is also a matter of course.

Figure 1-2 shows the most straightforward evolution of the Go code library's early codebase logs (Git is viewed with the `git log --before={2008-03-03} --reverse` command).

![](../images/ch1-2-go-log4.png)

_Figure 1-2 Go language development log_

It can also be seen from the early submission log that the Go language is gradually evolved from the B language invented by Ken Thompson and the C language invented by Dennis M. Ritchie. It is first a member of the C language family, so many people call the Go language. For the 21st century C language.

Figure 1-3 shows the evolution of the unique concurrent programming genes from Bell Labs in Go:

![](../images/ch1-3-go-history.png)

_Figure 1-3 Go language concurrent evolution history_

Throughout the development process of the entire Bell Labs programming language, from the B language, C language, Newsqueak, Alef, Limbo language, Go language inherits the half-century software design gene from Bell Labs, and finally completed The mission of C language innovation. Throughout the past few years, Go has become the most important basic programming language in the era of cloud computing and cloud storage.

## 1.1.2 Hello, the world

By convention, the first program that introduces all programming languages ​​is "Hello, World!". Although this teaching assumes that the reader already understands the Go language, we still don't want to break this convention (because this tradition is derived from the predecessor of the Go language C language). The code below shows that the Go language program outputs Chinese "Hello, World!".

```Go
Package main

Import "fmt"

Func main() {
fmt.Println("Hello, World!")
}
```

Save the above code in the `hello.go` file. Because there are non-ASCII Chinese characters in the code, we need to explicitly specify the encoding of the file as a UTF8 encoding without BOM (the source file is UTF8 encoded as required by the Go language specification). Then enter the command line and switch to the directory where the `hello.go` file is located. Currently we can use the Go language as a scripting language, and run `go run hello.go` directly from the command line to run the program. If everything is ok. It should be possible to see the result of the output "Hello, World!" on the command line.

Now let's briefly introduce the program. All Go programs are composed of the most basic functions and variables. The functions and variables are organized into separate Go source files. These source files are organized into appropriate packages according to the author's intention. Finally, these packages are organic. The ground constitutes a complete Go language program. The function is used to contain a series of statements (indicating the sequence of operations to be performed) and the variables that hold the data when the operation is performed. The name of the function in our program is main. Although there are not many restrictions on the name of a function in Go, the main function in the main package defaults to the entry point of each executable. Packages are used to wrap and organize related functions, variables, and constants. Before using a package, we need to import the package using the import statement. For example, we imported the fmt package (fmt is the abbreviation of the format word, indicating the format related package), and then we can use the Println function in the fmt package.

The double quotes contain "Hello, the world!" is the string denomination constant of the Go language. Unlike strings in C, the contents of strings in Go are immutable. When passing a string as a parameter to the fmt.Println function, the contents of the string are not copied - only the address and length of the string are passed (the structure of the string is defined in `reflect.StringHeader`). In the Go language, function arguments are passed in a way that is copied (not supported by reference) (in particular, the Go language closure function is used by reference to external variables).
