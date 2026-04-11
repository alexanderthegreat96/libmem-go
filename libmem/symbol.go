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

// EnumSymbols returns a list of all symbols in the given module.
func EnumSymbols(module *Module) ([]Symbol, error) {
	cm := moduleToC(module)
	var result []Symbol
	h := cgo.NewHandle(&result)
	defer h.Delete()
	hVal := C.uintptr_t(h)
	ret := C.bridge_enum_symbols(&cm, unsafe.Pointer(&hVal))
	if ret == C.LM_FALSE {
		return nil, errors.New("libmem: failed to enumerate symbols")
	}
	return result, nil
}

// FindSymbolAddress resolves a symbol's address within a module.
func FindSymbolAddress(module *Module, symbolName string) (uintptr, error) {
	cm := moduleToC(module)
	cname := C.CString(symbolName)
	defer C.free(unsafe.Pointer(cname))
	addr := C.LM_FindSymbolAddress(&cm, cname)
	if uintptr(addr) == AddressBad {
		return 0, errors.New("libmem: symbol not found")
	}
	return uintptr(addr), nil
}

// DemangleSymbol demangles a C++ symbol name.
func DemangleSymbol(symbolName string) (string, error) {
	cname := C.CString(symbolName)
	defer C.free(unsafe.Pointer(cname))
	var buf [4096]C.char
	result := C.LM_DemangleSymbol(cname, &buf[0], C.lm_size_t(len(buf)))
	if result == nil {
		return "", errors.New("libmem: failed to demangle symbol")
	}
	return C.GoString(result), nil
}

// EnumSymbolsDemangled returns a list of all demangled symbols in the given module.
func EnumSymbolsDemangled(module *Module) ([]Symbol, error) {
	cm := moduleToC(module)
	var result []Symbol
	h := cgo.NewHandle(&result)
	defer h.Delete()
	hVal := C.uintptr_t(h)
	ret := C.bridge_enum_symbols_demangled(&cm, unsafe.Pointer(&hVal))
	if ret == C.LM_FALSE {
		return nil, errors.New("libmem: failed to enumerate demangled symbols")
	}
	return result, nil
}

// FindSymbolAddressDemangled resolves a demangled symbol's address within a module.
func FindSymbolAddressDemangled(module *Module, symbolName string) (uintptr, error) {
	cm := moduleToC(module)
	cname := C.CString(symbolName)
	defer C.free(unsafe.Pointer(cname))
	addr := C.LM_FindSymbolAddressDemangled(&cm, cname)
	if uintptr(addr) == AddressBad {
		return 0, errors.New("libmem: demangled symbol not found")
	}
	return uintptr(addr), nil
}
