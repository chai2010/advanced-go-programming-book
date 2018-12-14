// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Go版本, 支持内联优化

package add

func Add(a, b int) int {
	return a + b
}

func AddSlice(dst, a, b []int) {
	for i := 0; i < len(dst) && i < len(a) && i < len(b); i++ {
		dst[i] = a[i] + b[i]
	}
	return
}
