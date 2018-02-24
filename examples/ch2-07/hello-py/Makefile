# Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
# License: https://creativecommons.org/licenses/by-nc-sa/4.0/

default:
	go build -o py3-config.out py3-config.go
	PKG_CONFIG=./py3-config.out go build -buildmode=c-shared -o gopkg.so main.go
	-rm py3-config.out
	python3 -c 'import gopkg; print(gopkg.sum(1, 2))'

clean:
	-rm *.so
