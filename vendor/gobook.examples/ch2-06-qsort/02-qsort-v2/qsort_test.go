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

	Sort(unsafe.Pointer(&values[0]),
		len(values), int(unsafe.Sizeof(values[0])),
		t_get_go_qsort_compare(),
	)

	isSorted := sort.SliceIsSorted(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	if !isSorted {
		t.Fatal("should be sorted")
	}
}
