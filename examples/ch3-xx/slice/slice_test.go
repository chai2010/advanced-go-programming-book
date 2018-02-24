// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// go test -bench=.

package slice

import (
	"testing"
)

func TestLoopAdd(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		if x := SumIntSlice([]int{1, 2, 3}); x != 6 {
			t.Fatalf("expect = %d, got = %d", 6, x)
		}
	})
	t.Run("asm", func(t *testing.T) {
		if x := AsmSumIntSlice([]int{1, 2, 3}); x != 6 {
			t.Fatalf("expect = %d, got = %d", 6, x)
		}
	})
	t.Run("asm.v2", func(t *testing.T) {
		if x := AsmSumIntSliceV2([]int{1, 2, 3}); x != 6 {
			t.Fatalf("expect = %d, got = %d", 6, x)
		}
	})
}

func BenchmarkLoopAdd(b *testing.B) {
	s10 := make([]int, 10)
	s100 := make([]int, 100)
	s1000 := make([]int, 1000)

	b.Run("go/len=10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			SumIntSlice(s10)
		}
	})
	b.Run("asm/len=10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AsmSumIntSlice(s10)
		}
	})
	b.Run("asm.v2/len=10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AsmSumIntSliceV2(s10)
		}
	})

	b.Run("go/len=100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			SumIntSlice(s100)
		}
	})
	b.Run("asm/len=100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AsmSumIntSlice(s100)
		}
	})
	b.Run("asm.v2/len=100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AsmSumIntSliceV2(s100)
		}
	})

	b.Run("go/len=1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			SumIntSlice(s1000)
		}
	})
	b.Run("asm/len=1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AsmSumIntSlice(s1000)
		}
	})
	b.Run("asm.v2/len=1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AsmSumIntSliceV2(s1000)
		}
	})
}
