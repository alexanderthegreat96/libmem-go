package libmem

/*
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"runtime/cgo"
	"unsafe"
)

// EnumThreads returns a list of all threads in the current process.
func EnumThreads() ([]Thread, error) {
	var result []Thread
	h := cgo.NewHandle(&result)
	defer h.Delete()
	hVal := C.uintptr_t(h)
	ret := C.bridge_enum_threads(unsafe.Pointer(&hVal))
	if ret == C.LM_FALSE {
		return nil, errors.New("libmem: failed to enumerate threads")
	}
	return result, nil
}

// EnumThreadsEx returns a list of all threads in the given process.
func EnumThreadsEx(process *Process) ([]Thread, error) {
	cp := processToC(process)
	var result []Thread
	h := cgo.NewHandle(&result)
	defer h.Delete()
	hVal := C.uintptr_t(h)
	ret := C.bridge_enum_threads_ex(&cp, unsafe.Pointer(&hVal))
	if ret == C.LM_FALSE {
		return nil, errors.New("libmem: failed to enumerate threads")
	}
	return result, nil
}

// GetThread returns information about the current thread.
func GetThread() (Thread, error) {
	var ct C.lm_thread_t
	ret := C.LM_GetThread(&ct)
	if ret == C.LM_FALSE {
		return Thread{}, errors.New("libmem: failed to get current thread")
	}
	return threadFromC(&ct), nil
}

// GetThreadEx returns a thread from the given process.
func GetThreadEx(process *Process) (Thread, error) {
	cp := processToC(process)
	var ct C.lm_thread_t
	ret := C.LM_GetThreadEx(&cp, &ct)
	if ret == C.LM_FALSE {
		return Thread{}, errors.New("libmem: failed to get thread")
	}
	return threadFromC(&ct), nil
}

// GetThreadProcess returns the owning process of a thread.
func GetThreadProcess(thread *Thread) (Process, error) {
	ct := threadToC(thread)
	var cp C.lm_process_t
	ret := C.LM_GetThreadProcess(&ct, &cp)
	if ret == C.LM_FALSE {
		return Process{}, errors.New("libmem: failed to get thread process")
	}
	return processFromC(&cp), nil
}
