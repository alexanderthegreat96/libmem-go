package libmem

/*
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

// GetArchitecture returns the CPU architecture of the current process.
func GetArchitecture() Arch {
	return Arch(C.LM_GetArchitecture())
}

// Assemble assembles a single instruction string into machine code.
func Assemble(code string) (Instruction, error) {
	ccode := C.CString(code)
	defer C.free(unsafe.Pointer(ccode))
	var ci C.lm_inst_t
	ret := C.LM_Assemble(ccode, &ci)
	if ret == C.LM_FALSE {
		return Instruction{}, errors.New("libmem: failed to assemble instruction")
	}
	return instructionFromC(&ci), nil
}

// AssembleEx assembles code for the given architecture and runtime address.
// Returns the assembled payload bytes.
func AssembleEx(code string, arch Arch, runtimeAddress uintptr) ([]byte, error) {
	ccode := C.CString(code)
	defer C.free(unsafe.Pointer(ccode))
	var payload *C.lm_byte_t
	size := C.LM_AssembleEx(ccode, C.lm_arch_t(arch), C.lm_address_t(runtimeAddress), &payload)
	if size == 0 || payload == nil {
		return nil, errors.New("libmem: failed to assemble code")
	}
	defer C.LM_FreePayload(payload)
	return C.GoBytes(unsafe.Pointer(payload), C.int(size)), nil
}

// Disassemble disassembles a single instruction at the given address.
func Disassemble(machineCode uintptr) (Instruction, error) {
	var ci C.lm_inst_t
	ret := C.LM_Disassemble(C.lm_address_t(machineCode), &ci)
	if ret == C.LM_FALSE {
		return Instruction{}, errors.New("libmem: failed to disassemble instruction")
	}
	return instructionFromC(&ci), nil
}

// DisassembleEx disassembles multiple instructions at the given address.
func DisassembleEx(machineCode uintptr, arch Arch, maxSize uint, instructionCount uint, runtimeAddress uintptr) ([]Instruction, error) {
	var cinsts *C.lm_inst_t
	count := C.LM_DisassembleEx(
		C.lm_address_t(machineCode),
		C.lm_arch_t(arch),
		C.lm_size_t(maxSize),
		C.lm_size_t(instructionCount),
		C.lm_address_t(runtimeAddress),
		&cinsts,
	)
	if count == 0 || cinsts == nil {
		return nil, errors.New("libmem: failed to disassemble code")
	}
	defer C.LM_FreeInstructions(cinsts)

	result := make([]Instruction, count)
	instSize := unsafe.Sizeof(C.lm_inst_t{})
	base := uintptr(unsafe.Pointer(cinsts))
	for i := C.lm_size_t(0); i < count; i++ {
		ci := (*C.lm_inst_t)(unsafe.Pointer(base + uintptr(i)*instSize))
		result[i] = instructionFromC(ci)
	}
	return result, nil
}

// CodeLength returns the minimum number of bytes of machine code at the given
// address that covers at least minLength bytes (aligned to instruction boundaries).
func CodeLength(machineCode uintptr, minLength uint) uint {
	return uint(C.LM_CodeLength(C.lm_address_t(machineCode), C.lm_size_t(minLength)))
}

// CodeLengthEx returns the minimum number of bytes of machine code at the given
// address in the given process that covers at least minLength bytes.
func CodeLengthEx(process *Process, machineCode uintptr, minLength uint) uint {
	cp := processToC(process)
	return uint(C.LM_CodeLengthEx(&cp, C.lm_address_t(machineCode), C.lm_size_t(minLength)))
}
