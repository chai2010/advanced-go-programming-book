// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// +build !amd64

// 对于没有汇编实现的环境, 临时采用Go版本代替

package add

func AsmAdd(a, b int) int {
	return Add(a, b)
}

func AsmAddSlice(dst, a, b []int) {
	AddSlice(dst, a, b)
}
