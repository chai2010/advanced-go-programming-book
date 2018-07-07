# Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
# License: https://creativecommons.org/licenses/by-nc-sa/4.0/

default:
	cd mystring && make
	go build -o a.out
	LD_LIBRARY_PATH=$(shell pwd)/mystring ./a.out

macos:
	cd mystring && make
	go build -o a.out
	DYLD_LIBRARY_PATH=$(shell pwd)/mystring ./a.out

windows:
	# set path

clean:
	-cd mystring && make clean
	-rm *.a
	-rm a.out
