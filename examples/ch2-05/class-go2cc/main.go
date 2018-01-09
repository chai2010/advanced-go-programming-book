// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package main

// #cgo CXXFLAGS: -std=c++11
// extern void Main();
import "C"

func main() {
	C.Main()
}
