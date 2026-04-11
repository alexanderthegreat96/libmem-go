package libmem

/*
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

// NewVMT creates a new virtual method table hook manager.
// vtable should be the address of the vtable (array of function pointers) in memory.
func NewVMT(vtable uintptr) (*VMT, error) {
	v := &VMT{}
	ret := C.LM_VmtNew((*C.lm_address_t)(unsafe.Pointer(vtable)), &v.vmt)
	if ret == C.LM_FALSE {
		return nil, errors.New("libmem: failed to create VMT")
	}
	return v, nil
}

// Hook replaces the virtual function at the given index with the function at 'to'.
func (v *VMT) Hook(fromIndex uint, to uintptr) error {
	ret := C.LM_VmtHook(&v.vmt, C.lm_size_t(fromIndex), C.lm_address_t(to))
	if ret == C.LM_FALSE {
		return errors.New("libmem: failed to hook VMT function")
	}
	return nil
}

// Unhook restores the original virtual function at the given index.
func (v *VMT) Unhook(fnIndex uint) error {
	ret := C.LM_VmtUnhook(&v.vmt, C.lm_size_t(fnIndex))
	if ret == C.LM_FALSE {
		return errors.New("libmem: failed to unhook VMT function")
	}
	return nil
}

// GetOriginal returns the original function address at the given index.
func (v *VMT) GetOriginal(fnIndex uint) uintptr {
	return uintptr(C.LM_VmtGetOriginal(&v.vmt, C.lm_size_t(fnIndex)))
}

// Reset restores all hooked virtual functions to their originals.
func (v *VMT) Reset() {
	C.LM_VmtReset(&v.vmt)
}

// Free releases the VMT hook manager resources.
func (v *VMT) Free() {
	C.LM_VmtFree(&v.vmt)
}
