# Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
# License: https://creativecommons.org/licenses/by-nc-sa/4.0/

default:
	go build -buildmode=c-archive -o number.a
	gcc -o a.out _test_main.c number.a
	./a.out

build_on_win64:
	go build -buildmode=c-archive -o number.a
	gcc -m64 -shared -o number-win64.dll number-win64.def number.a -Wl,--allow-multiple-definition -static -lstdc++ -lwinmm -lntdll -lWs2_32
	lib /def:number-win64.def /machine:x64

build_with_vc:
	cl -o a.out.exe  _test_main.c number-win64.lib

clean:
	-rm *.a *.lib
	-rm a.out
	-rm a.out.exe
