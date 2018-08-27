# 2.6 实战: 封装qsort

qsort快速排序函数是C语言的高阶函数，支持用于自定义排序比较函数，可以对任意类型的数组进行排序。本节我们尝试基于C语言的qsort函数封装一个Go语言版本的qsort函数。

## 2.6.1 认识qsort函数

qsort快速排序函数有`<stdlib.h>`标准库提供，函数的声明如下：

```c
void qsort(
	void* base, size_t num, size_t size,
	int (*cmp)(const void*, const void*)
);
```

其中base参数是要排序数组的首个元素的地址，num是数组中元素的个数，size是数组中每个元素的大小。最关键是cmp比较函数，用于对数组中任意两个元素进行排序。cmp排序函数的两个指针参数分别是要比较的两个元素的地址，如果第一个参数对应元素大于第二个参数对应的元素将返回结果大于0，如果两个元素相等则返回0，如果第一个元素小于第二个元素则返回结果小于0。

下面的例子是用C语言的qsort对一个int类型的数组进行排序：

```c
#include <stdio.h>
#include <stdlib.h>

#define DIM(x) (sizeof(x)/sizeof((x)[0]))

static int cmp(const void* a, const void* b) {
	const int* pa = (int*)a;
	const int* pb = (int*)b;
	return *pa - *pb;
}

int main() {
	int values[] = { 42, 8, 109, 97, 23, 25 };
	int i;

	qsort(values, DIM(values), sizeof(values[0]), cmp);

	for(i = 0; i < DIM(values); i++) {
		printf ("%d ",values[i]);
	}
	return 0;
}
```

其中`DIM(values)`宏用于计算数组元素的个数，`sizeof(values[0])`用于计算数组元素的大小。
cmp是用于排序时比较两个元素大小的回调函数。为了避免对全局名字空间的污染，我们将cmp回调函数定义为仅当前文件内可访问的静态函数。

## 2.6.2 将qsort函数从Go包导出

为了方便Go语言的非CGO用户使用qsort函数，我们需要将C语言的qsort函数包装为一个外部可以访问的Go函数。

用Go语言将qsort函数重新包装为`qsort.Sort`函数：

```go
package qsort

//typedef int (*qsort_cmp_func_t)(const void* a, const void* b);
import "C"
import "unsafe"

func Sort(
	base unsafe.Pointer, num, size C.size_t,
	cmp C.qsort_cmp_func_t,
) {
	C.qsort(base, num, size, cmp)
}
```

因为Go语言的CGO语言不好直接表达C语言的函数类型，因此在C语言空间将比较函数类型重新定义为一个`qsort_cmp_func_t`类型。

虽然Sort函数已经导出了，但是对于qsort包之外的用户依然不能直接使用该函数——Sort函数的参数还包含了虚拟的C包提供的类型。
在CGO的内部机制一节中我们已经提过，虚拟的C包下的任何名称其实都会被映射为包内的私有名字。比如`C.size_t`会被展开为`_Ctype_size_t`，`C.qsort_cmp_func_t`类型会被展开为`_Ctype_qsort_cmp_func_t`。

被CGO处理后的Sort函数的类型如下：

```go
func Sort(
	base unsafe.Pointer, num, size _Ctype_size_t,
	cmp _Ctype_qsort_cmp_func_t,
)
```

这样将会导致包外部用于无法构造`_Ctype_size_t`和`_Ctype_qsort_cmp_func_t`类型的参数而无法使用Sort函数。因此，导出的Sort函数的参数和返回值要避免对虚拟C包的依赖。

重新调整Sort函数的参数类型和实现如下：

```go
/*
#include <stdlib.h>

typedef int (*qsort_cmp_func_t)(const void* a, const void* b);
*/
import "C"
import "unsafe"

type CompareFunc C.qsort_cmp_func_t

func Sort(base unsafe.Pointer, num, size int, cmp CompareFunc) {
	C.qsort(base, C.size_t(num), C.size_t(size), C.qsort_cmp_func_t(cmp))
}
```

我们将虚拟C包中的类型通过Go语言类型代替，在内部调用C函数时重新转型为C函数需要的类型。因此外部用户将不再依赖qsort包内的虚拟C包。

以下代码展示的Sort函数的使用方式：

```go
package main

//extern int go_qsort_compare(void* a, void* b);
import "C"

import (
	"fmt"
	"unsafe"

	qsort "."
)

//export go_qsort_compare
func go_qsort_compare(a, b unsafe.Pointer) C.int {
	pa, pb := (*C.int)(a), (*C.int)(b)
	return C.int(*pa - *pb)
}

func main() {
	values := []int32{42, 9, 101, 95, 27, 25}

	qsort.Sort(unsafe.Pointer(&values[0]),
		len(values), int(unsafe.Sizeof(values[0])),
		qsort.CompareFunc(C.go_qsort_compare),
	)
	fmt.Println(values)
}
```

为了使用Sort函数，我们需要将Go语言的切片取首地址、元素个数、元素大小等信息作为调用参数，同时还需要提供一个C语言规格的比较函数。
其中go_qsort_compare是用Go语言实现的，并导出到C语言空间的函数，用于qsort排序时的比较函数。

目前已经实现了对C语言的qsort初步包装，并且可以通过包的方式被其它用户使用。但是`qsort.Sort`函数已经有很多不便使用之处：用户要提供C语言的比较函数，这对许多Go语言用户是一个挑战。下一步我们将继续改进qsort函数的包装函数，尝试通过闭包函数代替C语言的比较函数。

消除用户对CGO代码的直接依赖。

## 2.6.3 改进：闭包函数作为比较函数

在改进之前我们先回顾下Go语言sort包自带的排序函数的接口：

```go
func Slice(slice interface{}, less func(i, j int) bool)
```

标准库的sort.Slice因为支持通过闭包函数指定比较函数，对切片的排序非常简单：

```go
import "sort"

func main() {
	values := []int32{42, 9, 101, 95, 27, 25}

	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	fmt.Println(values)
}
```

我们也尝试将C语言的qsort函数包装为以下格式的Go语言函数：

```go
package qsort

func Sort(base unsafe.Pointer, num, size int, cmp func(a, b unsafe.Pointer) int)
```

闭包函数无法导出为C语言函数，因此无法直接将闭包函数传入C语言的qsort函数。
为此我们可以用Go构造一个可以导出为C语言的代理函数，然后通过一个全局变量临时保存当前的闭包比较函数。

代码如下：

```go
var go_qsort_compare_info struct {
	fn func(a, b unsafe.Pointer) int
	sync.Mutex
}

//export _cgo_qsort_compare
func _cgo_qsort_compare(a, b unsafe.Pointer) C.int {
	return C.int(go_qsort_compare_info.fn(a, b))
}
```

其中导出的C语言函数`_cgo_qsort_compare`是公用的qsort比较函数，内部通过`go_qsort_compare_info.fn`来调用当前的闭包比较函数。

新的Sort包装函数实现如下：

```go
/*
#include <stdlib.h>

typedef int (*qsort_cmp_func_t)(const void* a, const void* b);
extern int _cgo_qsort_compare(void* a, void* b);
*/
import "C"

func Sort(base unsafe.Pointer, num, size int, cmp func(a, b unsafe.Pointer) int) {
	go_qsort_compare_info.Lock()
	defer go_qsort_compare_info.Unlock()

	go_qsort_compare_info.fn = cmp

	C.qsort(base, C.size_t(num), C.size_t(size),
		C.qsort_cmp_func_t(C._cgo_qsort_compare),
	)
}
```

每次排序前，对全局的go_qsort_compare_info变量加锁，同时将当前的闭包函数保存到全局变量，然后调用C语言的qsort函数。

基于新包装的函数，我们可以简化之前的排序代码：

```go
func main() {
	values := []int32{42, 9, 101, 95, 27, 25}

	qsort.Sort(unsafe.Pointer(&values[0]), len(values), int(unsafe.Sizeof(values[0])),
		func(a, b unsafe.Pointer) int {
			pa, pb := (*int32)(a), (*int32)(b)
			return int(*pa - *pb)
		},
	)

	fmt.Println(values)
}
```

现在排序不再需要通过CGO实现C语言版本的比较函数了，可以传入Go语言闭包函数作为比较函数。
但是导入的排序函数依然依赖unsafe包，这是违背Go语言编程习惯的。

## 2.6.4 改进：消除用户对unsafe包的依赖

前一个版本的qsort.Sort包装函数已经比最初的C语言版本的qsort易用很多，但是依然保留了很多C语言底层数据结构的细节。
现在我们将继续改进包装函数，尝试消除对unsafe包的依赖，并实现一个类似标准库中sort.Slice的排序函数。

新的包装函数声明如下：

```go
package qsort

func Slice(slice interface{}, less func(a, b int) bool)
```

首先，我们将slice作为接口类型参数传入，这样可以适配不同的切片类型。
然后切片的首个元素的地址、元素个数和元素大小可以通过reflect反射包从切片中获取。

为了保存必要的排序上下文信息，我们需要在全局包变量增加要排序数组的地址、元素个数和元素大小等信息，比较函数改为less：

```go
var go_qsort_compare_info struct {
	base     unsafe.Pointer
	elemnum  int
	elemsize int
	less     func(a, b int) bool
	sync.Mutex
}
```

同样比较函数需要根据元素指针、排序数组的开始地址和元素的大小计算出元素对应数组的索引下标，
然后根据less函数的比较结果返回qsort函数需要格式的比较结果。

```go
//export _cgo_qsort_compare
func _cgo_qsort_compare(a, b unsafe.Pointer) C.int {
	var (
		// array memory is locked
		base     = uintptr(go_qsort_compare_info.base)
		elemsize = uintptr(go_qsort_compare_info.elemsize)
	)

	i := int((uintptr(a) - base) / elemsize)
	j := int((uintptr(b) - base) / elemsize)

	switch {
	case go_qsort_compare_info.less(i, j): // v[i] < v[j]
		return -1
	case go_qsort_compare_info.less(j, i): // v[i] > v[j]
		return +1
	default:
		return 0
	}
}
```

新的Slice函数的实现如下：

```go

func Slice(slice interface{}, less func(a, b int) bool) {
	sv := reflect.ValueOf(slice)
	if sv.Kind() != reflect.Slice {
		panic(fmt.Sprintf("qsort called with non-slice value of type %T", slice))
	}
	if sv.Len() == 0 {
		return
	}

	go_qsort_compare_info.Lock()
	defer go_qsort_compare_info.Unlock()

	defer func() {
		go_qsort_compare_info.base = nil
		go_qsort_compare_info.elemnum = 0
		go_qsort_compare_info.elemsize = 0
		go_qsort_compare_info.less = nil
	}()

	// baseMem = unsafe.Pointer(sv.Index(0).Addr().Pointer())
	// baseMem maybe moved, so must saved after call C.fn
	go_qsort_compare_info.base = unsafe.Pointer(sv.Index(0).Addr().Pointer())
	go_qsort_compare_info.elemnum = sv.Len()
	go_qsort_compare_info.elemsize = int(sv.Type().Elem().Size())
	go_qsort_compare_info.less = less

	C.qsort(
		go_qsort_compare_info.base,
		C.size_t(go_qsort_compare_info.elemnum),
		C.size_t(go_qsort_compare_info.elemsize),
		C.qsort_cmp_func_t(C._cgo_qsort_compare),
	)
}
```

首先需要判断传入的接口类型必须是切片类型。然后通过反射获取qsort函数需要的切片信息，并调用C语言的qsort函数。

基于新包装的函数我们可以采用和标准库相似的方式排序切片：

```go
import (
	"fmt"

	qsort "."
)

func main() {
	values := []int64{42, 9, 101, 95, 27, 25}

	qsort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	fmt.Println(values)
}
```

为了避免在排序过程中，排序数组的上下文信息`go_qsort_compare_info`被修改，我们进行了全局加锁。
因此目前版本的qsort.Slice函数是无法并发执行的，读者可以自己尝试改进这个限制。

