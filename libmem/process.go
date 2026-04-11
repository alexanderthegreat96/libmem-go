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

// EnumProcesses returns a list of all running processes.
func EnumProcesses() ([]Process, error) {
	var result []Process
	h := cgo.NewHandle(&result)
	defer h.Delete()
	hVal := C.uintptr_t(h)
	ret := C.bridge_enum_processes(unsafe.Pointer(&hVal))
	if ret == C.LM_FALSE {
		return nil, errors.New("libmem: failed to enumerate processes")
	}
	return result, nil
}

// GetProcess returns information about the current process.
func GetProcess() (Process, error) {
	var cp C.lm_process_t
	ret := C.LM_GetProcess(&cp)
	if ret == C.LM_FALSE {
		return Process{}, errors.New("libmem: failed to get current process")
	}
	return processFromC(&cp), nil
}

// GetProcessEx returns information about a process by PID.
func GetProcessEx(pid uint32) (Process, error) {
	var cp C.lm_process_t
	ret := C.LM_GetProcessEx(C.lm_pid_t(pid), &cp)
	if ret == C.LM_FALSE {
		return Process{}, errors.New("libmem: failed to get process")
	}
	return processFromC(&cp), nil
}

// GetCommandLine returns the command line arguments of a process.
func GetCommandLine(process *Process) ([]string, error) {
	cp := processToC(process)
	cmdline := C.LM_GetCommandLine(&cp)
	if cmdline == nil {
		return nil, errors.New("libmem: failed to get command line")
	}
	defer C.LM_FreeCommandLine(cmdline)

	var args []string
	ptrSize := unsafe.Sizeof(uintptr(0))
	base := uintptr(unsafe.Pointer(cmdline))
	for i := uintptr(0); ; i++ {
		p := *(**C.char)(unsafe.Pointer(base + i*ptrSize))
		if p == nil {
			break
		}
		args = append(args, C.GoString(p))
	}
	return args, nil
}

// FindProcess finds a process by name.
func FindProcess(name string) (Process, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var cp C.lm_process_t
	ret := C.LM_FindProcess(cname, &cp)
	if ret == C.LM_FALSE {
		return Process{}, errors.New("libmem: process not found")
	}
	return processFromC(&cp), nil
}

// IsProcessAlive checks if a process is still running.
func IsProcessAlive(process *Process) bool {
	cp := processToC(process)
	return C.LM_IsProcessAlive(&cp) == C.LM_TRUE
}

// GetBits returns the bitness of the current process (32 or 64).
func GetBits() uint {
	return uint(C.LM_GetBits())
}

// GetSystemBits returns the bitness of the operating system (32 or 64).
func GetSystemBits() uint {
	return uint(C.LM_GetSystemBits())
}
