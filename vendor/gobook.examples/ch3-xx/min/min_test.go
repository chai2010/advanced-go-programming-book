// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// go test -bench=.

package min

import (
	"testing"
)

func TestMin(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		if x := Min(1, 2); x != 1 {
			t.Fatalf("expect = %d, got = %d", 1, x)
		}
		if x := Min(2, 1); x != 1 {
			t.Fatalf("expect = %d, got = %d", 1, x)
		}
	})
	t.Run("asm", func(t *testing.T) {
		if x := AsmMin(1, 2); x != 1 {
			t.Fatalf("expect = %d, got = %d", 1, x)
		}
		if x := AsmMin(2, 1); x != 1 {
			t.Fatalf("expect = %d, got = %d", 1, x)
		}
	})
}
func TestMax(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		if x := Max(1, 2); x != 2 {
			t.Fatalf("expect = %d, got = %d", 2, x)
		}
		if x := Max(2, 1); x != 2 {
			t.Fatalf("expect = %d, got = %d", 2, x)
		}
	})
	t.Run("asm", func(t *testing.T) {
		if x := AsmMax(1, 2); x != 2 {
			t.Fatalf("expect = %d, got = %d", 2, x)
		}
		if x := AsmMax(2, 1); x != 2 {
			t.Fatalf("expect = %d, got = %d", 2, x)
		}
	})
}

func BenchmarkMin(b *testing.B) {
	b.Run("go", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Min(1, 2)
		}
	})
	b.Run("go.noinline", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			MinNoInline(1, 2)
		}
	})
	b.Run("asm", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AsmMin(1, 2)
		}
	})
}
