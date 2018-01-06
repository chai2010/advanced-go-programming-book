// Created by cgo - DO NOT EDIT

package main

import "unsafe"

import _ "runtime/cgo"

import "syscall"

var _ syscall.Errno
func _Cgo_ptr(ptr unsafe.Pointer) unsafe.Pointer { return ptr }

//go:linkname _Cgo_always_false runtime.cgoAlwaysFalse
var _Cgo_always_false bool
//go:linkname _Cgo_use runtime.cgoUse
func _Cgo_use(interface{})
type _Ctype_void [0]byte

//go:linkname _cgo_runtime_cgocall runtime.cgocall
func _cgo_runtime_cgocall(unsafe.Pointer, uintptr) int32

//go:linkname _cgo_runtime_cgocallback runtime.cgocallback
func _cgo_runtime_cgocallback(unsafe.Pointer, unsafe.Pointer, uintptr, uintptr)

//go:linkname _cgoCheckPointer runtime.cgoCheckPointer
func _cgoCheckPointer(interface{}, ...interface{})

//go:linkname _cgoCheckResult runtime.cgoCheckResult
func _cgoCheckResult(interface{})

//go:cgo_export_dynamic helloInt
//go:linkname _cgoexp_16f1900c27a8_helloInt _cgoexp_16f1900c27a8_helloInt
//go:cgo_export_static _cgoexp_16f1900c27a8_helloInt
//go:nosplit
//go:norace
func _cgoexp_16f1900c27a8_helloInt(a unsafe.Pointer, n int32, ctxt uintptr) {
	fn := _cgoexpwrap_16f1900c27a8_helloInt
	_cgo_runtime_cgocallback(**(**unsafe.Pointer)(unsafe.Pointer(&fn)), a, uintptr(n), ctxt);
}

func _cgoexpwrap_16f1900c27a8_helloInt(p0 int) {
	helloInt(p0)
}
//go:cgo_export_dynamic helloString
//go:linkname _cgoexp_16f1900c27a8_helloString _cgoexp_16f1900c27a8_helloString
//go:cgo_export_static _cgoexp_16f1900c27a8_helloString
//go:nosplit
//go:norace
func _cgoexp_16f1900c27a8_helloString(a unsafe.Pointer, n int32, ctxt uintptr) {
	fn := _cgoexpwrap_16f1900c27a8_helloString
	_cgo_runtime_cgocallback(**(**unsafe.Pointer)(unsafe.Pointer(&fn)), a, uintptr(n), ctxt);
}

func _cgoexpwrap_16f1900c27a8_helloString(p0 string) {
	helloString(p0)
}
//go:cgo_export_dynamic helloSlice
//go:linkname _cgoexp_16f1900c27a8_helloSlice _cgoexp_16f1900c27a8_helloSlice
//go:cgo_export_static _cgoexp_16f1900c27a8_helloSlice
//go:nosplit
//go:norace
func _cgoexp_16f1900c27a8_helloSlice(a unsafe.Pointer, n int32, ctxt uintptr) {
	fn := _cgoexpwrap_16f1900c27a8_helloSlice
	_cgo_runtime_cgocallback(**(**unsafe.Pointer)(unsafe.Pointer(&fn)), a, uintptr(n), ctxt);
}

func _cgoexpwrap_16f1900c27a8_helloSlice(p0 []byte) {
	helloSlice(p0)
}
