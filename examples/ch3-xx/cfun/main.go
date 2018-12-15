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
	if runtime.GOOS == "darwin" {
		asmpkg.SyscallWrite_Darwin(1, "hello syscall!\n")
	}
	if runtime.GOOS == "linux" {
		asmpkg.SyscallWrite_Linux(1, "hello syscall!\n")
	}

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

	var dst = make([]byte, 32)
	var src = []byte("1234567890123456789012345678901234567890")
	asmpkg.CopySlice_AVX2(dst, src, 32)
	fmt.Println(string(dst))
}
