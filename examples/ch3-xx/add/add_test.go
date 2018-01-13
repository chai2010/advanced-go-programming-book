// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// go test -bench=.

package add

import (
	"testing"
)

func TestAdd(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		if x := Add(1, 2); x != 3 {
			t.Fatalf("expect = %d, got = %d", 3, x)
		}
	})
	t.Run("asm", func(t *testing.T) {
		if x := AsmAdd(1, 2); x != 3 {
			t.Fatalf("expect = %d, got = %d", 3, x)
		}
	})
}

func TestAddSlice(t *testing.T) {
	a := []int{1, 2, 3, 4, 5}
	b := []int{10, 20, 30, 40, 50, 60}

	t.Run("go", func(t *testing.T) {
		x := make([]int, len(a))
		AddSlice(x, a, b)

		for i := 0; i < len(x) && i < len(a) && i < len(b); i++ {
			if x[i] != a[i]+b[i] {
				t.Fatalf("expect = %d, got = %d", x[i], a[i]+b[i])
			}
		}
	})

	t.Run("asm", func(t *testing.T) {
		x := make([]int, len(a))
		AsmAddSlice(x, a, b)

		for i := 0; i < len(x) && i < len(a) && i < len(b); i++ {
			if x[i] != a[i]+b[i] {
				t.Fatalf("expect = %d, got = %d", x[i], a[i]+b[i])
			}
		}
	})
}

func BenchmarkAdd(b *testing.B) {
	b.Run("go", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Add(1, 2)
		}
	})
	b.Run("asm", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AsmAdd(1, 2)
		}
	})
}

func BenchmarkAddSlice(b *testing.B) {
	s0 := make([]int, 10<<10)
	s1 := make([]int, 10<<10)
	dst := make([]int, 10<<10)

	b.Run("len=10", func(b *testing.B) {
		dst := dst[:10]
		for i := 0; i < b.N; i++ {
			AddSlice(dst, s0, s1)
		}
	})
	b.Run("len=50", func(b *testing.B) {
		dst := dst[:50]
		for i := 0; i < b.N; i++ {
			AddSlice(dst, s0, s1)
			_ = dst
		}
	})
	b.Run("len=100", func(b *testing.B) {
		dst := dst[:100]
		for i := 0; i < b.N; i++ {
			AddSlice(dst, s0, s1)
			_ = dst
		}
	})
	b.Run("len=1000", func(b *testing.B) {
		dst := dst[:1000]
		for i := 0; i < b.N; i++ {
			AddSlice(dst, s0, s1)
			_ = dst
		}
	})
}
