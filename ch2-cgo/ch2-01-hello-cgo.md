# 2.1 Quick Start

In this section, we will quickly grasp the basic usage of CGO through a series of small examples from the shallower to the deeper.

## 2.1.1 The simplest CGO program

Real CGO programs are generally more complicated. But we can go from shallow to deep, what is the simplest CGO program? To construct a minimal CGO program, first ignore some of the complex CGO features, and at the same time show the difference between CGO programs and pure Go programs. Here is the simplest CGO program we built:

```go
// hello.go
Package main

Import "C"

Func main() {
Println("hello cgo")
}
```

The code enables the CGO feature via the `import "C"` statement. The main function simply outputs the string via Go's built-in println function, which does not have any CGO-related code. Although the CGO related function is not called, the `go build` command will start the gcc compiler during the compile and link phases, which is already a complete CGO program.

## 2.1.2 Output string based on C standard library function

The first chapter of the CGO program is not simple enough, let's take a look at the simpler version:

```go
// hello.go
Package main

//#include <stdio.h>
Import "C"

Func main() {
C.puts(C.CString("Hello, World\n"))
}
```

We not only enable the CGO feature via the `import "C"` statement, but also include the C's `<stdio.h>` header file. Then convert the Go language string to a C language string through the CGO package's `C.CString` function, and finally call the CGO package's `C.puts` function to print the converted C string to the standard output window.

The biggest difference compared to the CGO program in the "Hello, World Revolution" section is that we didn't release the C language string created by `C.CString` before the program exits; we also use the `puts` function instead. Standard output printing, previously printed with standard output using `fputs`.

Failure to release a C-language string created with `C.CString` will result in a memory leak. But for this small program, this is no problem, because the operating system will automatically reclaim all the resources of the program after the program exits.

## 2.1.3 Using your own C function

Earlier we used the functions already in the standard library. Now let's customize a C function called `SayHello` to print, and then call this `SayHello` function from the Go locale:

```go
// hello.go
Package main

/*
#include <stdio.h>

Static void SayHello(const char* s) {
Puts(s);
}
*/
Import "C"

Func main() {
C.SayHello(C.CString("Hello, World\n"))
}
```

Except that the `SayHello` function is implemented by ourselves, the other parts are basically similar to the previous examples.

We can also put the `SayHello` function in a C source file in the current directory (the suffix must be `.c`). Because it is written in a separate C file, in order to allow external references, you need to remove the function's `static` modifier.

```c
// hello.c

#include <stdio.h>

Void SayHello(const char* s) {
Puts(s);
}
```

Then declare the `SayHello` function in the CGO section, leaving the rest unchanged:

```go
// hello.go
Package main

//void SayHello(const char* s);
Import "C"

Func main() {
C.SayHello(C.CString("Hello, World\n"))
}
```

Note that if the previously run command is `go run hello.go` or `go build hello.go`, you must use `go run "your/package"` or `go build "your/package"` . If you are in the package path, you can also run `go run .` or `go build` directly.

Since the `SayHello` function has been placed in a separate C file, we can naturally compile the corresponding C file into a static library or a dynamic library file for use. If the `SayHello` function is referenced in a static library or a dynamic library, the corresponding C source file needs to be moved out of the current directory (the CGO build program will automatically build the C source file in the current directory, causing C function name conflicts). Details such as static libraries will be explained later in the chapter.

## 2.1.4 Modification of C code

Abstraction and modularity are common means of simplifying complex problems during programming. When there are more code statements, we can wrap similar code into a single function; when there are more functions in the program, we split the function into different files or modules. The core of modular programming is programming interface programming (the interface here is not the interface of the Go language, but the concept of the API).

In the previous example, we can abstract a module named hello, and all the interface functions of the module are defined in the hello.h header file:

```c
// hello.h
Void SayHello(const char* s);
```

There is only one declaration for the SayHello function. But as a user of the hello module, you can safely use the SayHello function without worrying about the concrete implementation of the function. As the implementer of the SayHello function, the implementation of the function only needs to meet the specification of the function declaration in the header file. The following is the C language implementation of the SayHello function, corresponding to the hello.c file:

```c
// hello.c

#include "hello.h"
#include <stdio.h>

Void SayHello(const char* s) {
Puts(s);
}
```

At the beginning of the hello.c file, the implementer includes the declaration of the SayHello function via the `#include "hello.h"` statement, which ensures that the implementation of the function satisfies the interface exposed by the module.

The interface file hello.h is a common contract between the implementer and the user of the hello module, but the convention does not require that the C language be used to implement the SayHello function. We can also reimplement this C language function in C++:

```c++
// hello.cpp

#include <iostream>

Extern "C" {
#include "hello.h"
}

Void SayHello(const char* s) {
Std::cout << s;
}
```

In the C++ version of the SayHello function implementation, we output the stream output string through the C++-specific `std::cout` output stream. However, in order to ensure that the SayHello function implemented by the C++ language satisfies the function specification defined by the C language header file hello.h, we need to indicate the link symbol of the function to follow the C language rules through the `extern "C"` statement.

After programming with the C language API interface, we completely liberated the language shackles of the module implementer: the implementer can implement the module in any programming language, as long as the public API convention is finally met. We can implement the SayHello function in C language, or use the more complex C++ language to implement the SayHello function. Of course, we can reimplement the SayHello function in assembly language or even Go language.


## 2.1.5 Reimplement C function with Go

In fact, CGO is not only used to call C language functions in Go language, but also can be used to export Go language functions to C language function calls. In the previous example, we have abstracted a module named hello, and all the interface functions of the module are defined in the hello.h header file:

```c
// hello.h
Void SayHello(/*const*/ char* s);
```

Now let's create a hello.go file and reimplement the SayHello function of the C language interface in Go language:

```go
// hello.go
Package main

Import "C"

Import "fmt"

//export SayHello
Func SayHello(s *C.char) {
fmt.Print(C.GoString(s))
}
```

We use the `//export SayHello` directive of CGO to export the function `SayHello` implemented by Go to C language functions. In order to adapt the C language functions exported by CGO, we disable the const modifier in the declaration statement of the function. It should be noted that there are actually two versions of the `SayHello` function: one for the Go locale; the other is for the C locale. The C language version of the SayHello function generated by cgo will eventually call the Go version of the SayHello function via the bridge code.

Through the programming technology for the C language interface, we not only liberate the implementer of the function, but also simplify the user of the function. Now we can use SayHello as a function of a standard library (similar to the way the puts function is used):

```go
Package main

//#include <hello.h>
Import "C"

Func main() {
C.SayHello(C.CString("Hello, World\n"))
}
```

Everything seems to have returned to the beginning of the CGO code, but the code is richer.

## 2.1.6 Go programming for C interface

In the beginning example, all of our CGO code is in a Go file. Then, SayHello is split into different C files by the technology for C interface programming, and main is still a Go file. Then, the SayHello function of the C language interface is reimplemented with the Go function. But for the current example there is only one function, and splitting into three different files is a bit cumbersome.

The so-called long-term must be divided, long-term must be combined, we now try to re-consolidate several files in the example into a Go file. The following are the combined results:

```go
Package main

//void SayHello(char* s);
Import "C"

Import (
"fmt"
)

Func main() {
C.SayHello(C.CString("Hello, World\n"))
}

//export SayHello
Func SayHello(s *C.char) {
fmt.Print(C.GoString(s))
}
```

The proportion of C code in the current version of CGO code is very small, but we can still further refine our CGO code with the Go language. Through analysis, you can find that the parameters of the `SayHello` function are the most direct if you can directly use the Go string. In Go1.10, CGO added a new `_GoString_` predefined C language type to represent the Go language string. Here's the improved code:

```go
// +build go1.10

Package main

//void SayHello(_GoString_ s);
Import "C"

Import (
"fmt"
)

Func main() {
C.SayHello("Hello, World\n")
}

//export SayHello
Func SayHello(s string) {
fmt.Print(s)
}
```

Although it seems that all are Go language code, the implementation is from the `main` function of Go language to the C language version `SayHello` bridge function automatically generated by CGO, and finally returns to the `SayHello` function of Go language environment. . This code contains the essence of CGO programming, and the reader needs to understand it in depth.

* Questions: Is the main function and the SayHello function executed in the same Goroutine? *