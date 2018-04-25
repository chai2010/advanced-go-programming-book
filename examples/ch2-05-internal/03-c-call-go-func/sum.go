// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

//int sum(int a, int b);
import "C"

//export sum
func sum(a, b C.int) C.int {
	return a + b
}

func main() {}
