package maxgo

import (
	"reflect"

	"github.com/256dpi/maxgo/max"
)

// TODO: Cross compile and bundle CLI utility.

// Instance is a generic object instance.
type Instance interface {
	Init(obj max.Object, args []max.Atom)
	Describe(inlet bool, num int) string
	Handle(msg string, inlet int, data []max.Atom)
	Free()
}

// Init will initialize the Max class using the provided instance.
func Init(name string, prototype Instance) {
	// create instance map
	instances := map[uintptr]Instance{}

	// get type
	typ := reflect.TypeOf(prototype).Elem()

	// initialize max class
	max.Init(name, func(o max.Object, args []max.Atom) uintptr {
		// create instance
		value := reflect.New(typ)
		instance := value.Interface().(Instance)

		// initialize
		instance.Init(o, args)

		// store instance
		instances[value.Pointer()] = instance

		return value.Pointer()
	}, func(p uintptr, msg string, inlet int, atoms []max.Atom) {
		// lookup instance
		instance := instances[p]

		// handle message
		instance.Handle(msg, inlet, atoms)
	}, func(p uintptr, io int64, i int64) string {
		// lookup instance
		instance := instances[p]

		// describe port
		return instance.Describe(io == 1, int(i))
	}, func(p uintptr) {
		// lookup instance
		instance := instances[p]

		// free instance
		instance.Free()

		delete(instances, p)
	})
}
