package libmem

/*
#include "bridge.h"
*/
import "C"

import "errors"

// HookCode hooks a function at 'from' to redirect to 'to' in the current process.
// Returns the trampoline address (for calling the original function) and the hook size.
func HookCode(from, to uintptr) (trampoline uintptr, size uint, err error) {
	var ct C.lm_address_t
	s := C.LM_HookCode(C.lm_address_t(from), C.lm_address_t(to), &ct)
	if s == 0 {
		return 0, 0, errors.New("libmem: failed to hook code")
	}
	return uintptr(ct), uint(s), nil
}

// HookCodeEx hooks a function at 'from' to redirect to 'to' in the given process.
// Returns the trampoline address and the hook size.
func HookCodeEx(process *Process, from, to uintptr) (trampoline uintptr, size uint, err error) {
	cp := processToC(process)
	var ct C.lm_address_t
	s := C.LM_HookCodeEx(&cp, C.lm_address_t(from), C.lm_address_t(to), &ct)
	if s == 0 {
		return 0, 0, errors.New("libmem: failed to hook code")
	}
	return uintptr(ct), uint(s), nil
}

// UnhookCode removes a hook previously installed with HookCode.
func UnhookCode(from, trampoline uintptr, size uint) error {
	ret := C.LM_UnhookCode(C.lm_address_t(from), C.lm_address_t(trampoline), C.lm_size_t(size))
	if ret == C.LM_FALSE {
		return errors.New("libmem: failed to unhook code")
	}
	return nil
}

// UnhookCodeEx removes a hook previously installed with HookCodeEx.
func UnhookCodeEx(process *Process, from, trampoline uintptr, size uint) error {
	cp := processToC(process)
	ret := C.LM_UnhookCodeEx(&cp, C.lm_address_t(from), C.lm_address_t(trampoline), C.lm_size_t(size))
	if ret == C.LM_FALSE {
		return errors.New("libmem: failed to unhook code")
	}
	return nil
}
