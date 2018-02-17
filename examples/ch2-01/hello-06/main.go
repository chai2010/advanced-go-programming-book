// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// +build go1.10

package main

//void SayHello(_GoString_ s);
import "C"

import (
	"fmt"
)

func main() {
	C.SayHello("Hello, World\n")
	C.SayHello("")
}

//export SayHello
func SayHello(s string) {
	fmt.Println(int(C._GoStringLen(s)))
	fmt.Print(s)
}
