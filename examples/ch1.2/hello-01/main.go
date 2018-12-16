// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

//#include <stdio.h>
import "C"

func main() {
	C.puts(C.CString("Hello, World\n"))
}
