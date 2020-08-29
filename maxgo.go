package maxgo

import (
	"reflect"

	"github.com/256dpi/maxgo/max"
)

// Instance is a generic object instance.
type Instance interface {
	Init(obj *max.Object, args []max.Atom)
	Handle(msg string, inlet int, data []max.Atom)
	Free()
}

// Init will initialize the Max class using the provided instance. This function
// must be called from the main packages init() function as the main() function
// is never called by a Max external.
func Init(name string, prototype Instance) {
	// create instance map
	instances := map[*max.Object]Instance{}

	// get type
	typ := reflect.TypeOf(prototype).Elem()

	// initialize max class
	max.Init(name, func(o *max.Object, args []max.Atom) {
		instance := reflect.New(typ).Interface().(Instance)
		instance.Init(o, args)
		instances[o] = instance
	}, func(o *max.Object, msg string, inlet int, atoms []max.Atom) {
		instance := instances[o]
		instance.Handle(msg, inlet, atoms)
	}, func(o *max.Object) {
		instance := instances[o]
		instance.Free()
		delete(instances, o)
	})
}
