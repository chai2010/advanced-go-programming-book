# Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
# License: https://creativecommons.org/licenses/by-nc-sa/4.0/

default:
	gcc -c -o number.o number.c
	ar rcs libnumber.a number.o
	-rm *.o

clean:
	-rm *.a *.o
