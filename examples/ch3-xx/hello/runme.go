// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// +build ignore

package main

import (
	"fmt"

	. "."
)

func main() {
	s := "你好, 世界!\n"
	fmt.Printf("%d: %x\n", len(s), s)
	PrintHelloWorld()
	PrintHelloWorld_zh()
	PrintHelloWorld_var()
}
