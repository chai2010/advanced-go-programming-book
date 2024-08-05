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
	mdbook serve


build:
	-rm book
	mdbook build
	-rm book/.gitignore
	-rm book/.nojekyll
	-rm -rf book/.git

deploy:
	-@make clean
	mnbook build
	-rm book/.gitignore
	-rm -rf book/.git
	-rm -rf book/examples

	cd book && git init
	cd book && git add .
	cd book && git commit -m "first commit"
	cd book && git branch -M gh-pages
	cd book && git remote add origin git@github.com:chai2010/advanced-go-programming-book.git
	cd book && git push -f origin gh-pages

clean:
	-rm -rf book
