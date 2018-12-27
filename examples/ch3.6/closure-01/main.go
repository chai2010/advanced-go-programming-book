// Copyright © 2018 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

func NewTwiceFunClosure(x int) func() int {
	return func() int {
		x *= 2
		return x
	}
}

func main() {
	fnTwice := NewTwiceFunClosure(1)

	println(fnTwice()) // 1*2 => 2
	println(fnTwice()) // 2*2 => 4
	println(fnTwice()) // 4*2 => 8
}
