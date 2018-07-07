# Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
# License: https://creativecommons.org/licenses/by-nc-sa/4.0/

default:
	go build -buildmode=c-archive -o main.a
	gcc -o a.out _test_main.c main.a
	./a.out

clean:
	-rm *.a *.lib
	-rm a.out
	-rm a.out.exe
