// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

import "C"
import "fmt"

func main() {}

//export SayHello
func SayHello(name *C.char) {
	fmt.Printf("hello %s!\n", C.GoString(name))
}
