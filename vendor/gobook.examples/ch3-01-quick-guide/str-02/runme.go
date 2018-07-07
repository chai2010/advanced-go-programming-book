// +build ignore

package main

import (
	pkg "."
)

func main() {
	println(pkg.Name)

	pkg.NameData[0] = '?'
	println(pkg.Name)
}
