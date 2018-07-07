// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// go test -bench=.

package loop

import (
	"testing"
)

func TestLoopAdd(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		if x := LoopAdd(100, 0, 1); x != 100 {
			t.Fatalf("expect = %d, got = %d", 100, x)
		}
		if x := LoopAdd(100, 0, 2); x != 200 {
			t.Fatalf("expect = %d, got = %d", 200, x)
		}
		if x := LoopAdd(100, 0, -1); x != -100 {
			t.Fatalf("expect = %d, got = %d", -100, x)
		}
		if x := LoopAdd(100, 50, 1); x != 150 {
			t.Fatalf("expect = %d, got = %d", 150, x)
		}
	})
	t.Run("asm", func(t *testing.T) {
		if x := AsmLoopAdd(100, 0, 1); x != 100 {
			t.Fatalf("expect = %d, got = %d", 100, x)
		}
		if x := AsmLoopAdd(100, 0, 2); x != 200 {
			t.Fatalf("expect = %d, got = %d", 200, x)
		}
		if x := AsmLoopAdd(100, 0, -1); x != -100 {
			t.Fatalf("expect = %d, got = %d", -100, x)
		}
		if x := AsmLoopAdd(100, 50, 1); x != 150 {
			t.Fatalf("expect = %d, got = %d", 150, x)
		}
	})
}

func BenchmarkLoopAdd(b *testing.B) {
	b.Run("go", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			LoopAdd(1000, 0, 1)
		}
	})
	b.Run("asm", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AsmLoopAdd(1000, 0, 1)
		}
	})
}
