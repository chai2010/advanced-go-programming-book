package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, playground", LoopAdd(100, 1, 1))
}

func LoopAdd(cnt, v0, step int) int