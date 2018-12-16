// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

//void SayHello(const char* s);
import "C"

func main() {
	C.SayHello(C.CString("Hello, World\n"))
}
