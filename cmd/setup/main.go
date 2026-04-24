// Setup tool for libmem-go. Clones, builds, and installs the libmem C library
// so that the Go bindings can link against it.
//
// Usage:
//
//	# From any project (installs system-wide to /usr/local):
//	go install github.com/alexanderthegreat96/libmem-go/cmd/setup@latest
//	sudo setup [version]
//
//	# From the libmem-go repo (installs locally to libmem/deps/):
//	go run ./cmd/setup [--local] [version]
//
// The version defaults to "master". Pass a git tag or branch name to pin a release.
// Requires: git, cmake, and a C/C++ compiler (gcc/clang/mingw).
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	repoURL        = "https://github.com/rdbo/libmem.git"
	defaultVersion = "master"
)

func main() {
	local := false
	version := defaultVersion
	// On Windows the default is static — Go's "one self-contained exe" model
	// doesn't play well with DLLs (PATH lookups, distribution, etc.). On other
	// platforms dynamic is the norm (rpath + ldconfig handle it).
	static := runtime.GOOS == "windows"

	for _, arg := range os.Args[1:] {
		switch arg {
		case "--local":
			local = true
		case "--static":
			static = true
		case "--shared":
			static = false
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		default:
			version = arg
		}
	}

	linkMode := "shared"
	if static {
		linkMode = "static"
	}
	fmt.Printf("Setting up libmem (%s) for %s/%s (%s)\n\n",
		version, runtime.GOOS, runtime.GOARCH, linkMode)

	checkPrereqs()

	if runtime.GOOS == "windows" {
		ensureMSVC()
	}

	// Decide install location
	installDir := resolveInstallDir(local)
	libDir := filepath.Join(installDir, "lib")
	incDir := filepath.Join(installDir, "include", "libmem")

	// Check write permissions BEFORE doing the expensive clone+build
	step("Install target: %s", installDir)
	checkWritable(installDir)

	// Use a proper temp directory for the build
	tmpDir, err := os.MkdirTemp("", "libmem-build-*")
	check(err, "creating temp directory")
	defer os.RemoveAll(tmpDir)
	buildDir := filepath.Join(tmpDir, "build")

	srcDir := filepath.Join(tmpDir, "src")

	// Clone
	step("Cloning libmem (%s)...", version)
	runCmd("git", "clone", "--depth", "1", "--recurse-submodules",
		"--branch", version, repoURL, srcDir)

	step("Configuring build...")
	os.RemoveAll(buildDir)
	check(os.MkdirAll(buildDir, 0755), "creating build directory")

	cmakeArgs := []string{"-S", srcDir, "-B", buildDir, "-DLIBMEM_BUILD_TESTS=OFF"}
	if static {
		cmakeArgs = append(cmakeArgs, "-DLIBMEM_BUILD_STATIC=ON")
	}
	if runtime.GOOS == "windows" {
		// libmem's PreLoad.cmake forces CMAKE_GENERATOR="NMake Makefiles" on
		// native Windows, so -G must match. MSVC env is already set up by
		// ensureMSVC above.
		cmakeArgs = append(cmakeArgs, "-G", "NMake Makefiles")
	}
	runCmake(cmakeArgs...)

	// Build
	step("Building (this may take a few minutes)...")
	runCmake("--build", buildDir, "--config", "Release",
		"-j", fmt.Sprintf("%d", runtime.NumCPU()))

	// Copy artifacts
	step("Installing to %s...", installDir)
	check(os.MkdirAll(libDir, 0755), "creating lib directory")
	check(os.MkdirAll(incDir, 0755), "creating include directory")

	if !installLibrary(buildDir, libDir, static) {
		fatal("could not find built library in %s", buildDir)
	}
	installHeaders(filepath.Join(srcDir, "include", "libmem"), incDir)
	// tmpDir is cleaned up by defer

	fmt.Println()
	step("Done! libmem installed to %s", installDir)

	// Post-install hints
	if !local && runtime.GOOS == "linux" && !static {
		fmt.Println("  You may need to run: sudo ldconfig")
	}
	if local {
		fmt.Println("  You can now build with: go build ./libmem")
	} else {
		fmt.Println("  You can now use: go get github.com/alexanderthegreat96/libmem-go")
		if runtime.GOOS == "windows" {
			fmt.Printf("\n  Set build-time env vars (once per shell, or put in your profile):\n")
			fmt.Printf("    set CGO_CFLAGS=-I%s\n", filepath.Join(installDir, "include"))
			fmt.Printf("    set CGO_LDFLAGS=-L%s\n", filepath.Join(installDir, "lib"))
			if !static {
				fmt.Printf("    set PATH=%%PATH%%;%s\n", filepath.Join(installDir, "lib"))
				fmt.Println("  (PATH entry is required so Windows can find libmem.dll at runtime.)")
			} else {
				fmt.Println("  No runtime PATH entry needed — libmem is linked into your .exe.")
			}
		}
	}
}

func printUsage() {
	fmt.Println(`Usage: setup [options] [version]

Options:
  --local    Install to libmem/deps/ in the current repo (for development)
  --static   Build libmem as a static library (default on Windows)
  --shared   Build libmem as a shared library (default elsewhere)
  --help     Show this help

Arguments:
  version    Git tag or branch to build (default: "master")

Without --local, installs system-wide:
  Linux/macOS/FreeBSD: /usr/local  (may need sudo)
  Windows:             %LOCALAPPDATA%\libmem`)
}

// resolveInstallDir determines where to install based on flags and platform.
func resolveInstallDir(local bool) string {
	if local {
		// --local: install relative to cwd (for repo development)
		cwd, err := os.Getwd()
		check(err, "getting working directory")
		return filepath.Join(cwd, "libmem", "deps")
	}

	switch runtime.GOOS {
	case "windows":
		appdata := os.Getenv("LOCALAPPDATA")
		if appdata == "" {
			fatal("LOCALAPPDATA is not set")
		}
		return filepath.Join(appdata, "libmem")
	default:
		return "/usr/local"
	}
}

// checkWritable verifies we can write to the install directory before doing
// the expensive clone+build. Creates a test file and removes it.
func checkWritable(dir string) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		if os.IsPermission(err) && runtime.GOOS != "windows" {
			fatal("permission denied: cannot write to %s\n  Run with sudo: sudo env \"PATH=$PATH\" %s", dir, strings.Join(os.Args, " "))
		}
		check(err, "accessing install directory")
	}
	testFile := filepath.Join(dir, ".libmem-write-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		if os.IsPermission(err) && runtime.GOOS != "windows" {
			fatal("permission denied: cannot write to %s\n  Run with sudo: sudo env \"PATH=$PATH\" %s", dir, strings.Join(os.Args, " "))
		}
		check(err, "checking write access")
	}
	os.Remove(testFile)
}

// checkPrereqs verifies that required tools are available.
func checkPrereqs() {
	for _, name := range []string{"git", "cmake"} {
		if _, err := exec.LookPath(name); err != nil {
			fatal("%s is required but not found in PATH", name)
		}
	}
}

// ensureMSVC makes sure the MSVC toolchain (cl.exe + nmake.exe) is available
// for cmake. libmem's PreLoad.cmake forces NMake Makefiles on Windows, so MSVC
// is the only supported toolchain. We try to make this universal: if nmake
// isn't already on PATH, locate Visual Studio via vswhere.exe and activate
// vcvarsall.bat in-process so users don't have to open a Developer Command
// Prompt first.
func ensureMSVC() {
	if _, err := exec.LookPath("nmake"); err == nil {
		return // MSVC env already active
	}

	vcvars := findVcvarsall()
	if vcvars == "" {
		fatal("Visual Studio MSVC toolchain not found.\n" +
			"  libmem on Windows requires cl.exe + nmake.exe. Install Visual Studio 2019+\n" +
			"  (any edition) or the standalone Build Tools, with the\n" +
			"  \"Desktop development with C++\" workload:\n" +
			"    https://visualstudio.microsoft.com/downloads/\n" +
			"  Then re-run this setup.")
	}

	arch := "x64"
	switch runtime.GOARCH {
	case "386":
		arch = "x86"
	case "arm64":
		arch = "x64_arm64"
	}

	step("Activating MSVC environment (%s)", arch)
	fmt.Printf("  vcvarsall: %s\n", vcvars)
	applyVcvarsEnv(vcvars, arch)

	if _, err := exec.LookPath("nmake"); err != nil {
		fatal("MSVC activation completed but nmake is still not on PATH.\n" +
			"  Your Visual Studio install may be missing the C++ build tools.\n" +
			"  Open the Visual Studio Installer and add the\n" +
			"  \"Desktop development with C++\" workload.")
	}
}

// findVcvarsall locates vcvarsall.bat by asking vswhere for the latest VS
// install that has the MSVC x86/x64 compiler component. Falls back to probing
// standard install paths if vswhere is not present (rare on modern Windows).
func findVcvarsall() string {
	pf86 := os.Getenv("ProgramFiles(x86)")
	if pf86 == "" {
		pf86 = `C:\Program Files (x86)`
	}
	vswhere := filepath.Join(pf86, "Microsoft Visual Studio", "Installer", "vswhere.exe")
	if _, err := os.Stat(vswhere); err == nil {
		out, err := exec.Command(vswhere,
			"-latest",
			"-products", "*",
			"-requires", "Microsoft.VisualStudio.Component.VC.Tools.x86.x64",
			"-property", "installationPath",
		).Output()
		if err == nil {
			installPath := strings.TrimSpace(string(out))
			if installPath != "" {
				p := filepath.Join(installPath, "VC", "Auxiliary", "Build", "vcvarsall.bat")
				if _, err := os.Stat(p); err == nil {
					return p
				}
			}
		}
	}

	// Fallback probe — older machines without vswhere.
	pf := os.Getenv("ProgramFiles")
	if pf == "" {
		pf = `C:\Program Files`
	}
	for _, root := range []string{pf, pf86} {
		for _, year := range []string{"2022", "2019", "2017"} {
			for _, ed := range []string{"Enterprise", "Professional", "Community", "BuildTools"} {
				p := filepath.Join(root, "Microsoft Visual Studio", year, ed,
					"VC", "Auxiliary", "Build", "vcvarsall.bat")
				if _, err := os.Stat(p); err == nil {
					return p
				}
			}
		}
	}
	return ""
}

// applyVcvarsEnv runs vcvarsall.bat for the given arch in a subshell, captures
// the resulting environment, and applies it to the current process so all
// subsequent child processes (cmake, nmake) inherit it.
//
// We stage the commands in a temp .bat file to avoid cmd.exe's famously broken
// nested-quote handling when combined with Go's arg-escaping.
func applyVcvarsEnv(vcvars, arch string) {
	const delim = "---LIBMEM-VCVARS-ENV---"
	bat, err := os.CreateTemp("", "libmem-vcvars-*.bat")
	check(err, "creating vcvars shim")
	batPath := bat.Name()
	defer os.Remove(batPath)
	_, werr := fmt.Fprintf(bat,
		"@echo off\r\ncall \"%s\" %s >NUL\r\nif errorlevel 1 exit /b %%errorlevel%%\r\necho %s\r\nset\r\n",
		vcvars, arch, delim)
	check(werr, "writing vcvars shim")
	check(bat.Close(), "closing vcvars shim")

	out, err := exec.Command("cmd.exe", "/c", batPath).Output()
	if err != nil {
		fatal("running vcvarsall.bat: %v\n  output: %s", err, string(out))
	}

	idx := strings.Index(string(out), delim)
	if idx < 0 {
		fatal("vcvarsall.bat output missing delimiter (batch likely failed)")
	}
	envBlock := string(out)[idx+len(delim):]
	for _, raw := range strings.Split(envBlock, "\n") {
		kv := strings.TrimRight(raw, "\r\n ")
		eq := strings.Index(kv, "=")
		if eq <= 0 {
			continue
		}
		os.Setenv(kv[:eq], kv[eq+1:])
	}
}

// findProjectRoot walks up from cwd looking for go.mod.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}
		dir = parent
	}
}

// libFileNames returns the expected library filenames for the current OS and
// link mode. On Windows, shared mode produces libmem.dll + libmem.lib (import
// library); static mode produces a single libmem.lib (archive with everything
// bundled by lib.exe).
func libFileNames(static bool) []string {
	switch runtime.GOOS {
	case "darwin":
		if static {
			return []string{"liblibmem.a"}
		}
		return []string{"liblibmem.dylib"}
	case "windows":
		if static {
			return []string{"libmem.lib"}
		}
		return []string{"libmem.dll", "libmem.lib"}
	default: // linux, freebsd
		if static {
			return []string{"liblibmem.a"}
		}
		return []string{"liblibmem.so"}
	}
}

// installLibrary searches buildDir for built library files and copies them to destDir.
func installLibrary(searchDir, destDir string, static bool) bool {
	found := false
	for _, name := range libFileNames(static) {
		src := findFile(searchDir, name)
		if src == "" {
			fmt.Printf("  Warning: %s not found in build output\n", name)
			continue
		}
		check(copyFile(src, filepath.Join(destDir, name)), "copying "+name)
		fmt.Printf("  Installed %s\n", name)
		found = true
	}
	return found
}

// installHeaders copies all .h/.hpp files from srcDir to destDir.
func installHeaders(srcDir, destDir string) {
	entries, err := os.ReadDir(srcDir)
	check(err, "reading header directory")
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".h" && ext != ".hpp" {
			continue
		}
		src := filepath.Join(srcDir, e.Name())
		dst := filepath.Join(destDir, e.Name())
		check(copyFile(src, dst), "copying "+e.Name())
		fmt.Printf("  Installed %s\n", e.Name())
	}
}

// findFile recursively searches dir for a file with the given name.
func findFile(dir, name string) string {
	var result string
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if d.Name() == name {
			result = path
			return filepath.SkipAll
		}
		return nil
	})
	return result
}

// copyFile copies a file preserving its permissions.
func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// runCmd runs a command with stdout/stderr connected to the terminal,
// echoing the command first so users can see exactly what is executing.
func runCmd(name string, args ...string) {
	fmt.Printf("  $ %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	check(cmd.Run(), fmt.Sprintf("running %s", name))
}

// runCmake runs cmake with a sanitized environment. We strip CMAKE_GENERATOR*
// env vars because they can silently override the -G flag we pass explicitly
// (e.g. a VS Developer Command Prompt or user-wide env can set
// CMAKE_GENERATOR=NMake Makefiles, producing a "generator mismatch" error in
// a freshly-created build directory).
func runCmake(args ...string) {
	fmt.Printf("  $ cmake %s\n", strings.Join(args, " "))
	cmd := exec.Command("cmake", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	env := os.Environ()
	filtered := env[:0]
	for _, v := range env {
		if strings.HasPrefix(v, "CMAKE_GENERATOR=") ||
			strings.HasPrefix(v, "CMAKE_GENERATOR_") {
			continue
		}
		filtered = append(filtered, v)
	}
	cmd.Env = filtered
	check(cmd.Run(), "running cmake")
}

func step(format string, args ...any) {
	fmt.Printf("==> "+format+"\n", args...)
}

func check(err error, context string) {
	if err != nil {
		fatal("%s: %v", context, err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
