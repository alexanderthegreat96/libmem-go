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

// EnumModules returns a list of all loaded modules in the current process.
func EnumModules() ([]Module, error) {
	var result []Module
	h := cgo.NewHandle(&result)
	defer h.Delete()
	hVal := C.uintptr_t(h)
	ret := C.bridge_enum_modules(unsafe.Pointer(&hVal))
	if ret == C.LM_FALSE {
		return nil, errors.New("libmem: failed to enumerate modules")
	}
	return result, nil
}

// EnumModulesEx returns a list of all loaded modules in the given process.
func EnumModulesEx(process *Process) ([]Module, error) {
	cp := processToC(process)
	var result []Module
	h := cgo.NewHandle(&result)
	defer h.Delete()
	hVal := C.uintptr_t(h)
	ret := C.bridge_enum_modules_ex(&cp, unsafe.Pointer(&hVal))
	if ret == C.LM_FALSE {
		return nil, errors.New("libmem: failed to enumerate modules")
	}
	return result, nil
}

// FindModule finds a loaded module by name in the current process.
func FindModule(name string) (Module, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var cm C.lm_module_t
	ret := C.LM_FindModule(cname, &cm)
	if ret == C.LM_FALSE {
		return Module{}, errors.New("libmem: module not found")
	}
	return moduleFromC(&cm), nil
}

// FindModuleEx finds a loaded module by name in the given process.
func FindModuleEx(process *Process, name string) (Module, error) {
	cp := processToC(process)
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var cm C.lm_module_t
	ret := C.LM_FindModuleEx(&cp, cname, &cm)
	if ret == C.LM_FALSE {
		return Module{}, errors.New("libmem: module not found")
	}
	return moduleFromC(&cm), nil
}

// LoadModule loads a module (shared library) into the current process.
func LoadModule(path string) (Module, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	var cm C.lm_module_t
	ret := C.LM_LoadModule(cpath, &cm)
	if ret == C.LM_FALSE {
		return Module{}, errors.New("libmem: failed to load module")
	}
	return moduleFromC(&cm), nil
}

// LoadModuleEx injects a module into the given process.
func LoadModuleEx(process *Process, path string) (Module, error) {
	cp := processToC(process)
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	var cm C.lm_module_t
	ret := C.LM_LoadModuleEx(&cp, cpath, &cm)
	if ret == C.LM_FALSE {
		return Module{}, errors.New("libmem: failed to load module")
	}
	return moduleFromC(&cm), nil
}

// UnloadModule unloads a module from the current process.
func UnloadModule(module *Module) error {
	cm := moduleToC(module)
	ret := C.LM_UnloadModule(&cm)
	if ret == C.LM_FALSE {
		return errors.New("libmem: failed to unload module")
	}
	return nil
}

// UnloadModuleEx unloads a module from the given process.
func UnloadModuleEx(process *Process, module *Module) error {
	cp := processToC(process)
	cm := moduleToC(module)
	ret := C.LM_UnloadModuleEx(&cp, &cm)
	if ret == C.LM_FALSE {
		return errors.New("libmem: failed to unload module")
	}
	return nil
}
