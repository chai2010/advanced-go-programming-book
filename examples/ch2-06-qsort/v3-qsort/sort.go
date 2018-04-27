// Copyright © 2018 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package qsort

/*
#include <stdlib.h>

typedef int (*qsort_cmp_func_t)(const void* a, const void* b);
extern int _cgo_qsort_compare(void* a, void* b);
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

//export _cgo_qsort_compare
func _cgo_qsort_compare(a, b unsafe.Pointer) C.int {
	return C.int(go_qsort_compare_info.fn(a, b))
}

func Sort(base unsafe.Pointer, num, size int, cmp func(a, b unsafe.Pointer) int) {
	go_qsort_compare_info.Lock()
	defer go_qsort_compare_info.Unlock()

	go_qsort_compare_info.fn = cmp

	C.qsort(base, C.size_t(num), C.size_t(size),
		C.qsort_cmp_func_t(C._cgo_qsort_compare),
	)
}
