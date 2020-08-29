package maxgo

import (
	"reflect"
	"sync"

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
//
// The callbacks on the instance are never called from in parallel.
func Init(name string, prototype Instance) {
	// create mutex
	var mutex sync.Mutex

	// create instance map
	instances := map[*max.Object]Instance{}

	// get type
	typ := reflect.TypeOf(prototype).Elem()

	// initialize max class
	max.Init(name, func(obj *max.Object, args []max.Atom) {
		// allocate instance
		instance := reflect.New(typ).Interface().(Instance)

		// call init
		obj.Lock()
		instance.Init(obj, args)
		obj.Unlock()

		// store instance
		mutex.Lock()
		instances[obj] = instance
		mutex.Unlock()
	}, func(obj *max.Object, msg string, inlet int, atoms []max.Atom) {
		// get instance
		mutex.Lock()
		instance := instances[obj]
		mutex.Unlock()

		// handle message
		obj.Lock()
		instance.Handle(msg, inlet, atoms)
		obj.Unlock()
	}, func(obj *max.Object) {
		// get and delete instance
		mutex.Lock()
		instance := instances[obj]
		delete(instances, obj)
		mutex.Unlock()

		// call free
		obj.Lock()
		instance.Free()
		obj.Unlock()
	})
}
