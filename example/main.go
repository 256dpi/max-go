package main

import (
	"C"

	"github.com/256dpi/maxgo"
	"github.com/256dpi/maxgo/max"
)

type instance struct {
	out max.Outlet
}

func (i *instance) Init(object max.Object) {
	max.Pretty("init", i, object)

	object.AnyIn()
	i.out = object.AnyOut()
}

func (i *instance) Message(msg string, inlet int, atoms []max.Atom) {
	max.Pretty("message", i, msg, inlet, atoms)

	i.out.List(atoms)
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
