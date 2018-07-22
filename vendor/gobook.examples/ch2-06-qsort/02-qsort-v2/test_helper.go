// Copyright © 2018 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package qsort

//extern int t_go_qsort_compare(void* a, void* b);
import "C"

import (
	"unsafe"
)

func t_get_go_qsort_compare() CompareFunc {
	return CompareFunc(C.t_go_qsort_compare)
}

//export t_go_qsort_compare
func t_go_qsort_compare(a, b unsafe.Pointer) C.int {
	pa, pb := (*C.int)(a), (*C.int)(b)
	return C.int(*pa - *pb)
}
