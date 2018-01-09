# Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
# License: https://creativecommons.org/licenses/by-nc-sa/4.0/

default:
	go build -buildmode=c-shared -o say-hello.so main.go
	gcc -Wall _test_so.c ./say-hello.so
	./a.out

run-py3:
	python3 hello.py

clean:
	-rm *.so
	-rm *.out
