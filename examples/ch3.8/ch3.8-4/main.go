package main

import (
	"unsafe"
)

func getg() unsafe.Pointer

const g_goid_offset = 152 // Go1.8/Go1.9/Go1.10/Go1.11/Go1.12/Go1.13

func GetGroutineId() int64 {
	g := getg()
	p := (*int64)(unsafe.Pointer(uintptr(g) + g_goid_offset))
	return *p
}

func main() {
	println(GetGroutineId())
}
