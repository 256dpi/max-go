package maxgo

import (
	"reflect"
	"sync"
)

// Instance is a generic object instance.
type Instance interface {
	Init(obj *Object, args []Atom) bool
	Handle(msg string, inlet int, data []Atom)
	Free()
}

// Register will initialize the Max class using the provided instance. This
// function must be called from the main packages init() function as the main()
// function is never called by a Max external.
//
// The callbacks on the instance are never called in parallel.
func Register(name string, prototype Instance) {
	// create mutex
	var mutex sync.Mutex

	// create instance map
	instances := map[*Object]Instance{}

	// get type
	typ := reflect.TypeOf(prototype).Elem()

	// initialize max class
	Init(name, func(obj *Object, args []Atom) bool {
		// allocate instance
		instance := reflect.New(typ).Interface().(Instance)

		// call init
		obj.Lock()
		ok := instance.Init(obj, args)
		obj.Unlock()

		// return if not ok
		if !ok {
			return false
		}

		// store instance
		mutex.Lock()
		instances[obj] = instance
		mutex.Unlock()

		return true
	}, func(obj *Object, msg string, inlet int, atoms []Atom) {
		// get instance
		mutex.Lock()
		instance := instances[obj]
		mutex.Unlock()

		// return if nil
		if instance == nil {
			return
		}

		// handle message
		obj.Lock()
		instance.Handle(msg, inlet, atoms)
		obj.Unlock()
	}, func(obj *Object) {
		// get and delete instance
		mutex.Lock()
		instance := instances[obj]
		delete(instances, obj)
		mutex.Unlock()

		// return if nil
		if instance == nil {
			return
		}

		// call free
		obj.Lock()
		instance.Free()
		obj.Unlock()
	})
}
