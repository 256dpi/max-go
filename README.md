# maxgo

[![GoDoc](https://godoc.org/github.com/256dpi/maxgo?status.svg)](http://godoc.org/github.com/256dpi/maxgo)
[![Release](https://img.shields.io/github/release/256dpi/maxgo.svg)](https://github.com/256dpi/maxgo/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/256dpi/maxgo)](https://goreportcard.com/report/github.com/256dpi/maxgo)

**Toolkit for building Max externals with Go.**

## Installation

First you need to ensure you have recent version of [Go](https://golang.org) installed.

Then you can install the package using Go's module management:

```
go get github.com/256dpi/maxgo
```

This will also install the `maxgo` command line utility. 

## Usage

Add the following file to an empty directory:

```go
package main

import  "github.com/256dpi/maxgo"

type instance struct {
	in1   *maxgo.Inlet
	in2   *maxgo.Inlet
	out1  *maxgo.Outlet
	out2  *maxgo.Outlet
}

func (i *instance) Init(obj *maxgo.Object, args []maxgo.Atom) {
	// print to Max console
	maxgo.Pretty("init", args)

	// declare inlets
	i.in1 = obj.Inlet(maxgo.Any, "example inlet 1", true)
	i.in2 = obj.Inlet(maxgo.Float, "example inlet 2", false)

	// declare outlets
	i.out1 = obj.Outlet(maxgo.Any, "example outlet 1")
	i.out2 = obj.Outlet(maxgo.Bang, "example outlet 2")
}

func (i *instance) Handle(msg string, inlet int, data []maxgo.Atom) {
	// print to Max console
	maxgo.Pretty("handle", msg, inlet, data)

	// send to first outlet
	i.out1.Any(msg, data)
}

func (i *instance) Free() {
	// print to Max console
	maxgo.Pretty("free")
}

func init() {
	// initialize Max class
	maxgo.Register("example", &instance{})
}

func main() {
	// not called
}
```

Compile the external to the `out` directory:

```
maxgo -name example
```

You can also cross compile (macOS only) and install the external:

```
maxgo -name example -cross -install example
```
