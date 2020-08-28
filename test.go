package main

import "C"

type instance struct {
	out Outlet
}

func (i *instance) Define(class Class) {
	Pretty("define", i, class)
}

func (i *instance) Init(object Object) {
	Pretty("init", i, object)

	object.AnyIn()
	i.out = object.AnyOut()
}

func (i *instance) Message(msg string, inlet int, atoms []Atom) {
	Pretty("message", i, msg, inlet, atoms)

	i.out.List(atoms)
}

func (i *instance) Free() {
	Pretty("free", i)
}

func test() {
	Register("maxgo", &instance{})

	// // basics
	// Log("logging...")
	// Error("errored...")
	// Alert("alerting...")
	//
	// // create class
	// var outlet Outlet
	// class := NewClass("maxgo", func(object Object) unsafe.Pointer {
	// 	Pretty("init", object)
	//
	// 	object.IntIn()
	// 	object.AnyIn()
	// 	object.IntOut()
	//
	// 	outlet = object.ListOut()
	//
	// 	o := unsafe.Pointer(&foo{})
	// 	Pretty(o)
	//
	// 	return o
	// }, func(ptr unsafe.Pointer, msg string, inlet int, atoms []Atom) {
	// 	Pretty("handler", ptr, msg, inlet, atoms)
	//
	// 	outlet.List(atoms)
	// }, func(pointer unsafe.Pointer) {
	// 	Pretty("free", pointer)
	// })
	//
	// // add methods
	// class.AddMethod("in1") // TODO: Relay to bridge_int
	// class.AddMethod("in2") // TODO: Relay to bridge_gimme
	//
	// // register
	// class.Register()
}
