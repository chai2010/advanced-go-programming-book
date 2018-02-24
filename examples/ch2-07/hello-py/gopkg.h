/* Created by "go tool cgo" - DO NOT EDIT. */

/* package command-line-arguments */

/* Start of preamble from import "C" comments.  */


#line 6 "/Users/chai/go/src/github.com/chai2010/advanced-go-programming-book/examples/ch2-07/hello-py/main.go"

// macOS:


// linux


// windows
// should generate libpython3.a from python3.lib

#define Py_LIMITED_API
#include <Python.h>

extern PyObject* PyInit_gopkg();
extern PyObject* Py_gopkg_sum(PyObject *, PyObject *);

static int cgo_PyArg_ParseTuple_ii(PyObject *arg, int *a, int *b) {
	return PyArg_ParseTuple(arg, "ii", a, b);
}

static PyObject* cgo_PyInit_gopkg(void) {
	static PyMethodDef methods[] = {
		{"sum", Py_gopkg_sum, METH_VARARGS, "Add two numbers."},
		{NULL, NULL, 0, NULL},
	};
	static struct PyModuleDef module = {
		PyModuleDef_HEAD_INIT, "gopkg", NULL, -1, methods,
	};
	return PyModule_Create(&module);
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


extern PyObject* PyInit_gopkg();

extern PyObject* Py_gopkg_sum(PyObject* p0, PyObject* p1);

#ifdef __cplusplus
}
#endif
