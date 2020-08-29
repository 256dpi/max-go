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

var name = flag.String("name", "", "the name of the external")
var cross = flag.Bool("cross", false, "cross compile for Windows on macOS")
var install = flag.String("install", "", "install into specified package")

func main() {
	// parse flags
	flag.Parse()

	// check name
	if *name == "" {
		panic("missing external name")
	}

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

		// check mingw
		_, err := exec.LookPath("x86_64-w64-mingw32-gcc")
		if err != nil {
			panic("missing x86_64-w64-mingw32-gcc command (you may need to install mingw-w64)")
		}
	}

	// build
	switch runtime.GOOS {
	case "darwin":
		buildDarwin()
		if *cross {
			crossBuildWindows()
		}
	case "windows":
		buildWindows()
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

		// create path
		check(os.MkdirAll(dir, os.ModePerm))

		// copy external
		switch runtime.GOOS {
		case "darwin":
			check(copy.Copy(filepath.Join(".", "out", *name+".mxo"), filepath.Join(dir, *name+".mxo")))
		case "windows":
			check(copy.Copy(filepath.Join(".", "out", *name+".mxe64"), filepath.Join(dir, *name+".mxe64")))
		}
	}

	// log
	fmt.Println("==> done!")
}

func buildDarwin() {
	// log
	fmt.Println("==> building...")

	// build
	run("go",
		[]string{"build", "-v", "-buildmode=c-shared", "-o", filepath.Join("out", *name)},
		[]string{"CGO_ENABLED=1"},
	)

	// ensure directory
	check(os.MkdirAll(filepath.Join(".", "out", *name+".mxo", "MacOS"), os.ModePerm))

	// copy binary
	check(os.Rename(filepath.Join(".", "out", *name), filepath.Join(".", "out", *name+".mxo", "MacOS", *name)))

	// write info plist
	check(ioutil.WriteFile(filepath.Join(".", "out", *name+".mxo", "Info.plist"), []byte(infoPlist(*name)), os.ModePerm))

	// write package info
	check(ioutil.WriteFile(filepath.Join(".", "out", *name+".mxo", "PkgInfo"), []byte(pkgInfo), os.ModePerm))
}

func buildWindows() {
	// log
	fmt.Println("==> building...")

	// build
	run("go",
		[]string{"build", "-v", "-buildmode=c-shared", "-o", filepath.Join("out", *name+".mxe64")},
		[]string{"CGO_ENABLED=1"},
	)
}

func crossBuildWindows() {
	// log
	fmt.Println("==> cross building...")

	// build
	run("go",
		[]string{"build", "-v", "-buildmode=c-shared", "-o", filepath.Join("out", *name+".mxe64")},
		[]string{"CC=x86_64-w64-mingw32-gcc", "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=1"},
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
