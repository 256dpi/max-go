package max

import (
	"reflect"
	"sync"
)

// Instance is a generic object instance.
type Instance interface {
	Init(obj *Object, args []Atom) bool
	Handle(inlet int, msg string, data []Atom)
	Process(input [][]float64, output [][]float64)
	Free()
}

// AdvancedInstance is an object that responds to advanced messages.
type AdvancedInstance interface {
	Loaded()
	DoubleClicked()
}

// Register will initialize the Max class using the provided instance. This
// function must be called from the main packages main() function. The instance
// methods are usually called on the Max main thread. However, the handler may
// be called from an unknown thread in parallel to the other callbacks.
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
		ok := instance.Init(obj, args)
		if !ok {
			return false
		}

		// store instance
		mutex.Lock()
		instances[obj] = instance
		mutex.Unlock()

		return true
	}, func(obj *Object, inlet int, msg string, atoms []Atom) {
		// get instance
		mutex.Lock()
		instance := instances[obj]
		mutex.Unlock()

		// return if nil
		if instance == nil {
			return
		}

		// handle loadbang and dblclick
		if msg == "loadbang" || msg == "dblclick" {
			if adv, ok := instance.(AdvancedInstance); ok {
				switch msg {
				case "loadbang":
					adv.Loaded()
				case "dblclick":
					adv.DoubleClicked()
				}
			}
			return
		}

		// handle message
		instance.Handle(inlet, msg, atoms)
	}, func(obj *Object, input, output [][]float64) {
		// get instance
		mutex.Lock()
		instance := instances[obj]
		mutex.Unlock()

		// return if nil
		if instance == nil {
			return
		}

		// process audio
		instance.Process(input, output)
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
		instance.Free()
	})
}
