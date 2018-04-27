// Copyright © 2018 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

import (
	"fmt"
)

/*
#include <stdlib.h>

typedef int (*qsort_cmp_func_t)(const void* a, const void* b);
extern int go_qsort_compare(void* a, void* b);
*/
import "C"
import (
	"sync"
	"unsafe"
)

func main() {
	values := []int32{42, 9, 101, 95, 27, 25}

	go_qsort_compare_info.Lock()
	defer go_qsort_compare_info.Unlock()

	go_qsort_compare_info.fn = func(a, b unsafe.Pointer) C.int {
		pa := (*C.int)(a)
		pb := (*C.int)(b)
		return C.int(*pa - *pb)
	}

	C.qsort(unsafe.Pointer(&values[0]),
		C.size_t(len(values)), C.size_t(unsafe.Sizeof(values[0])),
		(C.qsort_cmp_func_t)(unsafe.Pointer(C.go_qsort_compare)),
	)

	fmt.Println(values)
}

//export go_qsort_compare
func go_qsort_compare(a, b unsafe.Pointer) C.int {
	return go_qsort_compare_info.fn(a, b)
}

var go_qsort_compare_info struct {
	fn func(a, b unsafe.Pointer) C.int
	sync.RWMutex
}
