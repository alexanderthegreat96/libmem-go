package libmem

/*
#include <libmem/libmem.h>
*/
import "C"

import (
	"runtime/cgo"
	"unsafe"
)

//export goProcessCallback
func goProcessCallback(proc *C.lm_process_t, arg unsafe.Pointer) C.lm_bool_t {
	h := cgo.Handle(*(*C.uintptr_t)(arg))
	slice := h.Value().(*[]Process)
	*slice = append(*slice, processFromC(proc))
	return C.LM_TRUE
}

//export goThreadCallback
func goThreadCallback(thread *C.lm_thread_t, arg unsafe.Pointer) C.lm_bool_t {
	h := cgo.Handle(*(*C.uintptr_t)(arg))
	slice := h.Value().(*[]Thread)
	*slice = append(*slice, threadFromC(thread))
	return C.LM_TRUE
}

//export goModuleCallback
func goModuleCallback(module *C.lm_module_t, arg unsafe.Pointer) C.lm_bool_t {
	h := cgo.Handle(*(*C.uintptr_t)(arg))
	slice := h.Value().(*[]Module)
	*slice = append(*slice, moduleFromC(module))
	return C.LM_TRUE
}

//export goSymbolCallback
func goSymbolCallback(symbol *C.lm_symbol_t, arg unsafe.Pointer) C.lm_bool_t {
	h := cgo.Handle(*(*C.uintptr_t)(arg))
	slice := h.Value().(*[]Symbol)
	*slice = append(*slice, symbolFromC(symbol))
	return C.LM_TRUE
}

//export goSegmentCallback
func goSegmentCallback(segment *C.lm_segment_t, arg unsafe.Pointer) C.lm_bool_t {
	h := cgo.Handle(*(*C.uintptr_t)(arg))
	slice := h.Value().(*[]Segment)
	*slice = append(*slice, segmentFromC(segment))
	return C.LM_TRUE
}
