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

var go_qsort_compare_info struct {
	fn func(a, b unsafe.Pointer) int
	sync.Mutex
}

//export go_qsort_compare
func go_qsort_compare(a, b unsafe.Pointer) C.int {
	return C.int(go_qsort_compare_info.fn(a, b))
}

func qsort(base unsafe.Pointer, num, size int, cmp func(a, b unsafe.Pointer) int) {
	go_qsort_compare_info.Lock()
	defer go_qsort_compare_info.Unlock()

	go_qsort_compare_info.fn = cmp

	C.qsort(base, C.size_t(num), C.size_t(size),
		C.qsort_cmp_func_t(C.go_qsort_compare),
	)
}

func main() {
	values := []int32{42, 9, 101, 95, 27, 25}

	qsort(unsafe.Pointer(&values[0]), len(values), int(unsafe.Sizeof(values[0])),
		func(a, b unsafe.Pointer) int {
			pa, pb := (*C.int)(a), (*C.int)(b)
			return int(*pa - *pb)
		},
	)

	fmt.Println(values)
}
