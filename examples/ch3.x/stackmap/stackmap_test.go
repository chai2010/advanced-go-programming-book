// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package stackmap

import (
	"bytes"
	"testing"
)

func TestX(t *testing.T) {
	b := make([]byte, 0, 3)

	for _, want := range [][]byte{
		mkSlice(3, 3),
		mkSlice(3, 3, 3),
		mkSlice(3, 3, 3, 3),
		mkSlice(10, 3, 3, 3, 3),
		mkSlice(10, 3, 3, 3, 3, 3),
		mkSlice(10, 3, 3, 3, 3, 3, 3),
		mkSlice(10, 3, 3, 3, 3, 3, 3, 3),
		mkSlice(10, 3, 3, 3, 3, 3, 3, 3, 3),
		mkSlice(10, 3, 3, 3, 3, 3, 3, 3, 3, 3),
		mkSlice(10, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3),
		mkSlice(20, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3),
	} {
		b = X(b)
		if !slicesEqual(b, want) {
			t.Fatalf("got %v[cap=%d]; want %v[cap=%d]",
				b, cap(b), want, cap(want))
		}
	}
}

func mkSlice(cap int, vs ...byte) []byte {
	b1 := make([]byte, 0, cap)
	for _, v := range vs {
		b1 = append(b1, v)
	}
	return b1
}
func slicesEqual(b0, b1 []byte) bool {
	if cap(b0) != cap(b1) {
		return false
	}
	return bytes.Equal(b0, b1)
}
