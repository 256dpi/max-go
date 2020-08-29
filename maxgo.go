package maxgo

import (
	"reflect"
	"sync/atomic"

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
	// prepare id counter
	var id uint64

	// create instance map
	instances := map[uint64]Instance{}

	// get type
	typ := reflect.TypeOf(prototype).Elem()

	// initialize max class
	max.Init(name, func(o max.Object, args []max.Atom) uint64 {
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
	}, func(ref uint64, io int64, i int64) string {
		// lookup instance
		instance := instances[ref]

		// describe port
		return instance.Describe(io == 1, int(i))
	}, func(ref uint64) {
		// lookup instance
		instance := instances[ref]

		// free instance
		instance.Free()

		delete(instances, ref)
	})
}
