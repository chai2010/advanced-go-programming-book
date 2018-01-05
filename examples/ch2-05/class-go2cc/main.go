package main

// #cgo CXXFLAGS: -std=c++11
// extern void Main();
import "C"

func main() {
	C.Main()
}
