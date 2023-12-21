package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/otiai10/copy"
)

// On macOS, we need to make sure that we always write external binaries to a
// new file. Otherwise, kernel-side code-signing caches become outdated and Max
// crashes with a SIGKILL when loading the modified external.
// https://developer.apple.com/documentation/security/updating_mac_software

var name = flag.String("name", "", "the name of the external")
var out = flag.String("out", "out", "the output directory")
var cross = flag.Bool("cross", false, "cross compile for Windows on macOS")
var install = flag.String("install", "", "install into specified package")

func main() {
	// parse flags
	flag.Parse()

	// log
	fmt.Println("==> checking system...")

	// check go
	_, err := exec.LookPath("go")
	if err != nil {
		panic("missing go command (you may need to install Go)")
	}

	// check cross compile
	if *cross {
		// check OS
		if runtime.GOOS != "darwin" {
			panic("cannot cross compile for macOS from Windows")
		}

		// check zig
		_, err := exec.LookPath("zig")
		if err != nil {
			panic("missing zig command (you may need to install zig)")
		}
	}

	// log
	fmt.Println("==> preparing build...")

	// check name
	if *name == "" {
		panic("missing external name")
	}

	// get out dir
	var outDir = filepath.Join(".", "out")
	if *out != "" {
		outDir, err = filepath.Abs(*out)
		check(err)
	}

	// print
	fmt.Printf("name: %s\n", *name)
	fmt.Printf("out: %s\n", outDir)

	// clear directory (see top notes)
	check(os.RemoveAll(outDir))
	check(os.MkdirAll(outDir, os.ModePerm))

	// build
	switch runtime.GOOS {
	case "darwin":
		buildDarwin(outDir)
		if *cross {
			crossBuildWindows(outDir)
		}
	case "windows":
		buildWindows(outDir)
	}

	// install
	if *install != "" {
		// log
		fmt.Println("==> installing external...")

		// get home dir
		user, err := os.UserHomeDir()
		check(err)

		// prepare path
		dir, err := filepath.Abs(filepath.Join(user, "Documents", "Max 8", "Packages", *install, "externals"))
		check(err)

		// log
		fmt.Printf("target: %s\n", dir)

		// create path
		check(os.MkdirAll(dir, os.ModePerm))

		// copy external (see top notes)
		switch runtime.GOOS {
		case "darwin":
			check(os.RemoveAll(filepath.Join(dir, *name+".mxo")))
			check(copy.Copy(filepath.Join(outDir, *name+".mxo"), filepath.Join(dir, *name+".mxo")))
		case "windows":
			check(os.RemoveAll(filepath.Join(dir, *name+".mxe64")))
			check(copy.Copy(filepath.Join(outDir, *name+".mxe64"), filepath.Join(dir, *name+".mxe64")))
		}
	}

	// log
	fmt.Println("==> done!")
}

func buildDarwin(outDir string) {
	// log
	fmt.Println("==> building...")

	// prepare bin file
	bin := filepath.Join("out", *name)

	// build arm64 and amd64
	run("go",
		[]string{"build", "-v", "-buildmode=c-shared", "-o", bin + "-arm64"},
		[]string{"CGO_ENABLED=1", "GOARCH=arm64", "CGO_LDFLAGS=-Wl,-no_fixup_chains"},
	)
	run("go",
		[]string{"build", "-v", "-buildmode=c-shared", "-o", bin + "-amd64"},
		[]string{"CGO_ENABLED=1", "GOARCH=amd64", "CGO_LDFLAGS=-Wl,-no_fixup_chains"},
	)

	// assemble universal binary
	run("lipo",
		[]string{"-create", "-output", bin, bin + "-amd64", bin + "-arm64"},
		nil,
	)

	// ensure directory
	check(os.MkdirAll(filepath.Join(outDir, *name+".mxo", "Contents", "MacOS"), os.ModePerm))

	// copy binary
	check(os.Rename(bin, filepath.Join(outDir, *name+".mxo", "Contents", "MacOS", *name)))

	// write info plist
	check(ioutil.WriteFile(filepath.Join(outDir, *name+".mxo", "Contents", "Info.plist"), []byte(infoPlist(*name)), os.ModePerm))

	// write package info
	check(ioutil.WriteFile(filepath.Join(outDir, *name+".mxo", "Contents", "PkgInfo"), []byte(pkgInfo), os.ModePerm))
}

func buildWindows(outDir string) {
	// log
	fmt.Println("==> building...")

	// build
	run("go",
		[]string{"build", "-v", "-buildmode=c-shared", "-o", filepath.Join(outDir, *name+".mxe64")},
		[]string{"CGO_ENABLED=1"},
	)
}

func crossBuildWindows(outDir string) {
	// log
	fmt.Println("==> cross building...")

	// build
	run("go",
		[]string{"build", "-v", "-buildmode=c-shared", "-o", filepath.Join(outDir, *name+".mxe64")},
		[]string{`CC=zig cc -target x86_64-windows-gnu`, "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=1"},
	)
}

func run(bin string, args []string, env []string) {
	// construct
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Env = append(env, os.Environ()...)

	// run
	err := cmd.Run()
	if err != nil {
		panic("command failed")
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
