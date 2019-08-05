package main

import "runtime"

func main() {
	var buf = make([]byte, 64)
	var stk = buf[:runtime.Stack(buf, false)]
	println(string(stk))
}
