# Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
# License: https://creativecommons.org/licenses/by-nc-sa/4.0/

# use go generate

default:
	swig -go -cgo -intgosize 64 -o swig_wrap.cc hello.i
	go run runme.go

clean:
	-rm a.out
