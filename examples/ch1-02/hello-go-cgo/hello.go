package main

// #include <stdio.h>
// #include <stdlib.h>
import "C"
import "unsafe"

func main() {
	msg := C.CString("Hello, World!\n")
	defer C.free(unsafe.Pointer(msg))

	C.fputs(msg, C.stdout)
}
