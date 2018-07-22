package main

/*
#include <stdio.h>
#include <stdint.h>

int64_t add(int64_t a, int64_t b) {
	return a+b;
}
void print_add_addr() {
	printf("%x\n", (int)(add));
}
*/
import "C"
import (
	"asmpkg"
	"fmt"
	"unsafe"
)

//go:noinline
//go:nosplit
func main() {
	println(C.add(1, 2))

	C.print_add_addr()
	fmt.Printf("%x\n", uintptr(unsafe.Pointer(C.add)))

	if true {
		c := asmpkg.CallCAdd(
			uintptr(unsafe.Pointer(C.add)),
			1, 2,
		)
		fmt.Printf("result: %x\n", c)
	}
}
