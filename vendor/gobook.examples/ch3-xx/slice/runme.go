// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// +build ignore

package main

import (
	. "."
)

func main() {
	println("SumIntSlice([]int{1,2,3}) =", SumIntSlice([]int{1, 2, 3}))
	println("AsmSumIntSlice([]int{1,2,3}) =", AsmSumIntSlice([]int{1, 2, 3}))
}
