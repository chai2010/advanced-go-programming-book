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

clean:
	-rm -rf book
