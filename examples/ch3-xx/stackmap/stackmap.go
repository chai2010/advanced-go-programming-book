// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package stackmap

func X(b []byte) []byte

//func X(b []byte) []byte {
//        if len(b) == cap(b) {
//                b = growSlice(b)
//        }
//        b = b[:len(b)+1]
//        b[len(b)-1] = 3
//        return b
//}

func growSlice(b []byte) []byte {
	newCap := 10
	if cap(b) > 5 {
		newCap = cap(b) * 2
	}
	b1 := make([]byte, len(b), newCap)
	copy(b1, b)
	return b1
}
