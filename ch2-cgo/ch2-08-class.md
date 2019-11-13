# 2.8 C++ class packaging

CGO is a bridge between C and Go. In principle, C++ classes cannot be directly supported. The root cause of CGO's lack of support for C++ syntax is that C++ has not yet had a Binary Interface Specification (ABI). How a C++ class constructor generates link symbol names when compiled into object files, methods are different between different platforms and even different versions of C++. But C++ is compatible with C language, so we can add a set of C language function interface as a bridge between C++ class and CGO, so that the interconnection between C++ and Go can be realized indirectly. Of course, because CGO only supports data types of C language median types, we can't directly use C++ reference parameters and other features.

## 2.8.1 C++ class to Go language object

Implementing the packaging of a C++ class to a Go language object requires the following steps: first, the C++ class is wrapped with a pure C function interface; secondly, the pure C function interface is mapped to the Go function by CGO; finally, a Go wrapper object is created. Implement C++ classes into methods using Go objects.

### 2.8.1.1 Preparing a C++ Class

For the sake of simplicity, we make a simple cache class, MyBuffer, based on `std::string`. In addition to the constructor and destructor, only two member functions return the underlying data pointer and the size of the cache. Because it is a binary cache, we can place arbitrary data in it.

```c++
// my_buffer.h
#include <string>

Struct MyBuffer {
Std::string* s_;

MyBuffer(int size) {
This->s_ = new std::string(size, char('\0'));
}
~MyBuffer() {
Delete this->s_;
}

Int Size() const {
Return this->s_->size();
}
Char* Data() {
Return (char*)this->s_->data();
}
};
```

We specify the size of the cache and allocate space in the constructor, and release the internally allocated memory space through the destructor after use. Here's how to use it:

```c++
Int main() {
Auto pBuf = new MyBuffer(1024);

Auto data = pBuf->Data();
Auto size = pBuf->Size();

Delete pBuf;
}
```

In order to facilitate the transition to the C language interface, here we deliberately did not define the C++ copy constructor. We must allocate and release cached objects with new and delete, not in a value-style way.

### 2.8.1.2 Encapsulating C++ classes with pure C function interfaces

If you want to wrap the above C++ class with a C language function interface, we can start with the usage. We can map new and delete to C language functions, and map object methods to C language functions.

In the C language we expect the MyBuffer class to be used like this:

```c
Int main() {
MyBuffer* pBuf = NewMyBuffer(1024);

Char* data = MyBuffer_Data(pBuf);
Auto size = MyBuffer_Size(pBuf);

DeleteMyBuffer(pBuf);
}
```

First think about what interface you need from the perspective of the C language interface user, and then create the `my_buffer_capi.h` header file interface specification:

```c++
// my_buffer_capi.h
Typedef struct MyBuffer_T MyBuffer_T;

MyBuffer_T* NewMyBuffer(int size);
Void DeleteMyBuffer(MyBuffer_T* p);

Char* MyBuffer_Data(MyBuffer_T* p);
Int MyBuffer_Size(MyBuffer_T* p);
```

Then you can define these C language wrapper functions based on the C++ MyBuffer class. We create the corresponding `my_buffer_capi.cc` file as follows:

```c++
// my_buffer_capi.cc

#include "./my_buffer.h"

Extern "C" {
#include "./my_buffer_capi.h"
}

Struct MyBuffer_T: MyBuffer {
MyBuffer_T(int size): MyBuffer(size) {}
~MyBuffer_T() {}
};

MyBuffer_T* NewMyBuffer(int size) {
Auto p = new MyBuffer_T(size);
Return p;
}
Void DeleteMyBuffer(MyBuffer_T* p) {
Delete p;
}

Char* MyBuffer_Data(MyBuffer_T* p) {
Return p->Data();
}
Int MyBuffer_Size(MyBuffer_T* p) {
Return p->Size();
}
```

Because the header file `my_buffer_capi.h` is for CGO, it must be a name modification rule using the C language specification. The `extern "C"` statement is required when the C++ source file is included. In addition, the implementation of MyBuffer_T is just a class that inherits from MyBuffer, which simplifies the implementation of wrapper code. At the same time, when communicating with CGO, we must pass the `MyBuffer_T` pointer. We can't expose the specific implementation to CGO because the implementation contains C++-specific syntax, and CGO does not recognize C++ features.

After wrapping the C++ class as a pure C interface, the next step is to convert the C function to a Go function.

### 2.8.1.3 Converting a pure C interface function to a Go function

The process of wrapping a pure C function into a corresponding Go function is relatively simple. Note that because our package contains the C++11 syntax, we need to open the C++11 option with `#cgo CXXFLAGS: -std=c++11`.

```go
// my_buffer_capi.go

Package main

/*
#cgo CXXFLAGS: -std=c++11

#include "my_buffer_capi.h"
*/
Import "C"

Type cgo_MyBuffer_T C.MyBuffer_T

Func cgo_NewMyBuffer(size int) *cgo_MyBuffer_T {
p := C.NewMyBuffer(C.int(size))
Return (*cgo_MyBuffer_T)(p)
}

Func cgo_DeleteMyBuffer(p *cgo_MyBuffer_T) {
C.DeleteMyBuffer((*C.MyBuffer_T)(p))
}

Func cgo_MyBuffer_Data(p *cgo_MyBuffer_T) *C.char {
Return C.MyBuffer_Data((*C.MyBuffer_T)(p))
}

Func cgo_MyBuffer_Size(p *cgo_MyBuffer_T) C.int {
Return C.MyBuffer_Size((*C.MyBuffer_T)(p))
}
```

To distinguish, we add a `cgo_` prefix to each type and function name in Go. For example, cgo_MyBuffer_T is the type of MyBuffer_T in C.

For the sake of simplicity, when packaging a pure C function to a Go function, in addition to the cgo_MyBuffer_T type, we still use the C language type for the underlying types of input parameters and return values.

### 2.8.1.4 Wrapper as a Go object

After wrapping the pure C interface as a Go function, we can easily construct a Go object based on the wrapped Go function. Because cgo_MyBuffer_T is a type imported from C language space, it can't define its own method, so we construct a new MyBuffer type, which holds the C language cache object pointed to by cgo_MyBuffer_T.

```go
// my_buffer.go

Package main

Import "unsafe"

Type MyBuffer struct {
Cptr *cgo_MyBuffer_T
}

Func NewMyBuffer(size int) *MyBuffer {
Return &MyBuffer{
Cptr: cgo_NewMyBuffer(size),
}
}

Func (p *MyBuffer) Delete() {
cgo_DeleteMyBuffer(p.cptr)
}

Func (p *MyBuffer) Data() []byte {
Data := cgo_MyBuffer_Data(p.cptr)
Size := cgo_MyBuffer_Size(p.cptr)
Return ((*[1 << 31]byte)(unsafe.Pointer(data)))[0:int(size):int(size)]
}
```

At the same time, because the Go language slice itself contains length information, we merge the cgo_MyBuffer_Data and cgo_MyBuffer_Size functions into the `MyBuffer.Data` method, which returns a slice corresponding to the underlying C language cache space.

Now we can easily use the wrapped cache object in the Go language (the underlying is based on C++'s `std::string` implementation):

```go
Package main

//#include <stdio.h>
Import "C"
Import "unsafe"

Func main() {
Buf := NewMyBuffer(1024)
Defer buf.Delete()

Copy(buf.Data(), []byte("hello\x00"))
C.puts((*C.char)(unsafe.Pointer(&(buf.Data()[0])))))
}
```

In the example, we created a 1024-byte cache and then populated a string with the copy function. In order to facilitate the processing of C language string functions, we default to the end of the string with '\0' in the filled string. Finally, we directly get the underlying data pointer of the cache, and print the contents of the cache using the C language puts function.

## 2.8.2 Go language objects to C++ classes

To implement the packaging of Go language objects into C++ classes, the following steps are required: first, map the Go object to an id; then export the corresponding C interface function based on the id; finally, package the C++ object based on the C interface function.

### 2.8.2.1 Constructing a Go object

For the sake of demonstration, we built a Person object in Go, each of which can have name and age information:

```go
Package main

Type Person struct {
Name string
Age int
}

Func NewPerson(name string, age int) *Person {
Return &Person{
Name: name,
Age: age,
}
}

Func (p *Person) Set(name string, age int) {
P.name = name
P.age = age
}

Func (p *Person) Get() (name string, age int) {
Return p.name, p.age
}
```

If the Person object wants to be accessed in C/C++, it needs to be accessed via the cgo export C interface.

### 2.8.2.2 Export C interface

We modeled the C++ object to the C interface process, and abstracted a set of C interfaces to describe the Person object. Create a `person_capi.h` file that corresponds to the C interface specification file:

```c
// person_capi.h
#include <stdint.h>

typedef uintptr_t person_handle_t;

Person_handle_t person_new(char* name, int age);
Void person_delete(person_handle_t p);

Void person_set(person_handle_t p, char* name, int age);
Char* person_get_name(person_handle_t p, char* buf, int size);
Int person_get_age(person_handle_t p);
```

Then this set of C functions is implemented in the Go language.

It should be noted that when exporting C functions through CGO, both input parameters and return value types do not support const modification, and also do not support variable parameter function types. At the same time, as described in the Memory Mode section, we cannot directly access Go memory objects in C/C++ for a long time. So we used the technique described in the previous section to map the Go object to an integer id.

The following is the `person_capi.go` file, which corresponds to the implementation of the C interface function:

```go
// person_capi.go
Package main

//#include "./person_capi.h"
Import "C"
Import "unsafe"

//export person_new
Func person_new(name *C.char, age C.int) C.person_handle_t {
Id := NewObjectId(NewPerson(C.GoString(name), int(age)))
Return C.person_handle_t(id)
}

//export person_delete
Func person_delete(h C.person_handle_t) {
ObjectId(h).Free()
}

//export person_set
Func person_set(h C.person_handle_t, name *C.char, age C.int) {
p := ObjectId(h).Get().(*Person)
p.Set(C.GoString(name), int(age))
}

//export person_get_name
Func person_get_name(h C.person_handle_t, buf *C.char, size C.int) *C.char {
p := ObjectId(h).Get().(*Person)
Name, _ := p.Get()

n := int(size) - 1
bufSlice := ((*[1 << 31]byte)(unsafe.Pointer(buf)))[0:n:n]
n = copy(bufSlice, []byte(name))
bufSlice[n] = 0

Return buf
}

//export person_get_age
Func person_get_age(h C.person_handle_t) C.int {
p := ObjectId(h).Get().(*Person)
_, age := p.Get()
Return C.int(age)
}
```

After creating the Go object, we map the Go correspondence to id via NewObjectId. Then force the id to be escaped as the person_handle_t type. The other interface functions are based on the id represented by person_handle_t, so that the corresponding Go object is parsed according to the id.

### 2.8.2.3 Encapsulating C++ objects

Encapsulating C++ objects with the C interface is relatively straightforward. A common practice is to create a new Person class, which contains a member of type person_handle_t corresponding to the real Go object, and then create a Go object through the C interface in the constructor of the Person class, and release the Go object through the C interface in the destructor. Here's an implementation using this technique:

```c++
Extern "C" {
#include "./person_capi.h"
}

Struct Person {
Person_handle_t goobj_;

Person(const char* name, int age) {
This->goobj_ = person_new((char*)name, age);
}
~Person() {
Person_delete(this->goobj_);
}

Void Set(char* name, int age) {
Person_set(this->goobj_, name, age);
}
Char* GetName(char* buf, int size) {
Return person_get_name(this->goobj_ buf, size);
}
Int GetAge() {
Return person_get_age(this->goobj_);
}
}
```

After packaging, we can use it like a normal C++ class:

```c++
#include "person.h"

#include <stdio.h>

Int main() {
Auto p = new Person("gopher", 10);

Char buf[64];
Char* name = p->GetName(buf, sizeof(buf)-1);
Int age = p->GetAge();

Printf("%s, %d years old.\n", name, age);
Delete p;

Return 0;
}
```

### 2.8.2.4 Packaging C++ Object Improvements

In the previous implementation of encapsulating C++ objects, each time you create a Person instance via new, you need to do two memory allocations: once for the C++ version of Person, and once again for the Go language version of Person. In fact, the C++ version of Person has only one id of person_handle_t type, which is used to map Go objects. We can use person_handle_t directly in the C++ object.

The following improved packaging methods:

```c++
Extern "C" {
#include "./person_capi.h"
}

Struct Person {
Static Person* New(const char* name, int age) {
Return (Person*)person_new((char*)name, age);
}
Void Delete() {
Person_delete(person_handle_t(this));
}

Void Set(char* name, int age) {
Person_set(person_handle_t(this), name, age);
}
Char* GetName(char* buf, int size) {
Return person_get_name(person_handle_t(this), buf, size);
}
Int GetAge() {
Return person_get_age(person_handle_t(this));
}
};
```

We added a new static member function to the Person class to create a new Person instance. In the New function, the Person instance is created by calling person_new, and the id of the `person_handle_t` type is returned. We cast it as a pointer to the `Person*` type. In other member functions, we reverse the transformation of the this pointer to the `person_handle_t` type, and then call the corresponding function through the C interface.

At this point, we have reached the goal of exporting the Go object as a C interface, and then re-packaging it as a C++ object based on the C interface.

## 2.8.3 Completely liberating C++'s this pointer

Familiarity with the usage of the Go language will reveal that methods in the Go language are bound to types. For example, if we define a new Int type based on int, we can have our own method:

```go
Type Int int

Func (p Int) Twice() int {
Return int(p)*2
}

Func main() {
Var x = Int(42)
fmt.Println(int(x))
fmt.Println(x.Twice())
}
```

This allows you to freely switch int and Int types to use variables without changing the underlying memory structure of the original data.

To achieve similar features in C++, the following implementations are generally used:

```c++
Class Int {
Int v_;

Int(v int) { this.v_ = v; }
Int Twice() const{ return this.v_*2; }
};

Int main() {
Int v(42);

Printf("%d\n", v); // error
Printf("%d\n", v.Twice());
}
```

The newly wrapped Int class adds the Twice method but loses the right to freely switch back to the int type. At this time, not only printf can not output the value of Int itself, but also lose all the features of the int type operation. This is the evil of the C++ constructor: in exchange for the charity of the class at the cost of losing all of its original features.

The root cause of this problem is the pointer type that is fixed to class in C++. We revisit the essence of this in the Go language:

```go
Func (this Int) Twice() int
Func Int_Twice(this Int) int
```

In Go, the type receiver parameter that has a similar function to this is just a normal function parameter. We can freely choose the value or pointer type.

If you think in terms of C, this is just a pointer to the normal `void*` type, and we can freely convert this to other types.

```c++
Struct Int {
Int Twice() {
Const int* p = (int*)(this);
Return (*p) * 2;
}
};
Int main() {
Int x = 42;
Printf("%d\n", x);
Printf("%d\n", ((Int*)(&x))->Twice());
Return 0;
}
```

This way we can construct an Int object by forcing the int type pointer to an Int type pointer instead of the default constructor.
Inside the Twice function, by rotating the this pointer back to the int pointer in the opposite operation, the original int type value can be parsed.
At this time, the Int type is just a shell at compile time and does not take up extra space at runtime.

Therefore, the C++ method can also be used for ordinary non-class types. C++ to ordinary member functions can also be bound to types.
Only pure virtual methods are bound to objects, and that is the interface.