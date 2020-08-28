package maxgo

import (
	"reflect"

	"github.com/256dpi/maxgo/max"
)

// TODO: Cross compile and bundle CLI utility.

// Instance is a generic object instance.
type Instance interface {
	Init(max.Object)
	Message(string, int, []max.Atom)
	Free()
}

// Register will register a new class using the specified prototype instance.
func Register(name string, prototype Instance) {
	// create instance map
	instances := map[uintptr]Instance{}

	// get type
	typ := reflect.TypeOf(prototype).Elem()

	// create class
	class := max.NewClass(name, func(o max.Object) uintptr {
		// create instance
		value := reflect.New(typ)
		instance := value.Interface().(Instance)

		// initialize
		instance.Init(o)

		// store instance
		instances[value.Pointer()] = instance

		return value.Pointer()
	}, func(p uintptr, msg string, inlet int, atoms []max.Atom) {
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

	// TODO: Add methods?

	// register class
	class.Register()
}
