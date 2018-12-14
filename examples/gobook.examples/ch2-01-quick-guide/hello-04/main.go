// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

/*
#include <stdio.h>

void cgoPuts(char* s);

static void SayHello(const char* s) {
	cgoPuts((char*)(s));
}
*/
import "C"
import "fmt"

func main() {
	C.SayHello(C.CString("Hello, World\n"))
}

//export cgoPuts
func cgoPuts(s *C.char) {
	fmt.Print(C.GoString(s))
}
