// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// +build ignore

package main

import (
	. "."
)

func main() {
	println("If(true, 1, 2) =", If(true, 1, 2))
	println("If(false, 1, 2) =", If(false, 1, 2))
	println("AsmIf(true, 1, 2) =", AsmIf(true, 1, 2))
	println("AsmIf(false, 1, 2) =", AsmIf(false, 1, 2))
	println("AsmIf(false, 2, 1) =", AsmIf(false, 2, 1))
}
