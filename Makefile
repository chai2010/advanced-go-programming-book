# Copyright 2016 <chaishushan{AT}gmail.com>. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

#
# fix gitbook build error on macOS(node@8.x and gitbook@2.6.7)
#
# gitbook fetch 3.2.3
# gitbook build --gitbook=3.2.3
#
# https://github.com/GitbookIO/gitbook/issues/1774
# https://github.com/GitbookIO/gitbook-cli/blob/master/README.md
#

default:
	gitbook build

macos:
	gitbook build --gitbook=3.2.3

macos-pdf:
	mv preface.md preface-bak.md && mv preface-pdf.md preface.md
	gitbook pdf --gitbook=3.2.3
	mv preface.md preface-pdf.md && mv preface-bak.md preface.md

server:
	go run server.go

cover:
	convert A20181610.jpg cover.png
	convert -resize 1800x2360! A20181610.jpg cover.jpg
	convert -resize 200x262!   A20181610.jpg cover_small.jpg


# https://chai2010.cn/advanced-go-programming-book
deploy:
	-rm -rf _book
	gitbook build

	cd _book && \
		git init && \
		git add . && \
		git commit -m "Update github pages" && \
		git push --force --quiet "https://github.com/chai2010/advanced-go-programming-book.git" master:gh-pages

	@echo deploy done

clean:
	-rm -rf _book
