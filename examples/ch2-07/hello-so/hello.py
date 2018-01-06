from ctypes import *

libso = CDLL("./say-hello.so")

SayHello = libso.SayHello
SayHello.argtypes = [c_char_p]
SayHello.restype = None

SayHello(c_char_p(b"hello"))
