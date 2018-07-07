// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// +build amd64

// 汇编版本, 不支持内联优化

package add

func AsmAdd(a, b int) int

func AsmAddSlice(dst, a, b []int) {
	AddSlice(dst, a, b)
}

func AsmAddSlice__todo(dst, a, b []int)
