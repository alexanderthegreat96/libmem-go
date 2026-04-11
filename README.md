# libmem-go

Go bindings for [libmem](https://github.com/rdbo/libmem) -- a cross-platform memory hacking library.

Supports **Linux**, **macOS**, **FreeBSD**, and **Windows**.

## Features

- Process and thread enumeration/discovery
- Module loading, unloading, and injection
- Memory read/write/allocate/protect (internal and cross-process)
- Pattern and signature scanning (byte patterns, masks, IDA-style signatures)
- Assembly and disassembly (via Keystone/Capstone)
- Function hooking (detours/trampolines)
- Virtual method table (VMT) hooking

## Prerequisites

- [Go](https://go.dev/dl/) 1.21+
- [Git](https://git-scm.com/)
- [CMake](https://cmake.org/) 3.14+
- C/C++ compiler (gcc, clang, or MinGW on Windows)

## Installation

```bash
go get github.com/alexanderthegreat96/libmem-go
```

Then build the C library locally:

```bash
go run ./cmd/setup
```

This clones, builds, and installs libmem into `libmem/deps/`. No system-wide installation required.

To pin a specific version:

```bash
go run ./cmd/setup v4.5.0
```

A `make setup` shorthand is also available on Linux/macOS.

## Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/alexanderthegreat96/libmem-go/libmem"
)

func main() {
    // Get current process info
    proc, err := libmem.GetProcess()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("PID: %d, Name: %s\n", proc.PID, proc.Name)

    // List all running processes
    procs, err := libmem.EnumProcesses()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d processes\n", len(procs))

    // Find a process by name
    target, err := libmem.FindProcess("target_app")
    if err != nil {
        log.Fatal(err)
    }

    // List modules in a remote process
    modules, err := libmem.EnumModulesEx(&target)
    if err != nil {
        log.Fatal(err)
    }
    for _, mod := range modules {
        fmt.Printf("  %s @ 0x%x\n", mod.Name, mod.Base)
    }

    // Read memory from a remote process
    data, err := libmem.ReadMemoryEx(&target, modules[0].Base, 64)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("First 64 bytes: %x\n", data)

    // Scan for an IDA-style signature
    addr, err := libmem.SigScanEx(&target, "48 89 5C 24 ?? 48 89 74", modules[0].Base, modules[0].Size)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Pattern found at: 0x%x\n", addr)
}
```

## API Overview

### Process

| Function | Description |
|---|---|
| `EnumProcesses()` | List all running processes |
| `GetProcess()` | Get current process info |
| `GetProcessEx(pid)` | Get process info by PID |
| `FindProcess(name)` | Find a process by name |
| `IsProcessAlive(process)` | Check if a process is running |
| `GetCommandLine(process)` | Get process command line arguments |
| `GetBits()` | Get current process bitness (32/64) |
| `GetSystemBits()` | Get OS bitness (32/64) |

### Thread

| Function | Description |
|---|---|
| `EnumThreads()` | List threads in current process |
| `EnumThreadsEx(process)` | List threads in a remote process |
| `GetThread()` | Get current thread info |
| `GetThreadEx(process)` | Get a thread from a remote process |
| `GetThreadProcess(thread)` | Get the owning process of a thread |

### Module

| Function | Description |
|---|---|
| `EnumModules()` / `EnumModulesEx(process)` | List loaded modules |
| `FindModule(name)` / `FindModuleEx(process, name)` | Find module by name |
| `LoadModule(path)` / `LoadModuleEx(process, path)` | Load/inject a module |
| `UnloadModule(module)` / `UnloadModuleEx(process, module)` | Unload a module |

### Memory

| Function | Description |
|---|---|
| `ReadMemory(addr, size)` / `ReadMemoryEx(...)` | Read memory |
| `WriteMemory(addr, data)` / `WriteMemoryEx(...)` | Write memory |
| `SetMemory(addr, byte, size)` / `SetMemoryEx(...)` | Fill memory |
| `ProtMemory(addr, size, prot)` / `ProtMemoryEx(...)` | Change protection |
| `AllocMemory(size, prot)` / `AllocMemoryEx(...)` | Allocate memory |
| `FreeMemory(addr, size)` / `FreeMemoryEx(...)` | Free memory |
| `DeepPointer(base, offsets)` / `DeepPointerEx(...)` | Resolve pointer chains |

### Scanning

| Function | Description |
|---|---|
| `DataScan(data, addr, size)` / `DataScanEx(...)` | Scan for exact bytes |
| `PatternScan(pattern, mask, addr, size)` / `PatternScanEx(...)` | Scan with mask (`x` = match, `?` = wildcard) |
| `SigScan(sig, addr, size)` / `SigScanEx(...)` | Scan for IDA-style signature |

### Assembly / Disassembly

| Function | Description |
|---|---|
| `GetArchitecture()` | Get CPU architecture |
| `Assemble(code)` | Assemble a single instruction |
| `AssembleEx(code, arch, addr)` | Assemble for a specific arch/address |
| `Disassemble(addr)` | Disassemble a single instruction |
| `DisassembleEx(addr, arch, maxSize, count, runtimeAddr)` | Disassemble multiple instructions |
| `CodeLength(addr, minLen)` / `CodeLengthEx(...)` | Get instruction-aligned code length |

### Hooking

| Function | Description |
|---|---|
| `HookCode(from, to)` / `HookCodeEx(...)` | Install a function hook |
| `UnhookCode(from, trampoline, size)` / `UnhookCodeEx(...)` | Remove a function hook |

### VMT (Virtual Method Table)

| Function | Description |
|---|---|
| `NewVMT(vtable)` | Create a VMT hook manager |
| `vmt.Hook(index, to)` | Hook a virtual function |
| `vmt.Unhook(index)` | Restore original virtual function |
| `vmt.GetOriginal(index)` | Get original function address |
| `vmt.Reset()` | Restore all virtual functions |
| `vmt.Free()` | Release VMT resources |

## Alternative Setup

If you prefer a system-wide install of libmem, the bindings will find it automatically:

```bash
git clone --recurse-submodules https://github.com/rdbo/libmem.git
cd libmem && mkdir build && cd build
cmake .. && make -j$(nproc)
sudo make install && sudo ldconfig
```

For custom install paths:

```bash
export CGO_CFLAGS="-I/path/to/include"
export CGO_LDFLAGS="-L/path/to/lib"
```

## License

These bindings are provided as-is. The underlying [libmem](https://github.com/rdbo/libmem) library is licensed under [GNU AGPLv3](https://github.com/rdbo/libmem/blob/master/LICENSE).
