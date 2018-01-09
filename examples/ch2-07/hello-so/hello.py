# Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
# License: https://creativecommons.org/licenses/by-nc-sa/4.0/

from ctypes import *

libso = CDLL("./say-hello.so")

SayHello = libso.SayHello
SayHello.argtypes = [c_char_p]
SayHello.restype = None

SayHello(c_char_p(b"hello"))
