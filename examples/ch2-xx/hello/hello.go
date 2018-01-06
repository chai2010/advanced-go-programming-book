package main

import "C"

func main() {
	helloString("hello") // _GoString_
}

//export helloInt
func helloInt(s int) {
	println(s)
}

//export helloString
func helloString(s string) {
	println(s)
}

//export helloSlice
func helloSlice(s []byte) {
	println(string(s))
}
