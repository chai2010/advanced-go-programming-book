// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package instr

func Add(n, m int64) int64 {
	return n + m
}

func Add2(n, m int64) int64

// BSF returns the index of the least significant set bit,
// or -1 if the input contains no set bits.
func BSF(n int64) int

func BSF32(n int32) int32

func Sum(s []int64) int64 {
	var ss int64
	for _, n := range s {
		ss += n
	}
	return ss
}

func Sum2(s []int64) int64
