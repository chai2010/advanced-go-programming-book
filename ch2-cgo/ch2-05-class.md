# 2.5. C++ 类包装(Doing)

CGO是C语言和Go语言之间的桥梁，原则上无法直接支持C++的类。CGO不支持C++语法的根本原因是C++至今为止还没有一个二进制接口规范(ABI)。一个C++类的构造函数在编译为目标文件时如何生成链接符号名称到方法在不太平台甚至是C++到不同版本之间都是不一样的。但是C++的最大优势是兼容C语言，我们可以通过增加一组C语言函数接口作为C++类和CGO之间的桥梁，这样就可以间接地实现C++和Go之间的互联。当然，因为CGO只支持C语言中值类型的数据类型，我们是无法直接使用C++的引用参数等特性的。

## C++ 类到 Go 语言对象

要实现C++类到Go语言对象的包装需要经过以下几个步骤：首先是用纯C函数接口包装该C++类；其次是通过CGO将纯C函数接口映射到Go函数；最后是做一个Go包装对象，将C++类到方法用Go对象的方法实现。

### 准备一个 C++ 类

为了演示简单，我们基于`std::string`做一个最简单的缓存对象MyBuffer。除了构造函数和析构函数之外，只有两个成员函数分别是返回底层的数据指针和缓存的大小。因为是二进制缓存，我们可以在里面中放置任意数据。

```c++
#include <string>

class MyBuffer {
	std::string* s_;

	MyBuffer(int size) {
		this->s_ = new std::string(n, char('\0'));
	}
	~MyBuffer() {
		delete this->s_;
	}

	int Size() const {
		return this->s_->size();
	}
	const char* Data() {
		return this->s_->data();
	}
};
```

我们在构造函数中指定缓存的大小并分配空间，在使用完之后通过哦析构函数释放内部分配到内存空间。下面是简单的使用方式：

```c++
int main() {
	auto pBuf = new MyBuffer(1024);

	auto data = pBuf->Data();
	auto size = pBuf->Size();

	delete pBuf;
}
```

为了方便向C语言接口过度，我们故意没有定义C++的拷贝构造函数。我们必须以new和delete来分配和释放缓存对象，而不能以值风格的方式来使用。

### 用纯C函数接口封装 C++ 类

如果要将上面的C++类用C语言函数接口封装，我们可以从使用方式入手。我们可以将new和delete映射为C语言函数，将对象的方法也映射为C语言函数。

在C语言中我们期望MyBuffer类可以这样使用：

```c
int main() {
	MyBuffer* pBuf = NewMyBuffer(1024);

	char* data = MyBuffer_Data(pBuf);
	auto size = MyBuffer_Size(pBuf);

	DeleteMyBuffer(pBuf);
}
```

先从C语言接口用户的角度思考需要什么样的接口，然后创建 `my_buffer_capi.h` 头文件接口规范：

```c++
// my_buffer_capi.h
typedef struct MyBuffer_T MyBuffer_T;

MyBuffer_T* NewMyBuffer(int size);
void DeleteMyBuffer(MyBuffer_T* p);

char* MyBuffer_Data(MyBuffer_T* p);
int MyBuffer_Size(MyBuffer_T* p);
```

然后就可以基于C++的MyBuffer类定义这些C语言包装函数。我们创建对应的`my_buffer_capi.cc`文件如下：

```c++
// my_buffer_capi.h

extern "C" {
	#incude "./my_buffer_capi.h"
}

struct MyBuffer_T: MyBuffer {};

MyBuffer_T* NewMyBuffer(int size) {
	auto p = new MyBuffer(size);
	return (MyBuffer_T*)(p);
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

因为头文件`my_buffer_capi.h`是用于CGO，必须是采用C语言规范的名字修饰规则。在C++源源文件包含时需要用`extern "C"`语句说明。另外MyBuffer_T的实现只是从MyBuffer继承的类，这样可以简化包装代码的实现。同时，和CGO通信时必须通过`MyBuffer_T`指针，我们无法将具体的实现暴漏给CGO，因为实现中包含了C++特有的语法，CGO无法识别C++特性。

将C++类包装为纯C接口之后，下一步的工作就是将C函数转为Go函数。

### 将纯C接口函数转为Go函数

### 包装为Go对象

## Go 语言对象到 C++ 类

TODO
