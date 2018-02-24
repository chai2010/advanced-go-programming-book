// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Go版本, 支持内联优化

package slice

func SumIntSlice(s []int) int {
	var sum int
	for _, v := range s {
		sum += v
	}
	return sum
}

func SumFloat32Slice(s []float32) float32 {
	var sum float32
	for _, v := range s {
		sum += v
	}
	return sum
}

func SumFloat64Slice(s []float64) float64 {
	var sum float64
	for _, v := range s {
		sum += v
	}
	return sum
}

func AsmSumInt16Slice(v []int16) int16

func AsmSumIntSlice(s []int) int
func AsmSumIntSliceV2(s []int) int
