package main

// #cgo CFLAGS: -I${SRCDIR}/max-sdk/source/c74support/max-includes
// #cgo LDFLAGS: -L${SRCDIR}/max-sdk/source/c74support/max-includes
// #cgo linux LDFLAGS: -Wl,-unresolved-symbols=ignore-all
// #cgo darwin LDFLAGS: -Wl,-undefined,dynamic_lookup
// #include "api.h"
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/kr/pretty"
)

/* Types */

// Atom is a Max a of type int64, float64 or string.
type Atom = interface{}

/* Basic */

// Log will print a message to the max console.
func Log(format string, args ...interface{}) {
	C.maxgo_log(C.CString(fmt.Sprintf(format, args...)))
}

// Error will print an error to the max console.
func Error(format string, args ...interface{}) {
	C.maxgo_error(C.CString(fmt.Sprintf(format, args...)))
}

// Alert will show an alert dialog.
func Alert(format string, args ...interface{}) {
	C.maxgo_alert(C.CString(fmt.Sprintf(format, args...)))
}

// Pretty will pretty print the provided values.
func Pretty(a ...interface{}) {
	Log(pretty.Sprint(a...))
}

/* Classes */

var classes = map[string]Class{}
var initMethods = map[string]func(Object) unsafe.Pointer{}
var handlers = map[string]func(unsafe.Pointer, string, int, []Atom){}
var freeMethods = map[string]func(unsafe.Pointer){}

//export gomaxGet
func gomaxGet(name *C.char) *C.t_class {
	return classes[C.GoString(name)].raw
}

//export gomaxInit
func gomaxInit(name *C.char, obj unsafe.Pointer) unsafe.Pointer {
	return initMethods[C.GoString(name)](Object{obj})
}

//export gomaxMessage
func gomaxMessage(name *C.char, ptr unsafe.Pointer, msg *C.char, inlet int64, argc int64, argv *C.t_atom) {
	atoms := decodeAtoms(argc, argv)
	handler := handlers[C.GoString(name)]
	if handler != nil {
		handler(ptr, C.GoString(msg), int(inlet), atoms)
	}
}

//export gomaxFree
func gomaxFree(name *C.char, ptr unsafe.Pointer) {
	free := freeMethods[C.GoString(name)]
	if free != nil {
		free(ptr)
	}
}

// Class is a Max object class.
type Class struct {
	raw *C.t_class
	reg bool
}

// NewClass will create a new class with the specified name using the provided callbacks to initialize and free objects.
func NewClass(name string, init func(Object) unsafe.Pointer, handler func(unsafe.Pointer, string, int, []Atom), free func(unsafe.Pointer)) Class {
	// register methods
	initMethods[name] = init
	handlers[name] = handler
	freeMethods[name] = free

	// create class
	class := Class{raw: C.maxgo_class_new(C.CString(name))}

	// register class
	classes[name] = class

	return class
}

// AddMethod will add a method with the specified name.
func (c Class) AddMethod(name string) {
	// check
	if c.reg {
		panic("maxgo: class already registered")
	}

	// add method
	C.maxgo_class_add_method(c.raw, C.CString(name))
}

// Register will register the class if not already registered.
func (c Class) Register() {
	// check
	if c.reg {
		panic("maxgo: class already registered")
	}

	// register class
	C.class_register(C.CLASS_BOX, c.raw)
}

/* Objects */

// Object is single Max object.
type Object struct {
	raw unsafe.Pointer
}

// Inlet is a single Max inlet.
type Inlet struct {
	raw unsafe.Pointer
}

// Outlet is a single MAx outlet.
type Outlet struct {
	raw unsafe.Pointer
}

// TODO: Object log (post), error and warn.

// AnyIn will create a generic inlet.
func (o *Object) AnyIn() Inlet {
	return Inlet{C.inlet_new(o.raw, nil)}
}

// BangIn will create a bang inlet.
func (o *Object) BangIn() Inlet {
	return Inlet{C.inlet_new(o.raw, C.CString("bang"))}
}

// IntIn will create an int inlet.
func (o *Object) IntIn() Inlet {
	return Inlet{C.intin(o.raw, C.short(1))}
}

// FloatIn will create a float inlet.
func (o *Object) FloatIn() Inlet {
	return Inlet{C.floatin(o.raw, C.short(1))}
}

// AnyOut will create a generic outlet.
func (o *Object) AnyOut() Outlet {
	return Outlet{C.outlet_new(o.raw, nil)}
}

// BangOut will create a bang outlet.
func (o *Object) BangOut() Outlet {
	return Outlet{C.bangout(o.raw)}
}

// IntOut will create an int outlet.
func (o *Object) IntOut() Outlet {
	return Outlet{C.intout(o.raw)}
}

// FloatOut will create a float outlet.
func (o *Object) FloatOut() Outlet {
	return Outlet{C.floatout(o.raw)}
}

// ListOut will create a list outlet.
func (o *Object) ListOut() Outlet {
	return Outlet{C.listout(o.raw)}
}

// Bang will send a bang.
func (o Outlet) Bang() {
	C.outlet_bang(o.raw)
}

// Int will send and int.
func (o Outlet) Int(n int64) {
	C.outlet_int(o.raw, C.longlong(n))
}

// Float will send a float.
func (o Outlet) Float(n float64) {
	C.outlet_float(o.raw, C.double(n))
}

// List will send a list.
func (o Outlet) List(atoms []Atom) {
	argc, argv := encodeAtoms(atoms)
	C.outlet_list(o.raw, nil, C.short(argc), argv)
}

// TODO: Support proxies?

/* Threads */

// TODO: Add critical regions management.

// func EnterCritical() {
// 	C.critical_enter(0)
// }
//
// func ExitCritical() {
// 	C.critical_exit(0)
// }

/* Atoms */

func decodeAtoms(argc int64, argv *C.t_atom) []Atom {
	// cast to slice
	var list []C.t_atom
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&list))
	sliceHeader.Cap = int(argc)
	sliceHeader.Len = int(argc)
	sliceHeader.Data = uintptr(unsafe.Pointer(argv))

	// allocate result
	atoms := make([]interface{}, 0, len(list))

	// add atoms
	for _, item := range list {
		switch item.a_type {
		case C.A_LONG:
			atoms = append(atoms, int64(C.atom_getlong(&item)))
		case C.A_FLOAT:
			atoms = append(atoms, float64(C.atom_getfloat(&item)))
		case C.A_SYM:
			atoms = append(atoms, C.GoString(C.atom_getsym(&item).s_name))
		default:
			atoms = append(atoms, nil)
		}
	}

	return atoms
}

func encodeAtoms(atoms []Atom) (argc int64, argv *C.t_atom) {
	// allocate atom array
	array := make([]C.t_atom, len(atoms))

	// set atoms
	for i, atom := range atoms {
		switch atom := atom.(type) {
		case int64:
			C.atom_setlong(&array[i], C.longlong(atom))
		case float64:
			C.atom_setfloat(&array[i], C.double(atom))
		case string:
			C.atom_setsym(&array[i], C.gensym(C.CString(atom)))
		}
	}

	return int64(len(atoms)), &array[0]
}
