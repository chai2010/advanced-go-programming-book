/* Created by "go tool cgo" - DO NOT EDIT. */

/* package command-line-arguments */

/* Start of preamble from import "C" comments.  */


#line 3 "/Users/chai/go/src/github.com/chai2010/advanced-go-programming-book/examples/ch2-07/hello-py/main.go"

// macOS:
// python3-config --cflags
// python3-config --ldflags



// linux


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

#line 1 "cgo-generated-wrapper"


/* End of preamble from import "C" comments.  */


/* Start of boilerplate cgo prologue.  */
#line 1 "cgo-gcc-export-header-prolog"

#ifndef GO_CGO_PROLOGUE_H
#define GO_CGO_PROLOGUE_H

typedef signed char GoInt8;
typedef unsigned char GoUint8;
typedef short GoInt16;
typedef unsigned short GoUint16;
typedef int GoInt32;
typedef unsigned int GoUint32;
typedef long long GoInt64;
typedef unsigned long long GoUint64;
typedef GoInt64 GoInt;
typedef GoUint64 GoUint;
typedef __SIZE_TYPE__ GoUintptr;
typedef float GoFloat32;
typedef double GoFloat64;
typedef float _Complex GoComplex64;
typedef double _Complex GoComplex128;

/*
  static assertion to make sure the file is being used on architecture
  at least with matching size of GoInt.
*/
typedef char _check_for_64_bit_pointer_matching_GoInt[sizeof(void*)==64/8 ? 1:-1];

typedef struct { const char *p; GoInt n; } GoString;
typedef void *GoMap;
typedef void *GoChan;
typedef struct { void *t; void *v; } GoInterface;
typedef struct { void *data; GoInt len; GoInt cap; } GoSlice;

#endif

/* End of boilerplate cgo prologue.  */

#ifdef __cplusplus
extern "C" {
#endif


extern PyObject* sum(PyObject* p0, PyObject* p1);

extern PyObject* PyInit_gopkg();

#ifdef __cplusplus
}
#endif
