// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Go版本, 支持内联优化

package ifelse

func If(ok bool, a, b int) int {
	if ok {
		return a
	}
	return b
}

func AsmIf(ok bool, a, b int) int
