// Copyright © 2018 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

/*
#include <stdlib.h>

typedef int (*qsort_cmp_func_t)(const void* a, const void* b);

extern int  go_qsort_compare(void* a, void* b);
extern void go_qsort_compare_save_base(void* base);

static void
qsort_proxy(
	void* base, size_t num, size_t size,
	int (*compare)(const void* a, const void* b)
) {
	go_qsort_compare_save_base(base);
	qsort(base, num, size, compare);
}
*/
import "C"

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

func main() {
	values := []int64{42, 9, 101, 95, 27, 25}

	fmt.Println(values)
	defer fmt.Println(values)

	qsort(values, func(i, j int) int {
		return int(values[i] - values[j])
	})
}

func qsort(slice interface{}, fn func(a, b int) int) {
	sv := reflect.ValueOf(slice)
	if sv.Kind() != reflect.Slice {
		panic(fmt.Sprintf("qsort called with non-slice value of type %T", slice))
	}
	if sv.Len() == 0 {
		return
	}

	go_qsort_compare_info.Lock()
	defer go_qsort_compare_info.Unlock()

	// baseMem = unsafe.Pointer(sv.Index(0).Addr().Pointer())
	// baseMem maybe moved, so must saved after call C.fn
	go_qsort_compare_info.elemsize = sv.Type().Elem().Size()
	go_qsort_compare_info.fn = fn

	C.qsort_proxy(
		unsafe.Pointer(unsafe.Pointer(sv.Index(0).Addr().Pointer())),
		C.size_t(sv.Len()),
		C.size_t(sv.Type().Elem().Size()),
		(C.qsort_cmp_func_t)(unsafe.Pointer(C.go_qsort_compare)),
	)
}

//export go_qsort_compare
func go_qsort_compare(a, b unsafe.Pointer) C.int {
	var (
		// array memory is locked
		base     = go_qsort_compare_info.base
		elemsize = go_qsort_compare_info.elemsize
	)

	i := int((uintptr(a) - base) / elemsize)
	j := int((uintptr(b) - base) / elemsize)

	return C.int(go_qsort_compare_info.fn(i, j))
}

//export go_qsort_compare_save_base
func go_qsort_compare_save_base(base unsafe.Pointer) {
	go_qsort_compare_info.base = uintptr(base)
}

var go_qsort_compare_info struct {
	base     uintptr
	elemsize uintptr
	fn       func(a, b int) int
	sync.RWMutex
}
