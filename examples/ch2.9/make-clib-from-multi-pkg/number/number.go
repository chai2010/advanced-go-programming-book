// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package number

import "C"

//export number_add_mod
func number_add_mod(a, b, mod C.int) C.int {
	return (a + b) % mod
}
