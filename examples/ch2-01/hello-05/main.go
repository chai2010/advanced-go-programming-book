package main

//void SayHello(char* s);
import "C"

import (
	"fmt"
)

func main() {
	C.SayHello(C.CString("Hello, World\n"))
}

//export SayHello
func SayHello(s *C.char) {
	fmt.Print(C.GoString(s))
}
