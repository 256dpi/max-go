# max-go

[![GoDoc](https://godoc.org/github.com/256dpi/max-go?status.svg)](http://godoc.org/github.com/256dpi/max-go)
[![Release](https://img.shields.io/github/release/256dpi/max-go.svg)](https://github.com/256dpi/max-go/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/256dpi/max-go)](https://goreportcard.com/report/github.com/256dpi/max-go)

**Toolkit for building Max externals with Go.** 

## Installation

First you need to ensure you have recent version of [Go](https://golang.org) installed. On macOS simply install it using [brew](https://brew.sh):

```sh
brew install go
```

Then you can install the package and CLI using Go's module management:

```sh
go get -u github.com/256dpi/max-go
go get -u github.com/256dpi/max-go/cmd/maxgo
``` 

This will install the `maxgo` command line utility. You may need to add Go's `bin` directory tou your `PATH` variable to access the CLI in the terminal:

```sh
echo 'export PATH=~/go/bin:$PATH' >> ~/.zprofile # for zsh
```

Cross compilation on macOS for Windows additionally requires the `mingw-w64` toolchain:

```sh
brew install mingw-w64
```

## Usage

Add the following file to an empty directory:

```go
package main

import  "github.com/256dpi/max-go"

type instance struct {
	in1   *max.Inlet
	in2   *max.Inlet
	out1  *max.Outlet
	out2  *max.Outlet
}

func (i *instance) Init(obj *max.Object, args []max.Atom) {
	// print to Max console
	max.Pretty("init", args)

	// declare inlets
	i.in1 = obj.Inlet(max.Any, "example inlet 1", true)
	i.in2 = obj.Inlet(max.Float, "example inlet 2", false)

	// declare outlets
	i.out1 = obj.Outlet(max.Any, "example outlet 1")
	i.out2 = obj.Outlet(max.Bang, "example outlet 2")
}

func (i *instance) Handle(msg string, inlet int, data []max.Atom) {
	// print to Max console
	max.Pretty("handle", msg, inlet, data)

	// send to first outlet
	i.out1.Any(msg, data)
}

func (i *instance) Free() {
	// print to Max console
	max.Pretty("free")
}

func init() {
	// initialize Max class
	max.Register("example", &instance{})
}

func main() {
	// not called
}
```

Compile the external to the `dist` directory:

```
maxgo -name example -out dist
```

You can also cross compile (macOS only) and install the external:

```
maxgo -name example -out dist -cross -install example
```
