package main

import (
	"fmt"

	"github.com/256dpi/maxgo"
	"github.com/256dpi/maxgo/max"
)

type instance struct {
	out max.Outlet
}

func (i *instance) Init(obj max.Object, args []max.Atom) {
	max.Pretty("init", i, obj, args)

	obj.Inlet(max.Any)
	i.out = obj.Outlet(max.Any)
}

func (i *instance) Describe(inlet bool, num int) string {
	max.Pretty("describe", i, inlet, num)

	if inlet {
		return fmt.Sprintf("input #%d", num)
	} else {
		return fmt.Sprintf("output #%d", num)
	}
}

func (i *instance) Handle(msg string, inlet int, data []max.Atom) {
	max.Pretty("handle", i, msg, inlet, data)

	i.out.Any(msg, data)
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
