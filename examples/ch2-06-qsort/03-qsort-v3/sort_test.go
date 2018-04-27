// Copyright © 2018 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package qsort

import (
	"sort"
	"testing"
	"unsafe"
)

func TestSort(t *testing.T) {
	values := []int32{42, 9, 101, 95, 27, 25}

	Sort(unsafe.Pointer(&values[0]), len(values), int(unsafe.Sizeof(values[0])),
		func(a, b unsafe.Pointer) int {
			pa, pb := (*int32)(a), (*int32)(b)
			return int(*pa - *pb)
		},
	)

	isSorted := sort.SliceIsSorted(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	if !isSorted {
		t.Fatal("should be sorted")
	}
}
