// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// +build ignore

package main

import (
	"fmt"

	. "."
)

func main() {
	fmt.Println(GetPkgValue())
	fmt.Println(GetPkgInfo())
}
