// Copyright © 2018 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

/*
#include <stdlib.h>

typedef int (*qsort_cmp_func_t)(const void* a, const void* b);

// disable static
int qsort_compare_callback(const void* a, const void* b) {
	const int* pa = (int*)a;
	const int* pb = (int*)b;
	return *pa - *pb;
}

static int compare(const void* a, const void* b) {
	const int* pa = (int*)a;
	const int* pb = (int*)b;
	return *pa - *pb;
}
static qsort_cmp_func_t get_compare_ptr() {
	return compare;
}
*/
import "C"

import "unsafe"
import "fmt"

func qsort(base unsafe.Pointer, num, size C.size_t, cmp C.qsort_cmp_func_t) {
	C.qsort(base, num, size, C.qsort_cmp_func_t(cmp))
}

func main() {
	values := []int32{42, 9, 101, 95, 27, 25}

	qsort(unsafe.Pointer(&values[0]),
		C.size_t(len(values)), C.size_t(unsafe.Sizeof(values[0])),
		C.get_compare_ptr(), // static callback must use proxy get
	)
	fmt.Println(values)

	values = []int32{42, 9, 101, 95, 27, 25}

	qsort(unsafe.Pointer(&values[0]),
		C.size_t(len(values)), C.size_t(unsafe.Sizeof(values[0])),
		C.qsort_cmp_func_t(unsafe.Pointer(C.qsort_compare_callback)),
	)
	fmt.Println(values)
}
