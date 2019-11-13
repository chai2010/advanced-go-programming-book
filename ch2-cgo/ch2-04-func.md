# 2.4 Function Call

The function is the core of C language programming. Through CGO technology, we can not only call C language function in Go language, but also export Go language function as C language function.

## 2.4.1 Go calls C function

For a program that enables CGO features, CGO constructs a virtual C package. The C language function can be called by this virtual C package.

```go
/*
Static int add(int a, int b) {
Return a+b;
}
*/
Import "C"

Func main() {
C.add(1, 1)
}
```

The above CGO code first defines an add function visible in the current file, and then passes `C.add`.

## 2.4.2 Return value of C function

For a C function with a return value, we can get the return value normally.

```go
/*
Static int div(int a, int b) {
Return a/b;
}
*/
Import "C"
Import "fmt"

Func main() {
v := C.div(6, 3)
fmt.Println(v)
}
```

The above div function implements an integer division operation and returns the result of the division by the return value.

However, there is no special treatment for the case where the divisor is zero. If you want to return an error when the divisor is 0, other times return normal results. Because the C language does not support returning multiple results, the `<errno.h>` standard library provides an `errno` macro for returning error status. We can approximate `errno` as a thread-safe global variable that can be used to record the status code of the most recent error.

The improved div function is implemented as follows:

```c
#include <errno.h>

Int div(int a, int b) {
If(b == 0) {
Errno = EINVAL;
Return 0;
}
Return a/b;
}
```

CGO also has special support for the `errno` macro of the `errno.h>` standard library: if there are two return values ​​when CGO calls a C function, the second return value will correspond to the `errno` error state.

```go
/*
#include <errno.h>

Static int div(int a, int b) {
If(b == 0) {
Errno = EINVAL;
Return 0;
}
Return a/b;
}
*/
Import "C"
Import "fmt"

Func main() {
V0, err0 := C.div(2, 1)
fmt.Println(v0, err0)

V1, err1 := C.div(1, 0)
fmt.Println(v1, err1)
}
```

Running this code will produce the following output:

```
2 <nil>
0 invalid argument
```

We can approximate the div function as a function of the following types:

```go
Func C.div(a, b C.int) (C.int, [error])
```

The second return value is a negligible error interface type, and the underlying corresponds to the `syscall.Errno` error type.

## 2.4.3 Return value of void function

C language functions also have a function that does not return a value type, and void refers to the return value type. In general, we can't get the return value of a void type function, because there is no return value to get. As mentioned in the previous example, cgo does a special treatment for errno, and can get the error state of C language through the second return value. This feature is still valid for void type functions.

The following code is to get an error status code with no return value function:

```go
//static void noreturn() {}
Import "C"
Import "fmt"

Func main() {
_, err := C.noreturn()
fmt.Println(err)
}
```

At this point, we ignore the first return value and only get the error code corresponding to the second return value.

We can also try to get the first return value, which corresponds to the Go language type corresponding to the C language void:

```go
//static void noreturn() {}
Import "C"
Import "fmt"

Func main() {
v, _ := C.noreturn()
fmt.Printf("%#v", v)
}
```

Running this code will produce the following output:

```
main._Ctype_void{}
```

We can see that the void type of C language corresponds to the `_Ctype_void` type in the current main package. In fact, the C language noreturn function is also considered to return a function of type `_Ctype_void`, so that you can directly get the return value of the void type function:

```go
//static void noreturn() {}
Import "C"
Import "fmt"

Func main() {
fmt.Println(C.noreturn())
}
```

Running this code will produce the following output:

```
[]
```

In fact, in the CGO generated code, the `_Ctype_void` type corresponds to a 0 long array type `[0]byte`, so `fmt.Println` outputs a square bracket representing a null value.

Although the above effective features may seem boring, we can accurately grasp the boundaries of CGO code through these examples, and we can think about the reasons for these strange features from a deeper design perspective.


## 2.4.4 C calls the Go export function

CGO also has a powerful feature: exporting Go functions as C language functions. In this case, we can define the C language interface and then implement it through the Go language. In the first section of this chapter, we have shown examples of Go language export C language functions.

Here's the re-implementation of the add function starting with this section in Go:

```go
Import "C"

//export add
Func add(a, b C.int) C.int {
Return a+b
}
```

The add function name begins with a lowercase letter and is a private function in the package for the Go language. But from the perspective of C language, the exported add function is a C language function that can be accessed globally. If there is an add function of the same name to be exported as a C language function in two different Go language packages, the problem of symbolic double name will occur in the final link phase.

The `_cgo_export.h` file generated by CGO will contain the declaration of the exported C language function. We can include the `_cgo_export.h` file in the pure C source file to reference the exported add function. If you want to use the exported C language add function in the current CGO file, you can't reference the `_cgo_export.h` file. Because the generation of the `_cgo_export.h` file depends on the current file to build properly, and if the current file internal loop depends on the `_cgo_export.h` file that has not been generated, the cgo command will be incorrect.

```c
#include "_cgo_export.h"

Void foo() {
Add(1, 1);
}
```

When exporting the C language interface, you need to ensure that the function parameters and return value types are C-friendly types, and the return value must not directly or indirectly contain pointers to the Go language memory space.