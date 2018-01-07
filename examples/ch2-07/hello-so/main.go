package main

import "C"
import "fmt"

func main() {}

//export SayHello
func SayHello(name *C.char) {
	fmt.Printf("hello %s!\n", C.GoString(name))
}
