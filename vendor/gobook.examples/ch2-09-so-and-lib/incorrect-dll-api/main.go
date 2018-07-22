// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

//#cgo CFLAGS: -I./mystring
//#cgo LDFLAGS: -L${SRCDIR}/mystring -lmystring
//
//#include "mystring.h"
//#include <stdlib.h>
import "C"
import (
	"fmt"
	"unsafe"
)

func main() {
	cs := C.make_string(C.CString("hello"))
	defer C.free(unsafe.Pointer(cs))

	fmt.Println(C.GoString(cs))
}
