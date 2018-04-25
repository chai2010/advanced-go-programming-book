// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

//int sum(int a, int b) { return a+b; }
import "C"

func main() {
	println(C.sum(1, 1))
}
