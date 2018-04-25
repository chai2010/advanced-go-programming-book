# Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
# License: https://creativecommons.org/licenses/by-nc-sa/4.0/

default:
	go tool cgo sum.go
	go build -buildmode=c-archive -o sum.a sum.go
	gcc -o a.out main.c sum.a
	./a.out

clean:
	-rm -rf ./_obj
	-rm sum.a sum.h
	-rm a.out

