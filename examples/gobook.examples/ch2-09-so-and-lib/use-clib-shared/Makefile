# Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
# License: https://creativecommons.org/licenses/by-nc-sa/4.0/

default:
	cd number && make
	go build -o a.out
	LD_LIBRARY_PATH=$(shell pwd)/number ./a.out

macos:
	cd number && make
	go build -o a.out
	DYLD_LIBRARY_PATH=$(shell pwd)/number ./a.out

windows:
	# set path

clean:
	-cd number && make clean
	-rm *.a
	-rm a.out
