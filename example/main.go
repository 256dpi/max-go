package main

import (
	"fmt"

	"C"

	"github.com/256dpi/maxgo"
	"github.com/256dpi/maxgo/max"
)

type instance struct {
	out max.Outlet
}

func (i *instance) Init(obj max.Object, args []max.Atom) {
	max.Pretty("init", i, obj, args)

	obj.AnyIn()
	i.out = obj.AnyOut()
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

	i.out.List(data)
}

func (i *instance) Free() {
	max.Pretty("free", i)
}

//export ext_main
func ext_main(uintptr) {
	main()
}

func main() {
	maxgo.Register("maxgo", &instance{})
}
