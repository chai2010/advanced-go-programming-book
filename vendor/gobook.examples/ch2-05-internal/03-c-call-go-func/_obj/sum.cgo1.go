// Created by cgo - DO NOT EDIT

//line sum.go:1
// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

//int sum(int a, int b);
import _ "unsafe"

//export sum
func sum(a, b _Ctype_int) _Ctype_int {
	return a + b
}

func main() {}
