/* Created by cgo - DO NOT EDIT. */
#include <stdlib.h>
#include "_cgo_export.h"

extern void crosscall2(void (*fn)(void *, int, __SIZE_TYPE__), void *, int, __SIZE_TYPE__);
extern __SIZE_TYPE__ _cgo_wait_runtime_init_done();
extern void _cgo_release_context(__SIZE_TYPE__);

extern char* _cgo_topofstack(void);
#define CGO_NO_SANITIZE_THREAD
#define _cgo_tsan_acquire()
#define _cgo_tsan_release()

extern void _cgoexp_16f1900c27a8_helloInt(void *, int, __SIZE_TYPE__);

CGO_NO_SANITIZE_THREAD
void helloInt(GoInt p0)
{
	__SIZE_TYPE__ _cgo_ctxt = _cgo_wait_runtime_init_done();
	struct {
		GoInt p0;
	} __attribute__((__packed__)) a;
	a.p0 = p0;
	_cgo_tsan_release();
	crosscall2(_cgoexp_16f1900c27a8_helloInt, &a, 8, _cgo_ctxt);
	_cgo_tsan_acquire();
	_cgo_release_context(_cgo_ctxt);
}
extern void _cgoexp_16f1900c27a8_helloString(void *, int, __SIZE_TYPE__);

CGO_NO_SANITIZE_THREAD
void helloString(GoString p0)
{
	__SIZE_TYPE__ _cgo_ctxt = _cgo_wait_runtime_init_done();
	struct {
		GoString p0;
	} __attribute__((__packed__)) a;
	a.p0 = p0;
	_cgo_tsan_release();
	crosscall2(_cgoexp_16f1900c27a8_helloString, &a, 16, _cgo_ctxt);
	_cgo_tsan_acquire();
	_cgo_release_context(_cgo_ctxt);
}
extern void _cgoexp_16f1900c27a8_helloSlice(void *, int, __SIZE_TYPE__);

CGO_NO_SANITIZE_THREAD
void helloSlice(GoSlice p0)
{
	__SIZE_TYPE__ _cgo_ctxt = _cgo_wait_runtime_init_done();
	struct {
		GoSlice p0;
	} __attribute__((__packed__)) a;
	a.p0 = p0;
	_cgo_tsan_release();
	crosscall2(_cgoexp_16f1900c27a8_helloSlice, &a, 24, _cgo_ctxt);
	_cgo_tsan_acquire();
	_cgo_release_context(_cgo_ctxt);
}
