# libmem-go API Documentation

Complete API reference with examples for every function.

> **Note:** Most cross-process operations (`Ex` functions) require elevated privileges on Linux.
> Run with `sudo` or set `ptrace_scope` to 0: `sudo sysctl kernel.yama.ptrace_scope=0`

---

## Table of Contents

- [Types and Constants](#types-and-constants)
- [Process API](#process-api)
- [Thread API](#thread-api)
- [Module API](#module-api)
- [Symbol API](#symbol-api)
- [Segment API](#segment-api)
- [Memory API](#memory-api)
- [Scan API](#scan-api)
- [Assembly / Disassembly API](#assembly--disassembly-api)
- [Hook API](#hook-api)
- [VMT API](#vmt-api)

---

## Types and Constants

### Types

```go
// Memory protection flags
type Prot uint32

// CPU architecture identifier
type Arch uint32

// OS process
type Process struct {
    PID       uint32
    PPID      uint32
    Arch      Arch
    Bits      uint
    StartTime uint64
    Path      string
    Name      string
}

// OS thread
type Thread struct {
    TID      uint32
    OwnerPID uint32
}

// Loaded shared library or executable
type Module struct {
    Base uintptr
    End  uintptr
    Size uint
    Path string
    Name string
}

// Memory region
type Segment struct {
    Base uintptr
    End  uintptr
    Size uint
    Prot Prot
}

// Symbol in a module
type Symbol struct {
    Name    string
    Address uintptr
}

// Assembled or disassembled CPU instruction
type Instruction struct {
    Address  uintptr
    Size     uint
    Bytes    [16]byte
    Mnemonic string
    OpStr    string
}

// Virtual method table hook manager
type VMT struct { /* internal */ }
```

### Protection Flags

```go
libmem.ProtNone  // No access
libmem.ProtR     // Read
libmem.ProtW     // Write
libmem.ProtX     // Execute
libmem.ProtXR    // Execute + Read
libmem.ProtXW    // Execute + Write
libmem.ProtRW    // Read + Write
libmem.ProtXRW   // Execute + Read + Write
```

### Architecture Constants

```go
libmem.ArchX86       // x86 (32-bit)
libmem.ArchX64       // x86_64 (64-bit)
libmem.ArchAArch64   // ARM 64-bit
libmem.ArchARMv7     // ARM v7
libmem.ArchARMv8     // ARM v8
// ... and more (MIPS, PowerPC, SPARC, SystemZ)
```

### Special Values

```go
libmem.AddressBad  // Invalid address (returned on failure)
libmem.PIDBad      // Invalid PID
libmem.TIDBad      // Invalid TID
libmem.PathMax     // Maximum path length (4096)
libmem.InstMax     // Maximum instruction size (16 bytes)
```

---

## Process API

### EnumProcesses

Returns a list of all running processes on the system.

```go
procs, err := libmem.EnumProcesses()
if err != nil {
    log.Fatal(err)
}

for _, p := range procs {
    fmt.Printf("PID: %d | Name: %s | Path: %s\n", p.PID, p.Name, p.Path)
}
```

### GetProcess

Returns information about the current (calling) process.

```go
proc, err := libmem.GetProcess()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Current process: %s (PID: %d, %d-bit)\n", proc.Name, proc.PID, proc.Bits)
fmt.Printf("Path: %s\n", proc.Path)
fmt.Printf("Parent PID: %d\n", proc.PPID)
```

### GetProcessEx

Returns information about a process by its PID.

```go
proc, err := libmem.GetProcessEx(1234)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Process %d: %s (%d-bit)\n", proc.PID, proc.Name, proc.Bits)
```

### FindProcess

Finds a process by name.

```go
proc, err := libmem.FindProcess("firefox")
if err != nil {
    fmt.Println("Firefox is not running")
    return
}

fmt.Printf("Found firefox at PID %d\n", proc.PID)
fmt.Printf("Path: %s\n", proc.Path)
```

### GetCommandLine

Returns the command line arguments of a process.

```go
proc, err := libmem.GetProcess()
if err != nil {
    log.Fatal(err)
}

args, err := libmem.GetCommandLine(&proc)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Command line arguments:")
for i, arg := range args {
    fmt.Printf("  argv[%d] = %s\n", i, arg)
}
```

### IsProcessAlive

Checks whether a process is still running.

```go
proc, err := libmem.FindProcess("myapp")
if err != nil {
    log.Fatal(err)
}

if libmem.IsProcessAlive(&proc) {
    fmt.Printf("Process %d is still running\n", proc.PID)
} else {
    fmt.Printf("Process %d has exited\n", proc.PID)
}
```

### GetBits

Returns the bitness of the current process (32 or 64).

```go
bits := libmem.GetBits()
fmt.Printf("Current process is %d-bit\n", bits)
```

### GetSystemBits

Returns the bitness of the operating system.

```go
sysBits := libmem.GetSystemBits()
fmt.Printf("Operating system is %d-bit\n", sysBits)
```

---

## Thread API

### EnumThreads

Returns a list of all threads in the current process.

```go
threads, err := libmem.EnumThreads()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Current process has %d threads:\n", len(threads))
for _, t := range threads {
    fmt.Printf("  TID: %d (Owner PID: %d)\n", t.TID, t.OwnerPID)
}
```

### EnumThreadsEx

Returns a list of all threads in a remote process.

```go
proc, err := libmem.FindProcess("target_app")
if err != nil {
    log.Fatal(err)
}

threads, err := libmem.EnumThreadsEx(&proc)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Process %s has %d threads\n", proc.Name, len(threads))
for _, t := range threads {
    fmt.Printf("  TID: %d\n", t.TID)
}
```

### GetThread

Returns information about the current thread.

```go
thread, err := libmem.GetThread()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Current thread: TID %d (Owner PID: %d)\n", thread.TID, thread.OwnerPID)
```

### GetThreadEx

Returns a thread from a remote process.

```go
proc, err := libmem.FindProcess("target_app")
if err != nil {
    log.Fatal(err)
}

thread, err := libmem.GetThreadEx(&proc)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Thread %d from process %s\n", thread.TID, proc.Name)
```

### GetThreadProcess

Returns the owning process of a given thread.

```go
thread, err := libmem.GetThread()
if err != nil {
    log.Fatal(err)
}

proc, err := libmem.GetThreadProcess(&thread)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Thread %d belongs to process %s (PID: %d)\n", thread.TID, proc.Name, proc.PID)
```

---

## Module API

### EnumModules

Returns all loaded modules (shared libraries) in the current process.

```go
modules, err := libmem.EnumModules()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Loaded modules (%d):\n", len(modules))
for _, m := range modules {
    fmt.Printf("  %s @ 0x%x - 0x%x (size: %d)\n", m.Name, m.Base, m.End, m.Size)
}
```

### EnumModulesEx

Returns all loaded modules in a remote process.

```go
proc, err := libmem.FindProcess("target_app")
if err != nil {
    log.Fatal(err)
}

modules, err := libmem.EnumModulesEx(&proc)
if err != nil {
    log.Fatal(err)
}

for _, m := range modules {
    fmt.Printf("  %s @ 0x%x (size: %d)\n", m.Name, m.Base, m.Size)
    fmt.Printf("    Path: %s\n", m.Path)
}
```

### FindModule / FindModuleEx

Finds a loaded module by name.

```go
// In current process
mod, err := libmem.FindModule("libc.so.6")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("libc base: 0x%x, size: %d\n", mod.Base, mod.Size)

// In a remote process
proc, _ := libmem.FindProcess("target_app")
mod, err = libmem.FindModuleEx(&proc, "target_app")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Main module base: 0x%x\n", mod.Base)
```

### LoadModule / LoadModuleEx

Loads a shared library into a process.

```go
// Load into current process
mod, err := libmem.LoadModule("/path/to/mylib.so")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Loaded %s at 0x%x\n", mod.Name, mod.Base)

// Inject into a remote process
proc, _ := libmem.FindProcess("target_app")
mod, err = libmem.LoadModuleEx(&proc, "/path/to/inject.so")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Injected %s into PID %d at 0x%x\n", mod.Name, proc.PID, mod.Base)
```

### UnloadModule / UnloadModuleEx

Unloads a module from a process.

```go
// Unload from current process
mod, _ := libmem.FindModule("mylib.so")
err := libmem.UnloadModule(&mod)
if err != nil {
    log.Fatal(err)
}

// Unload from a remote process
proc, _ := libmem.FindProcess("target_app")
mod, _ = libmem.FindModuleEx(&proc, "inject.so")
err = libmem.UnloadModuleEx(&proc, &mod)
if err != nil {
    log.Fatal(err)
}
```

---

## Symbol API

### EnumSymbols

Lists all symbols in a module.

```go
mod, _ := libmem.FindModule("libc.so.6")

symbols, err := libmem.EnumSymbols(&mod)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d symbols in %s:\n", len(symbols), mod.Name)
for _, sym := range symbols[:10] { // Print first 10
    fmt.Printf("  %s @ 0x%x\n", sym.Name, sym.Address)
}
```

### FindSymbolAddress

Resolves a specific symbol's address.

```go
mod, _ := libmem.FindModule("libc.so.6")

addr, err := libmem.FindSymbolAddress(&mod, "printf")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("printf is at 0x%x\n", addr)
```

### DemangleSymbol

Demangles a C++ mangled symbol name.

```go
demangled, err := libmem.DemangleSymbol("_ZNSt6vectorIiSaIiEE9push_backERKi")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Demangled: %s\n", demangled)
// Output: std::vector<int, std::allocator<int>>::push_back(int const&)
```

### EnumSymbolsDemangled

Lists all symbols with demangled names.

```go
mod, _ := libmem.FindModule("libstdc++.so.6")

symbols, err := libmem.EnumSymbolsDemangled(&mod)
if err != nil {
    log.Fatal(err)
}

for _, sym := range symbols[:10] {
    fmt.Printf("  %s @ 0x%x\n", sym.Name, sym.Address)
}
```

### FindSymbolAddressDemangled

Finds a symbol by its demangled name.

```go
mod, _ := libmem.FindModule("libstdc++.so.6")

addr, err := libmem.FindSymbolAddressDemangled(&mod, "std::cout")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("std::cout is at 0x%x\n", addr)
```

---

## Segment API

### EnumSegments

Returns all memory segments (mapped regions) in the current process.

```go
segments, err := libmem.EnumSegments()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Memory segments (%d):\n", len(segments))
for _, seg := range segments {
    prot := ""
    if seg.Prot&libmem.ProtR != 0 { prot += "r" } else { prot += "-" }
    if seg.Prot&libmem.ProtW != 0 { prot += "w" } else { prot += "-" }
    if seg.Prot&libmem.ProtX != 0 { prot += "x" } else { prot += "-" }

    fmt.Printf("  0x%x - 0x%x [%s] (size: %d)\n", seg.Base, seg.End, prot, seg.Size)
}
```

### EnumSegmentsEx

Returns all memory segments in a remote process.

```go
proc, _ := libmem.FindProcess("target_app")

segments, err := libmem.EnumSegmentsEx(&proc)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Process %s has %d memory segments\n", proc.Name, len(segments))
for _, seg := range segments {
    fmt.Printf("  0x%x - 0x%x (prot: %d, size: %d)\n", seg.Base, seg.End, seg.Prot, seg.Size)
}
```

### FindSegment / FindSegmentEx

Finds the memory segment containing a given address.

```go
// In current process
mod, _ := libmem.FindModule("libc.so.6")

seg, err := libmem.FindSegment(mod.Base)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Segment at libc base: 0x%x - 0x%x (prot: %d)\n", seg.Base, seg.End, seg.Prot)

// In a remote process
proc, _ := libmem.FindProcess("target_app")
targetMod, _ := libmem.FindModuleEx(&proc, "target_app")

seg, err = libmem.FindSegmentEx(&proc, targetMod.Base)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Remote segment: 0x%x - 0x%x\n", seg.Base, seg.End)
```

---

## Memory API

### ReadMemory / ReadMemoryEx

Reads bytes from a memory address.

```go
// Read from current process
mod, _ := libmem.FindModule("libc.so.6")

data, err := libmem.ReadMemory(mod.Base, 16)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("First 16 bytes of libc: %x\n", data)

// Read from a remote process
proc, _ := libmem.FindProcess("target_app")
targetMod, _ := libmem.FindModuleEx(&proc, "target_app")

data, err = libmem.ReadMemoryEx(&proc, targetMod.Base, 64)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("First 64 bytes: %x\n", data)
```

### WriteMemory / WriteMemoryEx

Writes bytes to a memory address.

```go
// Write to current process
addr, _ := libmem.AllocMemory(4096, libmem.ProtRW)

n, err := libmem.WriteMemory(addr, []byte("Hello, libmem!"))
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Wrote %d bytes to 0x%x\n", n, addr)

// Write to a remote process
proc, _ := libmem.FindProcess("target_app")
remoteAddr, _ := libmem.AllocMemoryEx(&proc, 4096, libmem.ProtRW)

n, err = libmem.WriteMemoryEx(&proc, remoteAddr, []byte{0x90, 0x90, 0x90}) // NOP sled
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Wrote %d bytes to remote 0x%x\n", n, remoteAddr)
```

### SetMemory / SetMemoryEx

Fills a memory region with a single byte value (like `memset`).

```go
// Zero out a region in current process
addr, _ := libmem.AllocMemory(1024, libmem.ProtRW)

n, err := libmem.SetMemory(addr, 0x00, 1024)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Zeroed %d bytes at 0x%x\n", n, addr)

// Fill with NOPs in a remote process
proc, _ := libmem.FindProcess("target_app")
n, err = libmem.SetMemoryEx(&proc, someAddr, 0x90, 16)
if err != nil {
    log.Fatal(err)
}
```

### ProtMemory / ProtMemoryEx

Changes memory protection flags. Returns the old protection.

```go
// Make a region executable in current process
addr, _ := libmem.AllocMemory(4096, libmem.ProtRW)

oldProt, err := libmem.ProtMemory(addr, 4096, libmem.ProtXRW)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Changed protection at 0x%x: %d -> %d\n", addr, oldProt, libmem.ProtXRW)

// Change protection in a remote process
proc, _ := libmem.FindProcess("target_app")
oldProt, err = libmem.ProtMemoryEx(&proc, remoteAddr, 4096, libmem.ProtXRW)
if err != nil {
    log.Fatal(err)
}
```

### AllocMemory / AllocMemoryEx

Allocates memory with specified protection flags.

```go
// Allocate in current process
addr, err := libmem.AllocMemory(4096, libmem.ProtRW)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Allocated 4096 bytes at 0x%x\n", addr)

// Allocate in a remote process
proc, _ := libmem.FindProcess("target_app")
remoteAddr, err := libmem.AllocMemoryEx(&proc, 4096, libmem.ProtXRW)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Allocated remote memory at 0x%x\n", remoteAddr)
```

### FreeMemory / FreeMemoryEx

Frees previously allocated memory.

```go
// Free in current process
addr, _ := libmem.AllocMemory(4096, libmem.ProtRW)
// ... use the memory ...
err := libmem.FreeMemory(addr, 4096)
if err != nil {
    log.Fatal(err)
}

// Free in a remote process
proc, _ := libmem.FindProcess("target_app")
remoteAddr, _ := libmem.AllocMemoryEx(&proc, 4096, libmem.ProtRW)
// ... use the memory ...
err = libmem.FreeMemoryEx(&proc, remoteAddr, 4096)
if err != nil {
    log.Fatal(err)
}
```

### DeepPointer / DeepPointerEx

Resolves multi-level pointer chains. Useful for navigating game/application data structures.

```go
// Resolve a pointer chain: [[base + 0x10] + 0x20] + 0x30
// This follows: base -> read pointer at +0x10 -> read pointer at +0x20 -> add 0x30
addr, err := libmem.DeepPointer(baseAddr, []uintptr{0x10, 0x20, 0x30})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Resolved address: 0x%x\n", addr)

// Same but in a remote process
proc, _ := libmem.FindProcess("game")
mod, _ := libmem.FindModuleEx(&proc, "game")

healthAddr, err := libmem.DeepPointerEx(&proc, mod.Base+0x1A2B3C, []uintptr{0x10, 0x48, 0x100})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Health value at: 0x%x\n", healthAddr)

// Read the value
data, _ := libmem.ReadMemoryEx(&proc, healthAddr, 4)
fmt.Printf("Health bytes: %x\n", data)
```

---

## Scan API

### DataScan / DataScanEx

Scans for an exact byte sequence in memory.

```go
// Scan in current process
mod, _ := libmem.FindModule("libc.so.6")

// Search for the ELF magic bytes
addr, err := libmem.DataScan([]byte{0x7f, 'E', 'L', 'F'}, mod.Base, mod.Size)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("ELF header found at: 0x%x\n", addr)

// Scan in a remote process
proc, _ := libmem.FindProcess("target_app")
targetMod, _ := libmem.FindModuleEx(&proc, "target_app")

addr, err = libmem.DataScanEx(&proc, []byte{0x7f, 'E', 'L', 'F'}, targetMod.Base, targetMod.Size)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Remote ELF header at: 0x%x\n", addr)
```

### PatternScan / PatternScanEx

Scans for a byte pattern with a mask. Use `x` for bytes to match and `?` for wildcards.

```go
// Pattern: match specific bytes, skip wildcards
pattern := []byte{0x48, 0x89, 0x5C, 0x24, 0x00, 0x48, 0x89, 0x74}
mask := "xxxx?xxx" // '?' = wildcard for the 5th byte

mod, _ := libmem.FindModule("libc.so.6")

addr, err := libmem.PatternScan(pattern, mask, mod.Base, mod.Size)
if err != nil {
    fmt.Println("Pattern not found")
    return
}
fmt.Printf("Pattern found at: 0x%x\n", addr)

// Remote process
proc, _ := libmem.FindProcess("target_app")
targetMod, _ := libmem.FindModuleEx(&proc, "target_app")

addr, err = libmem.PatternScanEx(&proc, pattern, mask, targetMod.Base, targetMod.Size)
if err != nil {
    fmt.Println("Pattern not found in remote process")
    return
}
fmt.Printf("Remote pattern at: 0x%x\n", addr)
```

### SigScan / SigScanEx

Scans for an IDA-style signature. Use `??` for wildcard bytes.

```go
// IDA-style signature with wildcards
sig := "48 89 5C 24 ?? 48 89 74 24 ?? 57 48 83 EC"

mod, _ := libmem.FindModule("libc.so.6")

addr, err := libmem.SigScan(sig, mod.Base, mod.Size)
if err != nil {
    fmt.Println("Signature not found")
    return
}
fmt.Printf("Signature found at: 0x%x (offset: +0x%x)\n", addr, addr-mod.Base)

// Remote process
proc, _ := libmem.FindProcess("target_app")
targetMod, _ := libmem.FindModuleEx(&proc, "target_app")

addr, err = libmem.SigScanEx(&proc, sig, targetMod.Base, targetMod.Size)
if err != nil {
    fmt.Println("Signature not found in remote process")
    return
}
fmt.Printf("Remote signature at: 0x%x\n", addr)
```

---

## Assembly / Disassembly API

### GetArchitecture

Returns the CPU architecture of the current process.

```go
arch := libmem.GetArchitecture()
fmt.Printf("Architecture: %d\n", arch)

switch arch {
case libmem.ArchX64:
    fmt.Println("Running on x86_64")
case libmem.ArchX86:
    fmt.Println("Running on x86")
case libmem.ArchAArch64:
    fmt.Println("Running on ARM64")
}
```

### Assemble

Assembles a single instruction string into machine code.

```go
inst, err := libmem.Assemble("nop")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Instruction: %s %s\n", inst.Mnemonic, inst.OpStr)
fmt.Printf("Bytes: %x\n", inst.Bytes[:inst.Size])
fmt.Printf("Size: %d bytes\n", inst.Size)
```

### AssembleEx

Assembles code for a specific architecture and runtime address. Returns raw bytes.

```go
// Assemble x64 code at a specific address
payload, err := libmem.AssembleEx("mov rax, 0x1337; ret", libmem.ArchX64, 0x401000)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Assembled %d bytes: %x\n", len(payload), payload)
```

### Disassemble

Disassembles a single instruction at an address in the current process.

```go
mod, _ := libmem.FindModule("libc.so.6")

inst, err := libmem.Disassemble(mod.Base)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("0x%x: %s %s (%d bytes)\n", inst.Address, inst.Mnemonic, inst.OpStr, inst.Size)
fmt.Printf("Raw bytes: %x\n", inst.Bytes[:inst.Size])
```

### DisassembleEx

Disassembles multiple instructions.

```go
mod, _ := libmem.FindModule("libc.so.6")

// Disassemble 10 instructions, up to 128 bytes
instructions, err := libmem.DisassembleEx(
    mod.Base,        // address
    libmem.ArchX64,  // architecture
    128,             // max bytes to read
    10,              // max instructions
    mod.Base,        // runtime address for display
)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Disassembly:")
for _, inst := range instructions {
    fmt.Printf("  0x%x: %-8s %s\n", inst.Address, inst.Mnemonic, inst.OpStr)
}
```

### CodeLength / CodeLengthEx

Returns the minimum number of bytes that covers at least `minLength` bytes, aligned to instruction boundaries. Useful for hooking to know how many bytes to overwrite.

```go
mod, _ := libmem.FindModule("libc.so.6")

// Find instruction-aligned length >= 5 bytes (size of a JMP on x86_64)
length := libmem.CodeLength(mod.Base, 5)
fmt.Printf("Need to overwrite %d bytes for a hook at 0x%x\n", length, mod.Base)

// For a remote process
proc, _ := libmem.FindProcess("target_app")
targetMod, _ := libmem.FindModuleEx(&proc, "target_app")

length = libmem.CodeLengthEx(&proc, targetMod.Base, 5)
fmt.Printf("Remote: need %d bytes for hook\n", length)
```

---

## Hook API

### HookCode / HookCodeEx

Installs a function hook (detour) that redirects execution from one address to another. Returns a trampoline address that can be used to call the original function.

```go
// Hook a function in the current process
// 'from' = address of function to hook
// 'to' = address of your replacement function
trampoline, size, err := libmem.HookCode(targetFuncAddr, replacementFuncAddr)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Hooked! Trampoline at 0x%x (hook size: %d bytes)\n", trampoline, size)
// Call the trampoline to invoke the original function

// Hook in a remote process
proc, _ := libmem.FindProcess("target_app")

trampoline, size, err = libmem.HookCodeEx(&proc, remoteFuncAddr, remoteReplacementAddr)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Remote hook installed, trampoline: 0x%x\n", trampoline)
```

### UnhookCode / UnhookCodeEx

Removes a previously installed hook, restoring the original code.

```go
// Remove the hook using the values from HookCode
err := libmem.UnhookCode(targetFuncAddr, trampoline, size)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Hook removed, original function restored")

// Remove a remote hook
proc, _ := libmem.FindProcess("target_app")
err = libmem.UnhookCodeEx(&proc, remoteFuncAddr, trampoline, size)
if err != nil {
    log.Fatal(err)
}
```

### Complete hooking example

```go
package main

import (
    "fmt"
    "log"

    "github.com/alexanderthegreat96/libmem-go/libmem"
)

func main() {
    // Find the target function
    mod, err := libmem.FindModule("libc.so.6")
    if err != nil {
        log.Fatal(err)
    }

    targetAddr, err := libmem.FindSymbolAddress(&mod, "puts")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("puts is at 0x%x\n", targetAddr)

    // Calculate instruction-aligned hook size
    hookLen := libmem.CodeLength(targetAddr, 5)
    fmt.Printf("Hook will overwrite %d bytes\n", hookLen)

    // Install the hook (you'd point 'to' to your replacement function)
    trampoline, size, err := libmem.HookCode(targetAddr, replacementAddr)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Hook installed! Trampoline: 0x%x\n", trampoline)

    // ... do work with the hook active ...

    // Clean up
    err = libmem.UnhookCode(targetAddr, trampoline, size)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Hook removed")
}
```

---

## VMT API

The VMT (Virtual Method Table) API allows hooking virtual functions in C++ objects by modifying the vtable.

### NewVMT

Creates a VMT hook manager from a vtable address.

```go
// vtableAddr = address of the vtable (array of function pointers)
vmt, err := libmem.NewVMT(vtableAddr)
if err != nil {
    log.Fatal(err)
}
defer vmt.Free() // Always free when done

fmt.Println("VMT hook manager created")
```

### Hook

Replaces a virtual function at a given index.

```go
vmt, _ := libmem.NewVMT(vtableAddr)
defer vmt.Free()

// Hook the function at index 3 in the vtable
err := vmt.Hook(3, replacementFuncAddr)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Virtual function at index 3 hooked")
```

### GetOriginal

Retrieves the original function address before hooking.

```go
vmt, _ := libmem.NewVMT(vtableAddr)
defer vmt.Free()

// Save the original before hooking
original := vmt.GetOriginal(3)
fmt.Printf("Original function at index 3: 0x%x\n", original)

// Hook it
vmt.Hook(3, replacementFuncAddr)
```

### Unhook

Restores a single virtual function to its original.

```go
vmt, _ := libmem.NewVMT(vtableAddr)
defer vmt.Free()

vmt.Hook(3, replacementFuncAddr)
// ... do work ...

// Restore just index 3
err := vmt.Unhook(3)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Virtual function at index 3 restored")
```

### Reset

Restores all hooked virtual functions to their originals.

```go
vmt, _ := libmem.NewVMT(vtableAddr)
defer vmt.Free()

// Hook multiple functions
vmt.Hook(0, replacement0)
vmt.Hook(3, replacement3)
vmt.Hook(7, replacement7)

// ... do work ...

// Restore everything at once
vmt.Reset()
fmt.Println("All virtual functions restored")
```

### Free

Releases the VMT hook manager and its resources. Always call this when done.

```go
vmt, _ := libmem.NewVMT(vtableAddr)

// ... use the VMT ...

vmt.Free()
```

### Complete VMT example

```go
package main

import (
    "fmt"
    "log"

    "github.com/alexanderthegreat96/libmem-go/libmem"
)

func main() {
    proc, err := libmem.FindProcess("target_app")
    if err != nil {
        log.Fatal(err)
    }

    mod, err := libmem.FindModuleEx(&proc, "target_app")
    if err != nil {
        log.Fatal(err)
    }

    // Read the vtable pointer from an object
    // (object address found via pattern scanning or other means)
    objectAddr := mod.Base + 0x1234 // example offset

    // The first pointer in a C++ object is typically the vtable pointer
    vtableData, err := libmem.ReadMemoryEx(&proc, objectAddr, 8) // 8 bytes for 64-bit pointer
    if err != nil {
        log.Fatal(err)
    }

    // For internal (same-process) VMT hooking:
    vmt, err := libmem.NewVMT(vtableAddr)
    if err != nil {
        log.Fatal(err)
    }
    defer vmt.Free()

    // Save original
    originalFunc := vmt.GetOriginal(2)
    fmt.Printf("Original virtual function [2]: 0x%x\n", originalFunc)

    // Hook virtual function at index 2
    err = vmt.Hook(2, myReplacementAddr)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("VMT hook installed")

    // ... let the hooked code run ...

    // Cleanup
    vmt.Reset()
    fmt.Println("All hooks removed")
}
```
