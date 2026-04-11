package libmem

/*
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

// DataScan scans for an exact byte sequence in the current process's memory.
func DataScan(data []byte, address uintptr, scanSize uint) (uintptr, error) {
	if len(data) == 0 {
		return 0, errors.New("libmem: empty scan data")
	}
	addr := C.LM_DataScan(
		(*C.lm_byte_t)(unsafe.Pointer(&data[0])),
		C.lm_size_t(len(data)),
		C.lm_address_t(address),
		C.lm_size_t(scanSize),
	)
	if uintptr(addr) == AddressBad {
		return 0, errors.New("libmem: data pattern not found")
	}
	return uintptr(addr), nil
}

// DataScanEx scans for an exact byte sequence in the given process's memory.
func DataScanEx(process *Process, data []byte, address uintptr, scanSize uint) (uintptr, error) {
	if len(data) == 0 {
		return 0, errors.New("libmem: empty scan data")
	}
	cp := processToC(process)
	addr := C.LM_DataScanEx(
		&cp,
		(*C.lm_byte_t)(unsafe.Pointer(&data[0])),
		C.lm_size_t(len(data)),
		C.lm_address_t(address),
		C.lm_size_t(scanSize),
	)
	if uintptr(addr) == AddressBad {
		return 0, errors.New("libmem: data pattern not found")
	}
	return uintptr(addr), nil
}

// PatternScan scans for a byte pattern with mask in the current process's memory.
// The mask uses 'x' for bytes to match and '?' for wildcards.
func PatternScan(pattern []byte, mask string, address uintptr, scanSize uint) (uintptr, error) {
	if len(pattern) == 0 {
		return 0, errors.New("libmem: empty pattern")
	}
	cmask := C.CString(mask)
	defer C.free(unsafe.Pointer(cmask))
	addr := C.LM_PatternScan(
		(*C.lm_byte_t)(unsafe.Pointer(&pattern[0])),
		cmask,
		C.lm_address_t(address),
		C.lm_size_t(scanSize),
	)
	if uintptr(addr) == AddressBad {
		return 0, errors.New("libmem: pattern not found")
	}
	return uintptr(addr), nil
}

// PatternScanEx scans for a byte pattern with mask in the given process's memory.
func PatternScanEx(process *Process, pattern []byte, mask string, address uintptr, scanSize uint) (uintptr, error) {
	if len(pattern) == 0 {
		return 0, errors.New("libmem: empty pattern")
	}
	cp := processToC(process)
	cmask := C.CString(mask)
	defer C.free(unsafe.Pointer(cmask))
	addr := C.LM_PatternScanEx(
		&cp,
		(*C.lm_byte_t)(unsafe.Pointer(&pattern[0])),
		cmask,
		C.lm_address_t(address),
		C.lm_size_t(scanSize),
	)
	if uintptr(addr) == AddressBad {
		return 0, errors.New("libmem: pattern not found")
	}
	return uintptr(addr), nil
}

// SigScan scans for an IDA-style signature in the current process's memory.
// Example signature: "48 89 5C 24 ?? 48 89 74 24"
func SigScan(signature string, address uintptr, scanSize uint) (uintptr, error) {
	csig := C.CString(signature)
	defer C.free(unsafe.Pointer(csig))
	addr := C.LM_SigScan(csig, C.lm_address_t(address), C.lm_size_t(scanSize))
	if uintptr(addr) == AddressBad {
		return 0, errors.New("libmem: signature not found")
	}
	return uintptr(addr), nil
}

// SigScanEx scans for an IDA-style signature in the given process's memory.
func SigScanEx(process *Process, signature string, address uintptr, scanSize uint) (uintptr, error) {
	cp := processToC(process)
	csig := C.CString(signature)
	defer C.free(unsafe.Pointer(csig))
	addr := C.LM_SigScanEx(&cp, csig, C.lm_address_t(address), C.lm_size_t(scanSize))
	if uintptr(addr) == AddressBad {
		return 0, errors.New("libmem: signature not found")
	}
	return uintptr(addr), nil
}
