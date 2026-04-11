// Package libmem provides Go bindings for the libmem memory hacking library.
//
// libmem is a cross-platform library (Windows, Linux, FreeBSD, Android)
// for process/memory manipulation, pattern scanning, code hooking, and
// assembly/disassembly.
//
// Run the setup tool to build and install the C library locally:
//
//	go run ./cmd/setup
//
// Or if libmem is installed system-wide, it will be found automatically.
// For non-standard locations, set CGO_CFLAGS and CGO_LDFLAGS:
//
//	export CGO_CFLAGS="-I/path/to/libmem/include"
//	export CGO_LDFLAGS="-L/path/to/libmem/lib"
package libmem

/*
#cgo CFLAGS: -I${SRCDIR}/deps/include
#cgo LDFLAGS: -L${SRCDIR}/deps/lib -llibmem
#cgo linux LDFLAGS: -Wl,-rpath,${SRCDIR}/deps/lib
#cgo darwin LDFLAGS: -Wl,-rpath,${SRCDIR}/deps/lib
#cgo freebsd LDFLAGS: -Wl,-rpath,${SRCDIR}/deps/lib
#include "bridge.h"
#include <string.h>
*/
import "C"

import "unsafe"

// Constants matching libmem's C definitions.
const (
	PathMax    = C.LM_PATH_MAX
	InstMax    = C.LM_INST_MAX
	CmdlineMax = C.LM_CMDLINE_MAX

	AddressBad = ^uintptr(0)
	PIDBad     = ^uint32(0)
	TIDBad     = ^uint32(0)
)

// Prot represents memory protection flags.
type Prot uint32

const (
	ProtNone Prot = C.LM_PROT_NONE
	ProtR    Prot = C.LM_PROT_R
	ProtW    Prot = C.LM_PROT_W
	ProtX    Prot = C.LM_PROT_X
	ProtXR   Prot = C.LM_PROT_XR
	ProtXW   Prot = C.LM_PROT_XW
	ProtRW   Prot = C.LM_PROT_RW
	ProtXRW  Prot = C.LM_PROT_XRW
)

// Arch represents a CPU architecture for assembly/disassembly.
type Arch uint32

const (
	ArchGeneric   Arch = C.LM_ARCH_GENERIC
	ArchARMv7     Arch = C.LM_ARCH_ARMV7
	ArchARMv8     Arch = C.LM_ARCH_ARMV8
	ArchThumbv7   Arch = C.LM_ARCH_THUMBV7
	ArchThumbv8   Arch = C.LM_ARCH_THUMBV8
	ArchARMv7EB   Arch = C.LM_ARCH_ARMV7EB
	ArchThumbv7EB Arch = C.LM_ARCH_THUMBV7EB
	ArchARMv8EB   Arch = C.LM_ARCH_ARMV8EB
	ArchThumbv8EB Arch = C.LM_ARCH_THUMBV8EB
	ArchAArch64   Arch = C.LM_ARCH_AARCH64
	ArchMIPS      Arch = C.LM_ARCH_MIPS
	ArchMIPS64    Arch = C.LM_ARCH_MIPS64
	ArchMIPSEL    Arch = C.LM_ARCH_MIPSEL
	ArchMIPSEL64  Arch = C.LM_ARCH_MIPSEL64
	ArchX86_16    Arch = C.LM_ARCH_X86_16
	ArchX86       Arch = C.LM_ARCH_X86
	ArchX64       Arch = C.LM_ARCH_X64
	ArchPPC32     Arch = C.LM_ARCH_PPC32
	ArchPPC64     Arch = C.LM_ARCH_PPC64
	ArchPPC64LE   Arch = C.LM_ARCH_PPC64LE
	ArchSPARC     Arch = C.LM_ARCH_SPARC
	ArchSPARC64   Arch = C.LM_ARCH_SPARC64
	ArchSPARCEL   Arch = C.LM_ARCH_SPARCEL
	ArchSysZ      Arch = C.LM_ARCH_SYSZ
	ArchMax       Arch = C.LM_ARCH_MAX
)

// Process represents an OS process.
type Process struct {
	PID       uint32
	PPID      uint32
	Arch      Arch
	Bits      uint
	StartTime uint64
	Path      string
	Name      string
}

// Thread represents an OS thread.
type Thread struct {
	TID      uint32
	OwnerPID uint32
}

// Module represents a loaded module (shared library or executable).
type Module struct {
	Base uintptr
	End  uintptr
	Size uint
	Path string
	Name string
}

// Segment represents a memory segment.
type Segment struct {
	Base uintptr
	End  uintptr
	Size uint
	Prot Prot
}

// Symbol represents a symbol within a module.
type Symbol struct {
	Name    string
	Address uintptr
}

// Instruction represents a single assembled or disassembled CPU instruction.
type Instruction struct {
	Address  uintptr
	Size     uint
	Bytes    [InstMax]byte
	Mnemonic string
	OpStr    string
}

// VMT manages virtual method table hooks.
type VMT struct {
	vmt C.lm_vmt_t
}

// --- C-to-Go struct conversion helpers ---

func processFromC(p *C.lm_process_t) Process {
	return Process{
		PID:       uint32(p.pid),
		PPID:      uint32(p.ppid),
		Arch:      Arch(p.arch),
		Bits:      uint(p.bits),
		StartTime: uint64(p.start_time),
		Path:      C.GoString((*C.char)(unsafe.Pointer(&p.path[0]))),
		Name:      C.GoString((*C.char)(unsafe.Pointer(&p.name[0]))),
	}
}

func processToC(p *Process) C.lm_process_t {
	var cp C.lm_process_t
	cp.pid = C.lm_pid_t(p.PID)
	cp.ppid = C.lm_pid_t(p.PPID)
	cp.arch = C.lm_arch_t(p.Arch)
	cp.bits = C.lm_size_t(p.Bits)
	cp.start_time = C.lm_time_t(p.StartTime)
	copyStringToC((*C.char)(unsafe.Pointer(&cp.path[0])), p.Path, PathMax)
	copyStringToC((*C.char)(unsafe.Pointer(&cp.name[0])), p.Name, PathMax)
	return cp
}

func threadFromC(t *C.lm_thread_t) Thread {
	return Thread{
		TID:      uint32(t.tid),
		OwnerPID: uint32(t.owner_pid),
	}
}

func threadToC(t *Thread) C.lm_thread_t {
	var ct C.lm_thread_t
	ct.tid = C.lm_tid_t(t.TID)
	ct.owner_pid = C.lm_pid_t(t.OwnerPID)
	return ct
}

func moduleFromC(m *C.lm_module_t) Module {
	return Module{
		Base: uintptr(m.base),
		End:  uintptr(m.end),
		Size: uint(m.size),
		Path: C.GoString((*C.char)(unsafe.Pointer(&m.path[0]))),
		Name: C.GoString((*C.char)(unsafe.Pointer(&m.name[0]))),
	}
}

func moduleToC(m *Module) C.lm_module_t {
	var cm C.lm_module_t
	cm.base = C.lm_address_t(m.Base)
	cm.end = C.lm_address_t(m.End)
	cm.size = C.lm_size_t(m.Size)
	copyStringToC((*C.char)(unsafe.Pointer(&cm.path[0])), m.Path, PathMax)
	copyStringToC((*C.char)(unsafe.Pointer(&cm.name[0])), m.Name, PathMax)
	return cm
}

func segmentFromC(s *C.lm_segment_t) Segment {
	return Segment{
		Base: uintptr(s.base),
		End:  uintptr(s.end),
		Size: uint(s.size),
		Prot: Prot(s.prot),
	}
}

func symbolFromC(s *C.lm_symbol_t) Symbol {
	return Symbol{
		Name:    C.GoString((*C.char)(unsafe.Pointer(s.name))),
		Address: uintptr(s.address),
	}
}

func instructionFromC(i *C.lm_inst_t) Instruction {
	inst := Instruction{
		Address:  uintptr(i.address),
		Size:     uint(i.size),
		Mnemonic: C.GoString((*C.char)(unsafe.Pointer(&i.mnemonic[0]))),
		OpStr:    C.GoString((*C.char)(unsafe.Pointer(&i.op_str[0]))),
	}
	for j := 0; j < InstMax; j++ {
		inst.Bytes[j] = byte(i.bytes[j])
	}
	return inst
}

func copyStringToC(dst *C.char, src string, maxLen int) {
	cs := C.CString(src)
	defer C.free(unsafe.Pointer(cs))
	C.strncpy(dst, cs, C.size_t(maxLen-1))
}
