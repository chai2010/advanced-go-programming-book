// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

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
