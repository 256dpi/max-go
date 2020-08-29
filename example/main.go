package main

import (
	"github.com/256dpi/maxgo"
	"github.com/256dpi/maxgo/max"
)

type instance struct {
	in1  *max.Inlet
	in2  *max.Inlet
	out1 *max.Outlet
	out2 *max.Outlet
}

func (i *instance) Init(obj *max.Object, args []max.Atom) {
	max.Pretty("init", i, obj, args)

	i.in1 = obj.Inlet(max.Any, "example inlet 1", true)
	i.in2 = obj.Inlet(max.Any, "example inlet 2", false)

	i.out1 = obj.Outlet(max.Any, "example outlet 1")
	i.out2 = obj.Outlet(max.Any, "example outlet 2")
}

func (i *instance) Handle(msg string, inlet int, data []max.Atom) {
	max.Pretty("handle", i, msg, inlet, data)

	i.out1.Any(msg, data)
}

func (i *instance) Free() {
	max.Pretty("free", i)
}

func init() {
	maxgo.Init("maxgo", &instance{})
}

func main() {
	// not called
}
