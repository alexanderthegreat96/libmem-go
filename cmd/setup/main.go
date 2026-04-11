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

	for _, arg := range os.Args[1:] {
		if arg == "--local" {
			local = true
		} else if arg == "--help" || arg == "-h" {
			printUsage()
			os.Exit(0)
		} else {
			version = arg
		}
	}

	fmt.Printf("Setting up libmem (%s) for %s/%s\n\n", version, runtime.GOOS, runtime.GOARCH)

	checkPrereqs()

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

	// Clone
	step("Cloning libmem (%s)...", version)
	runCmd("git", "clone", "--depth", "1", "--recurse-submodules",
		"--branch", version, repoURL, tmpDir+"/src")

	// Configure
	step("Configuring build...")
	check(os.MkdirAll(buildDir, 0755), "creating build directory")

	srcDir := filepath.Join(tmpDir, "src")
	cmakeArgs := []string{"-S", srcDir, "-B", buildDir, "-DLIBMEM_BUILD_TESTS=OFF"}
	if runtime.GOOS == "windows" {
		cmakeArgs = append(cmakeArgs, "-G", "MinGW Makefiles")
	}
	runCmd("cmake", cmakeArgs...)

	// Build
	step("Building (this may take a few minutes)...")
	runCmd("cmake", "--build", buildDir, "--config", "Release",
		"-j", fmt.Sprintf("%d", runtime.NumCPU()))

	// Copy artifacts
	step("Installing to %s...", installDir)
	check(os.MkdirAll(libDir, 0755), "creating lib directory")
	check(os.MkdirAll(incDir, 0755), "creating include directory")

	if !installLibrary(buildDir, libDir) {
		fatal("could not find built library in %s", buildDir)
	}
	installHeaders(filepath.Join(srcDir, "include", "libmem"), incDir)
	// tmpDir is cleaned up by defer

	fmt.Println()
	step("Done! libmem installed to %s", installDir)

	// Post-install hints
	if !local && runtime.GOOS == "linux" {
		fmt.Println("  You may need to run: sudo ldconfig")
	}
	if local {
		fmt.Println("  You can now build with: go build ./libmem")
	} else {
		fmt.Println("  You can now use: go get github.com/alexanderthegreat96/libmem-go")
	}
}

func printUsage() {
	fmt.Println(`Usage: setup [options] [version]

Options:
  --local    Install to libmem/deps/ in the current repo (for development)
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

// libFileNames returns the expected library filenames for the current OS.
func libFileNames() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"liblibmem.dylib"}
	case "windows":
		return []string{"liblibmem.dll", "liblibmem.dll.a"}
	default: // linux, freebsd
		return []string{"liblibmem.so"}
	}
}

// installLibrary searches buildDir for built library files and copies them to destDir.
func installLibrary(searchDir, destDir string) bool {
	found := false
	for _, name := range libFileNames() {
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

// runCmd runs a command with stdout/stderr connected to the terminal.
func runCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	check(cmd.Run(), fmt.Sprintf("running %s", name))
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
