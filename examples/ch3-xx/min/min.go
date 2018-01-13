// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Go版本, 支持内联优化

package min

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//go:noinline
func MinNoInline(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func AsmMin(a, b int) int
func AsmMax(a, b int) int
