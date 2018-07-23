package main

/*
#include <stdint.h>

int64_t myadd(int64_t a, int64_t b) {
	return a+b;
}
*/
import "C"

import (
	"asmpkg"
	"fmt"
	"runtime"
	"unsafe"
)

func main() {
	if runtime.GOOS == "windows" {
		fmt.Println(asmpkg.CallCAdd_Win64_ABI(
			uintptr(unsafe.Pointer(C.myadd)),
			123, 456,
		))
	} else {
		fmt.Println(asmpkg.CallCAdd_SystemV_ABI(
			uintptr(unsafe.Pointer(C.myadd)),
			123, 456,
		))
	}
}
