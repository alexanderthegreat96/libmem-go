package libmem

/*
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

// ReadMemory reads size bytes from the given address in the current process.
func ReadMemory(source uintptr, size uint) ([]byte, error) {
	if size == 0 {
		return nil, nil
	}
	buf := make([]byte, size)
	n := C.LM_ReadMemory(C.lm_address_t(source), (*C.lm_byte_t)(unsafe.Pointer(&buf[0])), C.lm_size_t(size))
	if uint(n) != size {
		return buf[:n], errors.New("libmem: incomplete read")
	}
	return buf, nil
}

// ReadMemoryEx reads size bytes from the given address in the given process.
func ReadMemoryEx(process *Process, source uintptr, size uint) ([]byte, error) {
	if size == 0 {
		return nil, nil
	}
	cp := processToC(process)
	buf := make([]byte, size)
	n := C.LM_ReadMemoryEx(&cp, C.lm_address_t(source), (*C.lm_byte_t)(unsafe.Pointer(&buf[0])), C.lm_size_t(size))
	if uint(n) != size {
		return buf[:n], errors.New("libmem: incomplete read")
	}
	return buf, nil
}

// WriteMemory writes data to the given address in the current process.
func WriteMemory(dest uintptr, data []byte) (uint, error) {
	if len(data) == 0 {
		return 0, nil
	}
	n := C.LM_WriteMemory(C.lm_address_t(dest), (*C.lm_byte_t)(unsafe.Pointer(&data[0])), C.lm_size_t(len(data)))
	if uint(n) != uint(len(data)) {
		return uint(n), errors.New("libmem: incomplete write")
	}
	return uint(n), nil
}

// WriteMemoryEx writes data to the given address in the given process.
func WriteMemoryEx(process *Process, dest uintptr, data []byte) (uint, error) {
	if len(data) == 0 {
		return 0, nil
	}
	cp := processToC(process)
	n := C.LM_WriteMemoryEx(&cp, C.lm_address_t(dest), (*C.lm_byte_t)(unsafe.Pointer(&data[0])), C.lm_size_t(len(data)))
	if uint(n) != uint(len(data)) {
		return uint(n), errors.New("libmem: incomplete write")
	}
	return uint(n), nil
}

// SetMemory fills size bytes at dest with the given byte value in the current process.
func SetMemory(dest uintptr, b byte, size uint) (uint, error) {
	n := C.LM_SetMemory(C.lm_address_t(dest), C.lm_byte_t(b), C.lm_size_t(size))
	if uint(n) != size {
		return uint(n), errors.New("libmem: incomplete set")
	}
	return uint(n), nil
}

// SetMemoryEx fills size bytes at dest with the given byte value in the given process.
func SetMemoryEx(process *Process, dest uintptr, b byte, size uint) (uint, error) {
	cp := processToC(process)
	n := C.LM_SetMemoryEx(&cp, C.lm_address_t(dest), C.lm_byte_t(b), C.lm_size_t(size))
	if uint(n) != size {
		return uint(n), errors.New("libmem: incomplete set")
	}
	return uint(n), nil
}

// ProtMemory changes the memory protection of a region in the current process.
// Returns the old protection flags.
func ProtMemory(address uintptr, size uint, prot Prot) (Prot, error) {
	var oldprot C.lm_prot_t
	ret := C.LM_ProtMemory(C.lm_address_t(address), C.lm_size_t(size), C.lm_prot_t(prot), &oldprot)
	if ret == C.LM_FALSE {
		return 0, errors.New("libmem: failed to change memory protection")
	}
	return Prot(oldprot), nil
}

// ProtMemoryEx changes the memory protection of a region in the given process.
// Returns the old protection flags.
func ProtMemoryEx(process *Process, address uintptr, size uint, prot Prot) (Prot, error) {
	cp := processToC(process)
	var oldprot C.lm_prot_t
	ret := C.LM_ProtMemoryEx(&cp, C.lm_address_t(address), C.lm_size_t(size), C.lm_prot_t(prot), &oldprot)
	if ret == C.LM_FALSE {
		return 0, errors.New("libmem: failed to change memory protection")
	}
	return Prot(oldprot), nil
}

// AllocMemory allocates memory in the current process with the given protection.
func AllocMemory(size uint, prot Prot) (uintptr, error) {
	addr := C.LM_AllocMemory(C.lm_size_t(size), C.lm_prot_t(prot))
	if uintptr(addr) == AddressBad {
		return 0, errors.New("libmem: failed to allocate memory")
	}
	return uintptr(addr), nil
}

// AllocMemoryEx allocates memory in the given process with the given protection.
func AllocMemoryEx(process *Process, size uint, prot Prot) (uintptr, error) {
	cp := processToC(process)
	addr := C.LM_AllocMemoryEx(&cp, C.lm_size_t(size), C.lm_prot_t(prot))
	if uintptr(addr) == AddressBad {
		return 0, errors.New("libmem: failed to allocate memory")
	}
	return uintptr(addr), nil
}

// FreeMemory frees previously allocated memory in the current process.
func FreeMemory(alloc uintptr, size uint) error {
	ret := C.LM_FreeMemory(C.lm_address_t(alloc), C.lm_size_t(size))
	if ret == C.LM_FALSE {
		return errors.New("libmem: failed to free memory")
	}
	return nil
}

// FreeMemoryEx frees previously allocated memory in the given process.
func FreeMemoryEx(process *Process, alloc uintptr, size uint) error {
	cp := processToC(process)
	ret := C.LM_FreeMemoryEx(&cp, C.lm_address_t(alloc), C.lm_size_t(size))
	if ret == C.LM_FALSE {
		return errors.New("libmem: failed to free memory")
	}
	return nil
}

// DeepPointer resolves a pointer chain (base + offsets) in the current process.
func DeepPointer(base uintptr, offsets []uintptr) (uintptr, error) {
	if len(offsets) == 0 {
		return base, nil
	}
	coffsets := make([]C.lm_address_t, len(offsets))
	for i, o := range offsets {
		coffsets[i] = C.lm_address_t(o)
	}
	addr := C.LM_DeepPointer(C.lm_address_t(base), &coffsets[0], C.size_t(len(coffsets)))
	if uintptr(addr) == AddressBad {
		return 0, errors.New("libmem: failed to resolve deep pointer")
	}
	return uintptr(addr), nil
}

// DeepPointerEx resolves a pointer chain (base + offsets) in the given process.
func DeepPointerEx(process *Process, base uintptr, offsets []uintptr) (uintptr, error) {
	if len(offsets) == 0 {
		return base, nil
	}
	cp := processToC(process)
	coffsets := make([]C.lm_address_t, len(offsets))
	for i, o := range offsets {
		coffsets[i] = C.lm_address_t(o)
	}
	addr := C.LM_DeepPointerEx(&cp, C.lm_address_t(base), &coffsets[0], C.lm_size_t(len(coffsets)))
	if uintptr(addr) == AddressBad {
		return 0, errors.New("libmem: failed to resolve deep pointer")
	}
	return uintptr(addr), nil
}
