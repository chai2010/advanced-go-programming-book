# 2.6 Actual combat: encapsulation qsort

The qsort quick sort function is a high-order function of the C language. It supports the use of custom sort comparison functions and can sort any type of array. In this section we try to encapsulate a Go language version of the qsort function based on the C language qsort function.

## 2.6.1 Understanding the qsort function

The qsort quick sort function is provided by the `<stdlib.h>` standard library. The function declaration is as follows:

```c
Void qsort(
Void* base, size_t num, size_t size,
Int (*cmp)(const void*, const void*)
);
```

The base parameter is the address of the first element of the array to be sorted, num is the number of elements in the array, and size is the size of each element in the array. The key is the cmp comparison function, which is used to sort any two elements in the array. The two pointer parameters of the cmp sort function are the addresses of the two elements to be compared. If the corresponding element of the first parameter is larger than the corresponding element of the second parameter, the result is greater than 0. If the two elements are equal, 0 is returned. The first element is less than the second element and returns a result less than zero.

The following example sorts an array of type int with qsort in C:

```c
#include <stdio.h>
#include <stdlib.h>

#define DIM(x) (sizeof(x)/sizeof((x)[0]))

Static int cmp(const void* a, const void* b) {
Const int* pa = (int*)a;
Const int* pb = (int*)b;
Return *pa - *pb;
}

Int main() {
Int values[] = { 42, 8, 109, 97, 23, 25 };
Int i;

Qsort(values, DIM(values), sizeof(values[0]), cmp);

For(i = 0; i < DIM(values); i++) {
Printf ("%d ", valuess[i]);
}
Return 0;
}
```

The `DIM(values)` macro is used to calculate the number of array elements, and `sizeof(values[0])` is used to calculate the size of array elements.
Cmp is a callback function that compares the size of two elements when sorting. To avoid contamination of the global namespace, we define the cmp callback function as a static function that is only accessible within the current file.

## 2.6.2 Exporting the qsort function from the Go package

In order to facilitate the non-CGO user of the Go language to use the qsort function, we need to wrap the C language qsort function as an externally accessible Go function.

Repack the qsort function as a `qsort.Sort` function in Go:

```go
Package qsort

//typedef int (*qsort_cmp_func_t)(const void* a, const void* b);
Import "C"
Import "unsafe"

Func Sort(
Base unsafe.Pointer, num, size C.size_t,
Cmp C.qsort_cmp_func_t,
) {
C.qsort(base, num, size, cmp)
}
```

Because the Go language CGO language does not directly express the C language function type, the comparison function type is redefined as a `qsort_cmp_func_t` type in the C language space.

Although the Sort function has been exported, the function is not directly available to users outside the qsort package. The parameters of the Sort function also contain the types provided by the virtual C package.
As we mentioned in the CGO Internal Mechanisms section, any name under the virtual C package will actually be mapped to a private name within the package. For example, `C.size_t` will be expanded to `_Ctype_size_t`, and the `C.qsort_cmp_func_t` type will be expanded to `_Ctype_qsort_cmp_func_t`.

The types of Sort functions processed by CGO are as follows:

```go
Func Sort(
Base unsafe.Pointer, num, size _Ctype_size_t,
Cmp _Ctype_qsort_cmp_func_t,
)
```

This will cause the package to be used externally for parameters that cannot construct `_Ctype_size_t` and `_Ctype_qsort_cmp_func_t` types and cannot use the Sort function. Therefore, the parameters and return values ​​of the exported Sort function should avoid dependencies on the virtual C package.

Re-adjust the parameter type and implementation of the Sort function as follows:

```go
/*
#include <stdlib.h>

Typedef int (*qsort_cmp_func_t)(const void* a, const void* b);
*/
Import "C"
Import "unsafe"

Type CompareFunc C.qsort_cmp_func_t

Func Sort(base unsafe.Pointer, num, size int, cmp CompareFunc) {
C.qsort(base, C.size_t(num), C.size_t(size), C.qsort_cmp_func_t(cmp))
}
```

We replace the type in the virtual C package with the Go language type, and re-transform to the type required by the C function when calling the C function internally. Therefore, external users will no longer rely on virtual C packages in the qsort package.

The following code shows how to use the Sort function:

```go
Package main

//extern int go_qsort_compare(void* a, void* b);
Import "C"

Import (
"fmt"
"unsafe"

Qsort "."
)

//export go_qsort_compare
Func go_qsort_compare(a, b unsafe.Pointer) C.int {
Pa, pb := (*C.int)(a), (*C.int)(b)
Return C.int(*pa - *pb)
}

Func main() {
Values ​​:= []int32{42, 9, 101, 95, 27, 25}

qsort.Sort(unsafe.Pointer(&values[0]),
Len(values), int(unsafe.Sizeof(values[0])),
qsort.CompareFunc(C.go_qsort_compare),
)
fmt.Println(values)
}
```

In order to use the Sort function, we need to take the information of the first address, the number of elements, the size of the element in the Go language as the calling parameter, and also provide a comparison function of the C language specification.
Where go_qsort_compare is implemented in Go language and exported to the C language space function for the comparison function of qsort sorting.

The initial packaging of the qsort for the C language has been implemented and can be used by other users through the package. But the `qsort.Sort` function has a lot of inconveniences: users need to provide C language comparison functions, which is a challenge for many Go language users. Next we will continue to improve the wrapper function of the qsort function, trying to replace the C language comparison function with the closure function.

Eliminate users' direct dependence on CGO code.

## 2.6.3 Improvement: Closure function as comparison function

Before the improvement, we will review the interface of the sort function that comes with the Go language sort package:

```go
Func Slice(slice interface{}, less func(i, j int) bool)
```

The sort.Slice of the standard library is very simple to sort the slices because it supports the comparison function specified by the closure function:

```go
Import "sort"

Func main() {
Values ​​:= []int32{42, 9, 101, 95, 27, 25}

sort.Slice(values, func(i, j int) bool {
Return values[i] < values[j]
})

fmt.Println(values)
}
```

We also try to wrap the C language qsort function as a Go language function in the following format:

```go
Package qsort

Func Sort(base unsafe.Pointer, num, size int, cmp func(a, b unsafe.Pointer) int)
```

The closure function cannot be exported as a C language function, so the closure function cannot be directly passed to the C language qsort function.
To do this, we can construct a proxy function that can be exported to C using Go, and then temporarily save the current closure comparison function through a global variable.

code show as below:

```go
Var go_qsort_compare_info struct {
Fn func(a, b unsafe.Pointer) int
sync.Mutex
}

//export _cgo_qsort_compare
Func _cgo_qsort_compare(a, b unsafe.Pointer) C.int {
Return C.int(go_qsort_compare_info.fn(a, b))
}
```

The exported C language function `_cgo_qsort_compare` is a public qsort comparison function, and the current closure comparison function is called internally by `go_qsort_compare_info.fn`.

The new Sort wrapper function is implemented as follows:

```go
/*
#include <stdlib.h>

Typedef int (*qsort_cmp_func_t)(const void* a, const void* b);
Extern int _cgo_qsort_compare(void* a, void* b);
*/
Import "C"

Func Sort(base unsafe.Pointer, num, size int, cmp func(a, b unsafe.Pointer) int) {
go_qsort_compare_info.Lock()
Defer go_qsort_compare_info.Unlock()

Go_qsort_compare_info.fn = cmp

C.qsort(base, C.size_t(num), C.size_t(size),
C.qsort_cmp_func_t(C._cgo_qsort_compare),
)
}
```

Before each sorting, lock the global go_qsort_compare_info variable, save the current closure function to the global variable, and then call the C language qsort function.

Based on the newly wrapped function, we can simplify the previous sorting code:

```go
Func main() {
Values ​​:= []int32{42, 9, 101, 95, 27, 25}

qsort.Sort(unsafe.Pointer(&values[0]), len(values), int(unsafe.Sizeof(values[0])),
Func(a, b unsafe.Pointer) int {
Pa, pb := (*int32)(a), (*int32)(b)
Return int(*pa - *pb)
},
)

fmt.Println(values)
}
```

Now sorting no longer needs to implement the C language version of the comparison function through CGO, you can pass the Go language closure function as a comparison function.
But the imported sort function still relies on the unsafe package, which is against the Go language programming habits.

## 2.6.4 Improvement: Eliminate user dependence on unsafe packages

The previous version of the qsort.Sort wrapper function has been much easier to use than the original C language version of qsort, but still retains a lot of the details of the C language underlying data structure.
Now we will continue to improveWrap the function, try to eliminate the dependency on the unsafe package, and implement a sort function similar to the sort.Slice in the standard library.

The new wrapper function is declared as follows:

```go
Package qsort

Func Slice(slice interface{}, less func(a, b int) bool)
```

First, we pass the slice as an interface type parameter so that we can adapt to different slice types.
Then the address, the number of elements, and the element size of the first element of the slice can be obtained from the slice by the reflect reflection packet.

In order to save the necessary sort context information, we need to increase the address of the array to be sorted, the number of elements and the size of the element in the global package variable. The comparison function is changed to less:

```go
Var go_qsort_compare_info struct {
Base unsafe.Pointer
Elemnum int
Elemsize int
Less func(a, b int) bool
sync.Mutex
}
```

The same comparison function needs to calculate the index subscript of the corresponding array of elements according to the element pointer, the start address of the sorted array, and the size of the element.
Then according to the comparison result of the less function, the comparison result of the format required by the qsort function is returned.

```go
//export _cgo_qsort_compare
Func _cgo_qsort_compare(a, b unsafe.Pointer) C.int {
Var (
// array memory is locked
Base = uintptr(go_qsort_compare_info.base)
Elemsize = uintptr(go_qsort_compare_info.elemsize)
)

i := int((uintptr(a) - base) / elemsize)
j := int((uintptr(b) - base) / elemsize)

Switch {
Case go_qsort_compare_info.less(i, j): // v[i] < v[j]
Return -1
Case go_qsort_compare_info.less(j, i): // v[i] > v[j]
Return +1
Default:
Return 0
}
}
```

The implementation of the new Slice function is as follows:

```go

Func Slice(slice interface{}, less func(a, b int) bool) {
Sv := reflect.ValueOf(slice)
If sv.Kind() != reflect.Slice {
Panic(fmt.Sprintf("qsort called with non-slice value of type %T", slice))
}
If sv.Len() == 0 {
Return
}

go_qsort_compare_info.Lock()
Defer go_qsort_compare_info.Unlock()

Defer func() {
Go_qsort_compare_info.base = nil
Go_qsort_compare_info.elemnum = 0
Go_qsort_compare_info.elemsize = 0
Go_qsort_compare_info.less = nil
}()

// baseMem = unsafe.Pointer(sv.Index(0).Addr().Pointer())
// baseMem maybe moved, so must saved after call C.fn
Go_qsort_compare_info.base = unsafe.Pointer(sv.Index(0).Addr().Pointer())
Go_qsort_compare_info.elemnum = sv.Len()
Go_qsort_compare_info.elemsize = int(sv.Type().Elem().Size())
Go_qsort_compare_info.less = less

C.qsort(
Go_qsort_compare_info.base,
C.size_t(go_qsort_compare_info.elemnum),
C.size_t(go_qsort_compare_info.elemsize),
C.qsort_cmp_func_t(C._cgo_qsort_compare),
)
}
```

First you need to determine that the type of interface passed in must be a slice type. Then get the slice information needed by the qsort function through reflection and call the qsort function of C language.

Based on the newly wrapped function we can sort the slices in a similar way to the standard library:

```go
Import (
"fmt"

Qsort "."
)

Func main() {
Values ​​:= []int64{42, 9, 101, 95, 27, 25}

qsort.Slice(values, func(i, j int) bool {
Return values[i] < values[j]
})

fmt.Println(values)
}
```

In order to avoid the sorting array's context information `go_qsort_compare_info` being modified during the sorting process, we have global locking.
Therefore, the current version of the qsort.Slice function cannot be executed concurrently, and the reader can try to improve this limitation by himself.