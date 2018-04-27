// Copyright © 2018 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

/*
#include <stdlib.h>

static int compare(const void* a, const void* b) {
	const int* pa = (int*)a;
	const int* pb = (int*)b;
	return *pa - *pb;
}

static void qsort_proxy(int* values, size_t len, size_t elemsize) {
	qsort(values, len, sizeof(values[0]), compare);
}
*/
import "C"

import "unsafe"
import "fmt"

func main() {
	values := []int32{42, 9, 101, 95, 27, 25}

	C.qsort_proxy(
		(*C.int)(unsafe.Pointer(&values[0])),
		C.size_t(len(values)),
		C.size_t(unsafe.Sizeof(values[0])),
	)

	fmt.Println(values)
}
