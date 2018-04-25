# Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
# License: https://creativecommons.org/licenses/by-nc-sa/4.0/

default: clean
	# ignore nocgo_*.go files
	go tool cgo main.go hello.go

clean:
	-rm -rf ./_obj
