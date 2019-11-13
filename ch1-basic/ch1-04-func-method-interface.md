# 1.4 Functions, Methods, and Interfaces

The function corresponds to the sequence of operations and is the basic component of the program. Functions in the Go language have a name and anonymity: a named function generally corresponds to a package-level function. It is a special case of an anonymous function. When an anonymous function references a variable in an external scope, it becomes a closure function. Package functions are the core of a functional programming language. The method is bound to a special function of a specific type. The methods in the Go language are dependent on the type and must be statically bound at compile time. An interface defines a collection of methods that depend on the interface object at runtime, so the methods corresponding to the interface are dynamically bound at runtime. The Go language implements the duck object-oriented model through an implicit interface mechanism.

The initialization and execution of the Go language program always starts with the `main.main` function. However, if the `main` package imports other packages, they will be included in the `main` package in order (the import order here depends on the implementation, and may be imported in the string order of the file name or package path name) . If a package is imported multiple times, it will only be imported once during execution. When a package is imported, if it also imports other packages, first include the other packages, then create and initialize the constants and variables of the package, and then call the `init` function in the package. If a package has If multiple `init` functions are used, the calling order is undefined (the implementation may be called in the order of the file name), and multiple `init`s in the same file are called in the order in which they appear (`init` is not a normal function, There can be more than one definition, so it can't be called by other functions). Finally, when all the package-level constants, variables, and initialization of the `main` package are completed, and the `init` function is executed, the `main.main` function will be entered and the program will start executing normally. The following figure is a schematic diagram of the startup sequence of the Go program function:

![](../images/ch1-11-init.ditaa.png)

*Figure 1-11 Package Initialization Process*


It should be noted that all code runs in the same goroutine before the execution of the `main.main` function, which is the main system thread of the program. Therefore, if a `init` function internally starts a new goroutine with the go keyword, the new goroutine may only be executed after entering the `main.main` function.

## 1.4.1 Function

In Go, a function is the first type of object, and we can keep the function in a variable. Functions are mainly named and anonymous. Package-level functions are generally named functions. Named functions are a special case of anonymous functions. Of course, each type in the Go language can also have its own methods, which is actually a function.

```go
// named function
Func Add(a, b int) int {
Return a+b
}

// anonymous function
Var Add = func(a, b int) int {
Return a+b
}
```

A function in the Go language can have multiple parameters and multiple return values. Both the parameter and the return value exchange data with the callee in a way that is passed by value. Syntactically, the function also supports a variable number of parameters, the variable number of parameters must be the last parameter, and the variable number of parameters is actually a slice type parameter.

```go
// multiple parameters and multiple return values
Func Swap(a, b int) (int, int) {
Return b, a
}

// variable number of parameters
// more corresponds to []int slice type
Func Sum(a int, more ...int) int {
For _, v := range more {
a += v
}
Return a
}
```

When the mutable argument is a null interface type, whether the caller unpacks the mutable argument will result in different results:

```go
Func main() {
Var a = []interface{}{123, "abc"}

Print(a...) // 123 abc
Print(a) // [123 abc]
}

Func Print(a ...interface{}) {
fmt.Println(a...)
}
```

The argument passed in the first `Print` call is `a...`, which is equivalent to calling `Print(123, "abc")` directly. The second `Print` call is passed to the unwrapped `a`, which is equivalent to calling `Print([]interface{}{123, "abc"})` directly.

Not only can a function's argument have a name, but it can also give the function's return value a name:

```go
Func Find(m map[int]int, key int) (value int, ok bool) {
Value, ok = m[key]
Return
}
```

If the return value is named, you can modify the return value by name, or you can modify the return value after the `return` statement with the `defer` statement:

```go
Func Inc() (v int) {
Defer func(){ v++ } ()
Return 42
}
```

The `defer` statement delays the execution of an anonymous function because this anonymous function captures the local variable `v` of the external function. This function is generally called a closure. The closure is not a value-by-value access to the captured external variable, but is accessed by reference.

The behavior of accessing external variables by this reference to closures can lead to some hidden problems:

```go
Func main() {
For i := 0; i < 3; i++ {
Defer func(){ println(i) } ()
}
}
// Output:
// 3
// 3
// 3
```

Because it is a closure, in the `for` iteration statement, each function deferred by the `defer` statement is referenced by the same `i` iteration variable. After the end of the loop, the value of this variable is 3, so the final output is Both are 3.

The idea is to generate a unique variable for each `defer` function in each iteration. There are two ways to do this:

```go
Func main() {
For i := 0; i < 3; i++ {
i := i // Define a local variable in the loop body i
Defer func(){ println(i) } ()
}
}

Func main() {
For i := 0; i < 3; i++ {
// pass in the function i
// The defer statement will evaluate the call parameters immediately
Defer func(i int){ println(i) } (i)
}
}
```

The first method is to define a local variable inside the loop body, so that the closure function of each iteration of the `defer` statement captures different variables whose values ​​correspond to the values ​​at the time of iteration. The second way is to pass the iterator variable through the parameters of the closure function, and the `defer` statement will immediately evaluate the call parameters. Both methods work. However, in general, executing the `defer` statement inside the `for` loop is not a good habit. It is only an example and is not recommended.

In the Go language, if a function is called with a slice as a parameter, sometimes an illusion of passing a reference is given to the parameter: because the element of the incoming slice can be modified inside the called function. In fact, any situation in which a call parameter can be modified by a function parameter is because the pointer parameter is explicitly or implicitly passed in the function parameter. The specification of the function parameter value is more accurate. It only refers to the fixed part of the data structure, such as the pointer or string length structure in the string or slice corresponding structure, but does not contain the pointer indirect pointing content. . Replacing the parameters of the slice type with a structure like `reflect.SliceHeader` is a good understanding of the meaning of the slice value:

```go
Func twice(x []int) {
For i := range x {
x[i] *= 2
}
}

Type IntSliceHeader struct {
Data []int
Len int
Cap int
}

Func twice(x IntSliceHeader) {
For i := 0; i < x.Len; i++ {
x.Data[i] *= 2
}
}
```

Because the underlying array part of the slice is passed by the implicit pointer (the pointer itself is still passed, but the pointer points to the same data), the called function can modify the data in the call parameter slice by pointer. . In addition to the data, the slice structure also contains slice length and slice capacity information, and these two pieces of information are also passed by value. If the `Len` or `Cap` information is modified in the called function, it cannot be reflected in the slice of the calling parameter. At this time, we usually update the previous slice by returning the modified slice. This is why the built-in `append` must return a slice.

In the Go language, functions can also call themselves directly or indirectly, that is, support recursive calls. There is no limit to the depth of the recursive call of the Go language function. The stack of the function call does not have an overflow error, because the Go language runtime dynamically adjusts the size of the function stack as needed. Each goroutine will only allocate a small stack (4 or 8KB, depending on the implementation) just after startup. The stack size can be dynamically adjusted as needed. The stack can reach the GB level (depending on the implementation, in the current implementation, 32 bits) The architecture is 250MB and the 64-bit architecture is 1GB). Prior to Go1.4, Go's dynamic stack used a segmented dynamic stack. In layman's terms, a linked list was used to implement dynamic stacks. The memory locations of the nodes in each linked list did not change. However, the dynamic stack implemented by the linked list has a greater impact on the performance of some hotspot calls that cause different nodes across the linked list, because adjacent linked list nodes are generally not adjacent in memory locations, which increases the chance of CPU cache hit failure. In order to solve the CPU cache hit rate problem of hotspot calls, Go1.4 uses a continuous dynamic stack implementation, that is, a structure similar to a dynamic array to represent the stack. However, the continuous dynamic stack also brings a new problem: when the continuous stack grows dynamically, it needs to move the previous data to the new memory space, which will cause the address of all variables in the previous stack to change. Although the Go language runtime automatically updates pointers to stack variables that reference address changes, the most important point is to understand that pointers in Go are no longer fixed (so you can't hold pointers to numeric variables at will, The address of the Go language cannot be saved to the environment that is not controlled by the GC, so the address of the Go language object cannot be held in the C language for a long time when using CGO.

Because the stack of Go language functions will automatically resize, ordinary Go programmers rarely need to care about the stack's operating mechanism. In the Go language specification, even the concept of stack and heap is not deliberately mentioned. We can't know if a function parameter or a local variable is stored on the stack or in the heap. We just need to know that they work fine. Take a look at the following example:

```go
Func f(x int) *int {
Return &x
}

Func g() int {
x = new(int)
Return *x
}
```

The first function directly returns the address of the function parameter variable - this seems to be impossible, because if the parameter variable is on the stack, the stack variable will be invalid after the function returns, and the returned address should naturally fail. But the compiler and runtime of the Go language are much smarter than us, and it guarantees that the variables pointed to by the pointer are in the right place. The second function, although the `new` function is called internally, creates a pointer object of type `*int`, but still does not know where it is stored. For programmers with C/C++ programming experience, it is important to emphasize that the compiler and the runtime will help us without worrying about the function stack and heap in Go. Also, don't assume that the position of the variable in memory is fixed. The pointer may change at any time, especially if you don't expect it to change.

## 1.4.2 Method

The method is generally a feature of object-oriented programming (OOP). In the C++ language, the method corresponds to a member function of a class object, which is associated with a virtual table on a specific object. However, the Go language method is associated with the type, so that the static binding of the method can be completed during the compilation phase. An object-oriented program uses methods to express the operations of its properties, so that users who use the object do not need to directly manipulate the object, but use methods to do these things. Object-oriented programming (OOP) enters the mainstream development field is generally considered to start from C++. C++ supports object-oriented features such as class on the basis of compatible C language. Then Java programming is called a pure object-oriented language, because functions in Java cannot exist independently, and each function must belong to a certain class.

Object-oriented programming is more of an idea. Many languages ​​that claim to support object-oriented programming simply incorporate features that are often used into the language. Although the Go language ancestor C language is not an object-oriented language, the File-related functions in the C language standard library also use the idea of ​​object-oriented programming. Below we implement a set of C language style File functions:

```go
// file object
Type File struct {
Fd int
}

// open a file
Func OpenFile(name string) (f *File, err error) {
//...
}

// close the file
Func CloseFile(f *File) error {
// ...
}

// read file data
Func ReadFile(f *File, offset int64, data []byte) int {
// ...
}
```

The `OpenFile`-like constructor is used to open file objects, the `CloseFile`-like destructor is used to close file objects, and the `ReadFile` is similar to ordinary member functions. These three functions are ordinary functions. `CloseFile` and `ReadFile` are ordinary functions that need to occupy the name resource in the package level space. However, the `CloseFile` and `ReadFile` functions are only for the operation of `File` type objects. At this time, we prefer that such functions and the types of operation objects are tightly bound together.

The Go language is to move the first argument of the `CloseFile` and `ReadFile` functions to the beginning of the function name:

```go
// close the file
Func (f *File) CloseFile() error {
// ...
}

// read file data
Func (f *File) ReadFile(offset int64, data []byte) int {
// ...
}
```

In this case, the `CloseFile` and `ReadFile` functions become methods unique to the `File` type (instead of the `File` object method). They also no longer occupy the name resources in the package-level space, and the `File` type has already clarified their operation objects, so the method names are generally simplified to `Close` and `Read`:

```go
// close the file
Func (f *File) Close() error {
// ...
}

// read file data
Func (f *File) Read(offset int64, data []byte) int {
// ...
}
```

Moving the first function argument to the front of the function is a minor change from a code perspective, but from a programming philosophy point of view, the Go language is already in the ranks of object-oriented languages. We can add one or more methods to any custom type. The method for each type must be in the same package as the type definition, so it is not possible to add methods to built-in types like `int` (because the definition of the method and the definition of the type are not in a package). For a given type, the name of each method must be unique, and methods and functions do not support overloading.

The method is derived from the function, just moving the first object parameter of the function to the front of the function name. So we can still use the method in the original procedural thinking. You can restore a method to a normal type of function by calling the properties of a method expression:

```go
// does not depend on specific file objects
// func CloseFile(f *File) error
Var CloseFile = (*File).Close

// does not depend on specific file objects
// func ReadFile(f *File, offset int64, data []byte) int
Var ReadFile = (*File).Read

/ / File processing
f, _ := OpenFile("foo.dat")
ReadFile(f, 0, data)
CloseFile(f)
```

In some scenarios, I care more about a similar set of operations: for example, `Read` reads some arrays and then calls `Close` to close. In this environment, the user does not care about the type of the operation object, as long as it can satisfy the general `Read` and `Close` behaviors. However, in the method expression, because the `ReadFile` and `CloseFile` function parameters contain the unique type parameter of `File`, this makes the `File` related method not the same as the other `File` type but the same. The objects of the Read` and `Close` methods are seamlessly adapted. This kind of small difficulty can't help us Go language language farmers, we can eliminate the difference of the first parameter type in the method expression by combining the closure property:

```go
/ / Open the file object first
f, _ := OpenFile("foo.dat")

// bound to the f object
// func Close() error
Var Close = func() error {
Return (*File).Close(f)
}

// bound to the f object
// func Read(offset int64, data []byte) int
Var Read = func(offset int64, data []byte) int {
Return (*File).Read(f, offset, data)
}

/ / File processing
Read(0, data)
Close()
```

This is exactly the problem that the method value also needs to solve. We can simplify the implementation with method value features:

```go
/ / Open the file object first
f, _ := OpenFile("foo.dat")

// method value: bound to the f object
// func Close() error
Var Close = f.Close

// method value: bound to the f object
// func Read(offset int64, data []byte) int
Var Read = f.Read

/ / File processing
Read(0, data)
Close()
```

The Go language does not support the inheritance features of traditional object-oriented, but supports the inheritance of methods in its own unique combination. In the Go language, inheritance is achieved by building anonymous members in the structure:

```go
Import "image/color"

Type Point struct{ X, Y float64 }

Type ColoredPoint struct {
Point
Color color.RGBA
}
```

Although we can define `ColoredPoint` as a flat structure with three fields, we will embed `Point` in `ColoredPoint` to provide the fields `X` and `Y`.

```go
Var cp ColoredPoint
cp.X = 1
fmt.Println(cp.Point.X) // "1"
cp.Point.Y = 2
fmt.Println(cp.Y) // "2"
```

By embedding anonymous members, we can inherit not only the internal members of anonymous members, but also the methods corresponding to anonymous member types. We generally think of Point as a base class and ColoredPoint as its inherited or subclass. However, the method inherited in this way does not implement the polymorphic nature of virtual functions in C++. The recipient parameter for all inherited methods is still the anonymous member itself, not the current variable.

```go
Type Cache struct {
m map[string]string
sync.Mutex
}

Func (p *Cache) Lookup(key string) string {
p.Lock()
Defer p.Unlock()

Return p.m[key]
}
```

The `Cache` struct type inherits its `Lock` and `Unlock` methods by embedding an anonymous `sync.Mutex`. But when calling `p.Lock()` and `p.Unlock()`, ` P` is not the true recipient of the `Lock` and `Unlock` methods, but instead expands them to `p.Mutex.Lock()` and `p.Mutex.Unlock()` calls. This expansion is compiled. The period is completed, and there is no runtime cost.

In the traditional object-oriented language (eg. C++ or Java) inheritance, the subclass method is dynamically bound to the object at runtime, so some methods of the base class implementation see `this` may not be the base class. The type corresponding to the object, this feature will cause the uncertainty of the base class method to run. In the Go language, by embedding anonymous members to "inherit" the base class method, `this` is the object that implements the type of the method. The Go language method is statically bound at compile time. If you need the polymorphic nature of virtual functions, we need to implement it with the Go language interface.

## 1.4.3 Interface

Rob Pike, the father of the Go language, once said a famous saying: Languages ​​that try to avoid idiot behavior eventually become idiotic languages ​​(Languages ​​that try to disallow idiocy become themselves idiotic). General static programming languages ​​have strict type systems, which allows the compiler to drill down to see if the programmer has made any unusual moves. However, an overly strict type system can make programming too cumbersome, and let programmers waste a lot of youth in the struggle with the compiler. The Go language attempts to give programmers a balance between safe and flexible programming. It provides support for duck types through interface types while providing strict type checking, making safe and dynamic programming relatively easy.

Go's interface type is an abstraction and generalization of other types of behavior; because the interface type is not tied to specific implementation details, we can make the object more flexible and adaptable through this abstraction. Many object-oriented languages ​​have similar interface concepts, but the interface type in Go is unique in that it is a duck type that satisfies the implicit implementation. The so-called duck type says: As long as you walk like a duck and call it like a duck, you can use it as a duck. This is the case with object-oriented in the Go language. If an object looks like an implementation of an interface type, then it can be used as the interface type. This design allows you to create a new interface type that satisfies the existing type without having to destroy the original definition of those types; this design is especially flexible and useful when the type we use comes from a package that is not under our control. The interface type of the Go language is a delay binding, which can implement polymorphic functions like virtual functions.

The interface is ubiquitous in the Go language. In the "Hello world" example, the design of the `fmt.Printf` function is completely interface-based, and its real functionality is done by the `fmt.Fprintf` function. The `error` type used to indicate an error is a built-in interface type. In C, `printf` can only print several limited basic data types into file objects. However, the Go language flexible interface feature, `fmt.Fprintf` can print to any custom output stream object, can be printed to a file or standard output, can also be printed to the network, or even can be printed to a compressed file; at the same time, printed The data is not limited to the basic types built into the language. Any object that implicitly satisfies the `fmt.Stringer` interface can be printed. The interface that does not satisfy the `fmt.Stringer` can still be printed by the reflection technique. The signature of the `fmt.Fprintf` function is as follows:

```go
Func Fprintf(w io.Writer, format string, args ...interface{}) (int, error)
```

Where `io.Writer` is used for output interface, `error` is a built-in error interface, and their definitions are as follows:

```go
Type io.Writer interface {
Write(p []byte) (n int, err error)
}

Type error interface {
Error() string
}
```

We can output each character by converting it to uppercase characters by customizing its own output object:

```go
Type UpperWriter struct {
io.Writer
}

Func (p *UpperWriter) Write(data []byte) (n int, err error) {
Return p.Writer.Write(bytes.ToUpper(data))
}

Func main() {
fmt.Fprintln(&UpperWriter{os.Stdout}, "hello, world")
}
```

Of course, we can also define our own print format to achieve the effect of converting each character to uppercase characters. For each object to be printed, if the `fmt.Stringer` interface is satisfied, the result returned by the object's `String` method is printed by default:

```go
Type UpperString string

Func (s UpperString) String() string {
Return strings.ToUpper(string(s))
}

Type fmt.Stringer interface {
String() string
}

Func main() {
fmt.Fprintln(os.Stdout, UpperString("hello, world"))
}
```

In the Go language, implicit conversion is not supported for the underlying type (non-interface type). We cannot assign a value of type `int` directly to a variable of type `int64`, nor can we assign a value of type `int` to The underlying is a variable of the newly defined named type of the `int` type. The Go language's type consistency requirements for the underlying type are very strict, but the Go language is very flexible for interface type conversion. Conversion between objects and interfaces, transitions between interfaces and interfaces can all be implicit conversions. You can see the following example:

```go
Var (
a io.ReadCloser = (*os.File)(f) // implicit conversion, *os.File satisfies the io.ReadCloser interface
b io.Reader = a // implicit conversion, io.ReadCloser satisfies the io.Reader interface
c io.Closer = a // implicit conversion, io.ReadCloser satisfies the io.Closer interface
d io.Reader = c.(io.Reader) // Explicit conversion, io.Closer does not satisfy io.Reader interface
)
```

Sometimes the object and interface are too flexible, which leads us to artificially limit this unintentional adaptation. A common practice is to define a special method to distinguish interfaces. For example, the `Error` interface in the `runtime` package defines a unique `RuntimeError` method to prevent other types from inadvertently adapting the interface:

```go
Type runtime.Error interface {
Error

// RuntimeError is a no-op function but
//serve to distinguish types that are run time
// errors from ordinary errors: a type is a
// run time error if it has a RuntimeError method.
RuntimeError()
}
```

In protobuf, the `Message` interface also uses a similar method, and also defines a unique `ProtoMessage` to prevent other types from inadvertently adapting the interface:

```go
Type proto.Message interface {
Reset()
String() string
ProtoMessage()
}
```

However, this practice is only a gentleman's agreement. It is also very easy if someone deliberately forges a `proto.Message` interface. A more rigorous approach is to define a private method for the interface. Only objects that satisfy this private method can satisfy this interface, and the name of the private method contains the absolute path name of the package, so this private method can only be implemented inside the package to satisfy this interface. The `testing.TB` interface in the test package uses a similar technique:

```go
Type testing.TB interface {
Error(args ...interface{})
Errorf(format string, args ...interface{})
...

// A private method to prevent users implementing the
// interface and so future additions to it will not
// violate Go 1 compatibility.
Private()
}
```

However, this method of prohibiting external objects from implementing interfaces through private methods has a cost: first, this interface can only be used internally by the package, and external packages cannot normally create objects that satisfy the interface; secondly, this protection is also Not absolute, malicious users can still bypass this protection mechanism.

As we mentioned in the previous method section, you can inherit methods of anonymous types by embedding anonymous type members in the structure. In fact, this embedded anonymous member is not necessarily a normal type, but also an interface type. We can fake the private `private` method by embedding the anonymous `testing.TB` interface, because the interface method is a late binding, and it doesn't matter if the `private` method actually exists at compile time.

```go
Package main

Import (
"fmt"
"testing"
)

Type TB struct {
Testing.TB
}

Func (p *TB) Fatal(args ...interface{}) {
fmt.Println("TB.Fatal disabled!")
}

Func main() {
Var tb testing.TB = new(TB)
tb.Fatal("Hello, playground")
}
```

We reimplemented the `Fatal` method in our `TB` struct type and then implicitly converted the object to the `testing.TB` interface type (because the anonymous `testing.TB` object is embedded, so Satisfy the `testing.TB` interface) and then call our own `Fatal` method via the `testing.TB` interface.

This practice of embedding inheritance by embedding an anonymous interface or embedding an anonymous pointer object is actually a pure virtual inheritance. We inherit only the specification specified by the interface, and the real implementation is injected at runtime. For example, we can simulate a plugin that implements a gRPC:

```go
Type grpcPlugin struct {
*generator.Generator
}

Func (p *grpcPlugin) Name() string { return "grpc" }

Func (p *grpcPlugin) Init(g *generator.Generator) {
p.Generator = g
}

Func (p *grpcPlugin) GenerateImports(file *generator.FileDescriptor) {
If len(file.Service) == 0 {
Return
}

p.P(`import "google.golang.org/grpc"`)
// ...
}
```

The constructed `grpcPlugin` type object must satisfy the `generate.Plugin` interface (in the "github.com/golang/protobuf/protoc-gen-go/generator" package):

```go
Type Plugin interface {
// Name identifies the plugin.
Name() string
// Init is called once after data structures are built but before
// code generation begins.
Init(g *Generator)
// Generate produces the code generated by the plugin for this file,
//except for the imports, by calling the generator's methods
// P, In, and Out.
Generate(file *FileDescriptor)
// GenerateImports produces the import declarations for this file.
// It is called after Generate.
GenerateImports(file *FileDescriptor)
}
```

The `p.P(...)` function used in the `GenerateImports` method of the `grpcPlugin` type corresponding to the `generate.Plugin` interface is implemented by the `generator.Generator` object injected by the `Init` function. Here `generator.Generator` corresponds to a concrete type, but if `generator.Generator` is an interface type, we can even pass in a direct implementation.

The Go language easily implements advanced features such as duck object-oriented and virtual inheritance through a combination of several simple features, which is really incredible.