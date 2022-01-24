# 2.8 C++ 类包装

CGO 是 C 语言和 Go 语言之间的桥梁，原则上无法直接支持 C++ 的类。CGO 不支持 C++ 语法的根本原因是 C++ 至今为止还没有一个二进制接口规范 (ABI)。一个 C++ 类的构造函数在编译为目标文件时如何生成链接符号名称、方法在不同平台甚至是 C++ 的不同版本之间都是不一样的。但是 C++ 是兼容 C 语言，所以我们可以通过增加一组 C 语言函数接口作为 C++ 类和 CGO 之间的桥梁，这样就可以间接地实现 C++ 和 Go 之间的互联。当然，因为 CGO 只支持 C 语言中值类型的数据类型，所以我们是无法直接使用 C++ 的引用参数等特性的。

## 2.8.1 C++ 类到 Go 语言对象

实现 C++ 类到 Go 语言对象的包装需要经过以下几个步骤：首先是用纯 C 函数接口包装该 C++ 类；其次是通过 CGO 将纯 C 函数接口映射到 Go 函数；最后是做一个 Go 包装对象，将 C++ 类到方法用 Go 对象的方法实现。

### 2.8.1.1 准备一个 C++ 类

为了演示简单，我们基于 `std::string` 做一个最简单的缓存类 MyBuffer。除了构造函数和析构函数之外，只有两个成员函数分别是返回底层的数据指针和缓存的大小。因为是二进制缓存，所以我们可以在里面中放置任意数据。

```c++
// my_buffer.h
#include <string>

struct MyBuffer {
	std::string* s_;

	MyBuffer(int size) {
		this->s_ = new std::string(size, char('\0'));
	}
	~MyBuffer() {
		delete this->s_;
	}

	int Size() const {
		return this->s_->size();
	}
	char* Data() {
		return (char*)this->s_->data();
	}
};
```

我们在构造函数中指定缓存的大小并分配空间，在使用完之后通过析构函数释放内部分配的内存空间。下面是简单的使用方式：

```c++
int main() {
	auto pBuf = new MyBuffer(1024);

	auto data = pBuf->Data();
	auto size = pBuf->Size();

	delete pBuf;
}
```

为了方便向 C 语言接口过渡，在此处我们故意没有定义 C++ 的拷贝构造函数。我们必须以 new 和 delete 来分配和释放缓存对象，而不能以值风格的方式来使用。

### 2.8.1.2 用纯 C 函数接口封装 C++ 类

如果要将上面的 C++ 类用 C 语言函数接口封装，我们可以从使用方式入手。我们可以将 new 和 delete 映射为 C 语言函数，将对象的方法也映射为 C 语言函数。

在 C 语言中我们期望 MyBuffer 类可以这样使用：

```c
int main() {
	MyBuffer* pBuf = NewMyBuffer(1024);

	char* data = MyBuffer_Data(pBuf);
	auto size = MyBuffer_Size(pBuf);

	DeleteMyBuffer(pBuf);
}
```

先从 C 语言接口用户的角度思考需要什么样的接口，然后创建 `my_buffer_capi.h` 头文件接口规范：

```c++
// my_buffer_capi.h
typedef struct MyBuffer_T MyBuffer_T;

MyBuffer_T* NewMyBuffer(int size);
void DeleteMyBuffer(MyBuffer_T* p);

char* MyBuffer_Data(MyBuffer_T* p);
int MyBuffer_Size(MyBuffer_T* p);
```

然后就可以基于 C++ 的 MyBuffer 类定义这些 C 语言包装函数。我们创建对应的 `my_buffer_capi.cc` 文件如下：

```c++
// my_buffer_capi.cc

#include "./my_buffer.h"

extern "C" {
	#include "./my_buffer_capi.h"
}

struct MyBuffer_T: MyBuffer {
	MyBuffer_T(int size): MyBuffer(size) {}
	~MyBuffer_T() {}
};

MyBuffer_T* NewMyBuffer(int size) {
	auto p = new MyBuffer_T(size);
	return p;
}
void DeleteMyBuffer(MyBuffer_T* p) {
	delete p;
}

char* MyBuffer_Data(MyBuffer_T* p) {
	return p->Data();
}
int MyBuffer_Size(MyBuffer_T* p) {
	return p->Size();
}
```

因为头文件 `my_buffer_capi.h` 是用于 CGO，必须是采用 C 语言规范的名字修饰规则。在 C++ 源文件包含时需要用 `extern "C"` 语句说明。另外 MyBuffer_T 的实现只是从 MyBuffer 继承的类，这样可以简化包装代码的实现。同时和 CGO 通信时必须通过 `MyBuffer_T` 指针，我们无法将具体的实现暴露给 CGO，因为实现中包含了 C++ 特有的语法，CGO 无法识别 C++ 特性。

将 C++ 类包装为纯 C 接口之后，下一步的工作就是将 C 函数转为 Go 函数。

### 2.8.1.3 将纯 C 接口函数转为 Go 函数

将纯 C 函数包装为对应的 Go 函数的过程比较简单。需要注意的是，因为我们的包中包含 C++11 的语法，因此需要通过 `#cgo CXXFLAGS: -std=c++11` 打开 C++11 的选项。

```go
// my_buffer_capi.go

package main

/*
#cgo CXXFLAGS: -std=c++11

#include "my_buffer_capi.h"
*/
import "C"

type cgo_MyBuffer_T C.MyBuffer_T

func cgo_NewMyBuffer(size int) *cgo_MyBuffer_T {
	p := C.NewMyBuffer(C.int(size))
	return (*cgo_MyBuffer_T)(p)
}

func cgo_DeleteMyBuffer(p *cgo_MyBuffer_T) {
	C.DeleteMyBuffer((*C.MyBuffer_T)(p))
}

func cgo_MyBuffer_Data(p *cgo_MyBuffer_T) *C.char {
	return C.MyBuffer_Data((*C.MyBuffer_T)(p))
}

func cgo_MyBuffer_Size(p *cgo_MyBuffer_T) C.int {
	return C.MyBuffer_Size((*C.MyBuffer_T)(p))
}
```

为了区分，我们在 Go 中的每个类型和函数名称前面增加了 `cgo_` 前缀，比如 cgo_MyBuffer_T 是对应 C 中的 MyBuffer_T 类型。

为了处理简单，在包装纯 C 函数到 Go 函数时，除了 cgo_MyBuffer_T 类型外，对输入参数和返回值的基础类型，我们依然是用的 C 语言的类型。

### 2.8.1.4 包装为 Go 对象

在将纯 C 接口包装为 Go 函数之后，我们就可以很容易地基于包装的 Go 函数构造出 Go 对象来。因为 cgo_MyBuffer_T 是从 C 语言空间导入的类型，它无法定义自己的方法，因此我们构造了一个新的 MyBuffer 类型，里面的成员持有 cgo_MyBuffer_T 指向的 C 语言缓存对象。

```go
// my_buffer.go

package main

import "unsafe"

type MyBuffer struct {
	cptr *cgo_MyBuffer_T
}

func NewMyBuffer(size int) *MyBuffer {
	return &MyBuffer{
		cptr: cgo_NewMyBuffer(size),
	}
}

func (p *MyBuffer) Delete() {
	cgo_DeleteMyBuffer(p.cptr)
}

func (p *MyBuffer) Data() []byte {
	data := cgo_MyBuffer_Data(p.cptr)
	size := cgo_MyBuffer_Size(p.cptr)
	return ((*[1 << 31]byte)(unsafe.Pointer(data)))[0:int(size):int(size)]
}
```

同时，因为 Go 语言的切片本身含有长度信息，我们将 cgo_MyBuffer_Data 和 cgo_MyBuffer_Size 两个函数合并为 `MyBuffer.Data` 方法，它返回一个对应底层 C 语言缓存空间的切片。

现在我们就可以很容易在 Go 语言中使用包装后的缓存对象了（底层是基于 C++ 的 `std::string` 实现）：

```go
package main

//#include <stdio.h>
import "C"
import "unsafe"

func main() {
	buf := NewMyBuffer(1024)
	defer buf.Delete()

	copy(buf.Data(), []byte("hello\x00"))
	C.puts((*C.char)(unsafe.Pointer(&(buf.Data()[0]))))
}
```

例子中，我们创建了一个 1024 字节大小的缓存，然后通过 copy 函数向缓存填充了一个字符串。为了方便 C 语言字符串函数处理，我们在填充字符串的默认用'\0'表示字符串结束。最后我们直接获取缓存的底层数据指针，用 C 语言的 puts 函数打印缓存的内容。

## 2.8.2 Go 语言对象到 C++ 类

要实现 Go 语言对象到 C++ 类的包装需要经过以下几个步骤：首先是将 Go 对象映射为一个 id；然后基于 id 导出对应的 C 接口函数；最后是基于 C 接口函数包装为 C++ 对象。

### 2.8.2.1 构造一个 Go 对象

为了便于演示，我们用 Go 语言构建了一个 Person 对象，每个 Person 可以有名字和年龄信息：

```go
package main

type Person struct {
	name string
	age  int
}

func NewPerson(name string, age int) *Person {
	return &Person{
		name: name,
		age:  age,
	}
}

func (p *Person) Set(name string, age int) {
	p.name = name
	p.age = age
}

func (p *Person) Get() (name string, age int) {
	return p.name, p.age
}
```

Person 对象如果想要在 C/C++ 中访问，需要通过 cgo 导出 C 接口来访问。

### 2.8.2.2 导出 C 接口

我们前面仿照 C++ 对象到 C 接口的过程，也抽象一组 C 接口描述 Person 对象。创建一个 `person_capi.h` 文件，对应 C 接口规范文件：

```c
// person_capi.h
#include <stdint.h>

typedef uintptr_t person_handle_t;

person_handle_t person_new(char* name, int age);
void person_delete(person_handle_t p);

void person_set(person_handle_t p, char* name, int age);
char* person_get_name(person_handle_t p, char* buf, int size);
int person_get_age(person_handle_t p);
```

然后是在 Go 语言中实现这一组 C 函数。

需要注意的是，通过 CGO 导出 C 函数时，输入参数和返回值类型都不支持 const 修饰，同时也不支持可变参数的函数类型。同时如内存模式一节所述，我们无法在 C/C++ 中直接长期访问 Go 内存对象。因此我们使用前一节所讲述的技术将 Go 对象映射为一个整数 id。

下面是 `person_capi.go` 文件，对应 C 接口函数的实现：

```go
// person_capi.go
package main

//#include "./person_capi.h"
import "C"
import "unsafe"

//export person_new
func person_new(name *C.char, age C.int) C.person_handle_t {
	id := NewObjectId(NewPerson(C.GoString(name), int(age)))
	return C.person_handle_t(id)
}

//export person_delete
func person_delete(h C.person_handle_t) {
	ObjectId(h).Free()
}

//export person_set
func person_set(h C.person_handle_t, name *C.char, age C.int) {
	p := ObjectId(h).Get().(*Person)
	p.Set(C.GoString(name), int(age))
}

//export person_get_name
func person_get_name(h C.person_handle_t, buf *C.char, size C.int) *C.char {
	p := ObjectId(h).Get().(*Person)
	name, _ := p.Get()

	n := int(size) - 1
	bufSlice := ((*[1 << 31]byte)(unsafe.Pointer(buf)))[0:n:n]
	n = copy(bufSlice, []byte(name))
	bufSlice[n] = 0

	return buf
}

//export person_get_age
func person_get_age(h C.person_handle_t) C.int {
	p := ObjectId(h).Get().(*Person)
	_, age := p.Get()
	return C.int(age)
}
```

在创建 Go 对象后，我们通过 NewObjectId 将 Go 对应映射为 id。然后将 id 强制转义为 person_handle_t 类型返回。其它的接口函数则是根据 person_handle_t 所表示的 id，让根据 id 解析出对应的 Go 对象。

### 2.8.2.3 封装 C++ 对象

有了 C 接口之后封装 C++ 对象就比较简单了。常见的做法是新建一个 Person 类，里面包含一个 person_handle_t 类型的成员对应真实的 Go 对象，然后在 Person 类的构造函数中通过 C 接口创建 Go 对象，在析构函数中通过 C 接口释放 Go 对象。下面是采用这种技术的实现：

```c++
extern "C" {
	#include "./person_capi.h"
}

struct Person {
	person_handle_t goobj_;

	Person(const char* name, int age) {
		this->goobj_ = person_new((char*)name, age);
	}
	~Person() {
		person_delete(this->goobj_);
	}

	void Set(char* name, int age) {
		person_set(this->goobj_, name, age);
	}
	char* GetName(char* buf, int size) {
		return person_get_name(this->goobj_ buf, size);
	}
	int GetAge() {
		return person_get_age(this->goobj_);
	}
}
```

包装后我们就可以像普通 C++ 类那样使用了：

```c++
#include "person.h"

#include <stdio.h>

int main() {
	auto p = new Person("gopher", 10);

	char buf[64];
	char* name = p->GetName(buf, sizeof(buf)-1);
	int age = p->GetAge();

	printf("%s, %d years old.\n", name, age);
	delete p;

	return 0;
}
```

### 2.8.2.4 封装 C++ 对象改进

在前面的封装 C++ 对象的实现中，每次通过 new 创建一个 Person 实例需要进行两次内存分配：一次是针对 C++ 版本的 Person，再一次是针对 Go 语言版本的 Person。其实 C++ 版本的 Person 内部只有一个 person_handle_t 类型的 id，用于映射 Go 对象。我们完全可以将 person_handle_t 直接当中 C++ 对象来使用。

下面时改进后的包装方式：

```c++
extern "C" {
	#include "./person_capi.h"
}

struct Person {
	static Person* New(const char* name, int age) {
		return (Person*)person_new((char*)name, age);
	}
	void Delete() {
		person_delete(person_handle_t(this));
	}

	void Set(char* name, int age) {
		person_set(person_handle_t(this), name, age);
	}
	char* GetName(char* buf, int size) {
		return person_get_name(person_handle_t(this), buf, size);
	}
	int GetAge() {
		return person_get_age(person_handle_t(this));
	}
};
```

我们在 Person 类中增加了一个叫 New 静态成员函数，用于创建新的 Person 实例。在 New 函数中通过调用 person_new 来创建 Person 实例，返回的是 `person_handle_t` 类型的 id，我们将其强制转型作为 `Person*` 类型指针返回。在其它的成员函数中，我们通过将 this 指针再反向转型为 `person_handle_t` 类型，然后通过 C 接口调用对应的函数。

到此，我们就达到了将 Go 对象导出为 C 接口，然后基于 C 接口再包装为 C++ 对象以便于使用的目的。

## 2.8.3 彻底解放 C++ 的 this 指针

熟悉 Go 语言的用法会发现 Go 语言中方法是绑定到类型的。比如我们基于 int 定义一个新的 Int 类型，就可以有自己的方法：

```go
type Int int

func (p Int) Twice() int {
	return int(p)*2
}

func main() {
	var x = Int(42)
	fmt.Println(int(x))
	fmt.Println(x.Twice())
}
```

这样就可以在不改变原有数据底层内存结构的前提下，自由切换 int 和 Int 类型来使用变量。

而在 C++ 中要实现类似的特性，一般会采用以下实现：

```c++
class Int {
	int v_;

	Int(v int) { this.v_ = v; }
	int Twice() const{ return this.v_*2;}
};

int main() {
	Int v(42);

	printf("%d\n", v); // error
	printf("%d\n", v.Twice());
}
```

新包装后的 Int 类虽然增加了 Twice 方法，但是失去了自由转回 int 类型的权利。这时候不仅连 printf 都无法输出 Int 本身的值，而且也失去了 int 类型运算的所有特性。这就是 C++ 构造函数的邪恶之处：以失去原有的一切特性的代价换取 class 的施舍。

造成这个问题的根源是 C++ 中 this 被固定为 class 的指针类型了。我们重新回顾下 this 在 Go 语言中的本质：

```go
func (this Int) Twice() int
func Int_Twice(this Int) int
```

在 Go 语言中，和 this 有着相似功能的类型接收者参数其实只是一个普通的函数参数，我们可以自由选择值或指针类型。

如果以 C 语言的角度来思考，this 也只是一个普通的 `void*` 类型的指针，我们可以随意自由地将 this 转换为其它类型。

```c++
struct Int {
	int Twice() {
		const int* p = (int*)(this);
		return (*p) * 2;
	}
};
int main() {
	int x = 42;
	printf("%d\n", x);
	printf("%d\n", ((Int*)(&x))->Twice());
	return 0;
}
```

这样我们就可以通过将 int 类型指针强制转为 Int 类型指针，代替通过默认的构造函数后 new 来构造 Int 对象。
在 Twice 函数的内部，以相反的操作将 this 指针转回 int 类型的指针，就可以解析出原有的 int 类型的值了。
这时候 Int 类型只是编译时的一个壳子，并不会在运行时占用额外的空间。

因此 C++ 的方法其实也可以用于普通非 class 类型，C++ 到普通成员函数其实也是可以绑定到类型的。
只有纯虚方法是绑定到对象，那就是接口。
