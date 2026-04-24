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

The Go bindings require the libmem C library to be installed. A setup tool is included to automate this.

### Quick start (for users)

```bash
# 1. Install the setup tool
go install github.com/alexanderthegreat96/libmem-go/cmd/setup@latest

# 2. Build and install libmem system-wide (Linux/macOS, requires sudo)
sudo "$(go env GOPATH)/bin/setup"
sudo ldconfig

# 3. Use in your project
go get github.com/alexanderthegreat96/libmem-go
```

### Windows Setup

```cmd
REM 1. Install the setup tool
go install github.com/alexanderthegreat96/libmem-go/cmd/setup@latest

REM 2. Run setup (installs to %LOCALAPPDATA%\libmem)
setup
```

**Important:** Windows requires environment variables to be set before compiling or running. After setup completes, it will print the exact commands. Copy them to your shell:

**PowerShell:**
```powershell
$env:CGO_CFLAGS = '-IC:\Users\<your-username>\AppData\Local\libmem\include'
$env:CGO_LDFLAGS = '-LC:\Users\<your-username>\AppData\Local\libmem\lib'
$env:PATH = "$env:PATH;C:\Users\<your-username>\AppData\Local\libmem\lib"
```

**cmd.exe:**
```cmd
set CGO_CFLAGS=-IC:\Users\<your-username>\AppData\Local\libmem\include
set CGO_LDFLAGS=-LC:\Users\<your-username>\AppData\Local\libmem\lib
set PATH=%PATH%;C:\Users\<your-username>\AppData\Local\libmem\lib
```

To make this permanent, add these to your shell profile or user environment variables (System Properties → Environment Variables).

Then test:
```powershell
go run .    # Run your Go program
go build .  # Compile to executable
```

**Note:** When you compile with `go build`, the resulting `.exe` needs `libmem.dll` at runtime. Either:
- Keep the DLL on PATH (set above), or
- Copy `libmem.dll` to the same directory as your `.exe`

To pin a specific libmem version:
```bash
sudo "$(go env GOPATH)/bin/setup" v4.5.0
```

### For development (cloned repo)

```bash
git clone https://github.com/alexanderthegreat96/libmem-go.git
cd libmem-go
go run ./cmd/setup --local   # builds libmem into libmem/deps/
go build ./libmem
```

A `make setup` shorthand is also available on Linux/macOS (runs with `--local`).

## Documentation

For the full API reference with examples for every function, see [DOCUMENTATION.md](DOCUMENTATION.md).

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
## Building for Distribution

When you run `go build`, you get a single `.exe` file. However, the resulting executable is not fully self-contained — it has a runtime dependency on `libmem.dll`.

### Option 1: Ship with DLL (Recommended)
Bundle `libmem.dll` alongside your `.exe`:
```
my-app/
  ├── my-app.exe
  └── libmem.dll
```
Windows will automatically find the DLL next to your executable.

### Option 2: Add DLL to PATH
Ensure `C:\Users\<username>\AppData\Local\libmem\lib` is on the system PATH. Users can then run your `.exe` from anywhere.

### Option 3: Embed DLL (Advanced)
Use a tool like `go-embed` or [`pkger`](https://github.com/markbates/pkger) to embed `libmem.dll` inside your `.exe` and extract it at runtime. This is complex and rarely necessary.

### Troubleshooting "DLL not found"
If you get `libmem.dll not found` when running your `.exe`:
1. Check that `libmem.dll` is in the same directory as your `.exe`, or
2. Verify the DLL's directory is on PATH: `echo %PATH%` (cmd) or `$env:PATH` (PowerShell)

## Advanced Usage ([tether-go](https://github.com/alexanderthegreat96/tether-go))
You may use tether-go scaffolding in order to start with something with very little effort. 
The project has all you need design-pattern wise to implement your features.

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
