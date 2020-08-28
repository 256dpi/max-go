package main

import (
	"reflect"
)

// Instance is a generic object instance.
type Instance interface {
	Define(Class)
	Init(Object)
	Message(string, int, []Atom)
	Free()
}

// Register will register a new class using the specified prototype instance.
func Register(name string, prototype Instance) {
	// create instance map
	instances := map[uintptr]Instance{}

	// get type
	typ := reflect.TypeOf(prototype).Elem()

	// create class
	class := NewClass(name, func(o Object) uintptr {
		// create instance
		value := reflect.New(typ)
		instance := value.Interface().(Instance)

		// initialize
		instance.Init(o)

		// store instance
		instances[value.Pointer()] = instance

		return value.Pointer()
	}, func(p uintptr, msg string, inlet int, atoms []Atom) {
		// lookup instance
		instance := instances[p]

		// send message
		instance.Message(msg, inlet, atoms)
	}, func(p uintptr) {
		// lookup instance
		instance := instances[p]

		// free instance
		instance.Free()

		delete(instances, p)
	})

	// define class
	prototype.Define(class)

	// register class
	class.Register()
}
