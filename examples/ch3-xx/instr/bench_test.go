// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package instr

import "testing"

var g int64

func BenchmarkSum(b *testing.B) {
	ns := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for i := 0; i < b.N; i++ {
		g = Sum(ns)
	}
}

func BenchmarkSum2(b *testing.B) {
	ns := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for i := 0; i < b.N; i++ {
		g = Sum2(ns)
	}
}
