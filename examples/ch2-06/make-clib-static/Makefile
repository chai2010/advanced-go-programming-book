# Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
# License: https://creativecommons.org/licenses/by-nc-sa/4.0/

default:
	go build -buildmode=c-archive -o number.a
	gcc -o a.out _test_main.c number.a
	./a.out

clean:
	-rm *.a
	-rm a.out
