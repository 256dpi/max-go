package maxgo

import (
	"reflect"
	"sync/atomic"

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
	// prepare id counter
	var id uint64

	// create instance map
	instances := map[uint64]Instance{}

	// get type
	typ := reflect.TypeOf(prototype).Elem()

	// initialize max class
	max.Init(name, func(o *max.Object, args []max.Atom) uint64 {
		// get reference
		ref := atomic.AddUint64(&id, 1)

		// create instance
		instance := reflect.New(typ).Interface().(Instance)

		// initialize
		instance.Init(o, args)

		// store instance
		instances[ref] = instance

		return ref
	}, func(ref uint64, msg string, inlet int, atoms []max.Atom) {
		// lookup instance
		instance := instances[ref]

		// handle message
		instance.Handle(msg, inlet, atoms)
	}, func(ref uint64) {
		// lookup instance
		instance := instances[ref]

		// free instance
		instance.Free()

		delete(instances, ref)
	})
}
