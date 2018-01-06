package main

/*
// macOS:
// python3-config --cflags
// python3-config --ldflags
#cgo darwin CFLAGS: -I/Library/Frameworks/Python.framework/Versions/3.6/include/python3.6m -I/Library/Frameworks/Python.framework/Versions/3.6/include/python3.6m -fno-strict-aliasing -Wsign-compare -fno-common -dynamic -DNDEBUG -g -fwrapv -O3 -Wall -Wstrict-prototypes  -g
#cgo darwin LDFLAGS: -L/Library/Frameworks/Python.framework/Versions/3.6/lib/python3.6/config-3.6m-darwin -lpython3.6m -ldl -framework CoreFoundation

// linux
#cgo linux pkg-config: python3

// windows
// should generate libpython3.a from python3.lib

#define Py_LIMITED_API
#include <Python.h>

static PyObject *
spam_system(PyObject *self, PyObject *args) {
	const char *command;
	if (!PyArg_ParseTuple(args, "s", &command)) {
		return NULL;
	}

	int status = system(command);
	return PyLong_FromLong(status);
}

extern PyObject *sum(PyObject *self, PyObject *args);

static PyMethodDef modMethods[] = {
	{"system",  spam_system, METH_VARARGS, "Execute a shell command."},
	{"sum",  sum, METH_VARARGS, "Execute a shell command."},
	{NULL, NULL, 0, NULL}
};

static PyObject* PyInit_gopkg_(void) {
	static struct PyModuleDef module = {
		PyModuleDef_HEAD_INIT, "gopkg", NULL, -1, modMethods,
	};
	return (void*)PyModule_Create(&module);
}
*/
import "C"

import (
	"fmt"
)

func main() {}

//  export SayHello
func SayHello(name *C.char) {
	fmt.Printf("hello %s!\n", C.GoString(name))
}

//export sum
func sum(self, args *C.PyObject) *C.PyObject {
	return C.PyLong_FromLongLong(9527) // TODO
}

//export PyInit_gopkg
func PyInit_gopkg() *C.PyObject {
	return C.PyInit_gopkg_()
}
