# Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
# License: https://creativecommons.org/licenses/by-nc-sa/4.0/

default: clean
	go tool cgo main.go
	go run main.go

clean:
	-rm -rf ./_obj
