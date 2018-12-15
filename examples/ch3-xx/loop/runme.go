// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// +build ignore

package main

import (
	. "."
)

func main() {
	println("LoopAdd(100,0,1) =", LoopAdd(100, 0, 1))
	println("LoopAdd(100,0,2) =", LoopAdd(100, 0, 2))
	println("LoopAdd(100,200,-1) =", LoopAdd(100, 200, -1))
	println("LoopAdd(100,0,-1) =", LoopAdd(100, 0, -1))
}
