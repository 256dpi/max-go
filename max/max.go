package max

// #cgo CFLAGS: -I${SRCDIR}/../max-sdk/source/c74support/max-includes
// #cgo windows CFLAGS: -DWIN_VERSION=1
// #cgo darwin CFLAGS: -DMAC_VERSION=1
// #cgo darwin LDFLAGS: -Wl,-undefined,dynamic_lookup
// #cgo windows LDFLAGS: -L${SRCDIR}/../max-sdk/source/c74support/max-includes/x64 -lMaxAPI
// #include "max.h"
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/kr/pretty"
)

// TODO: Support hot and cold inlets.
//  https://cycling74.com/sdk/max-sdk-7.3.3/html/chapter_enhancements.html

/* Types */

// Type describes an inlet or outlet type.
type Type string

// The available inlet and outlet types.
const (
	Any   Type = "any"
	Bang  Type = "bang"
	Int   Type = "int"
	Float Type = "float"
	List  Type = "list"
)

// Atom is a Max atom of type int64, float64 or string.
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
var initializers = map[string]func(Object, []Atom) uintptr{}
var handlers = map[string]func(uintptr, string, int, []Atom){}
var assists = map[string]func(uintptr, int64, int64) string{}
var finalizers = map[string]func(uintptr){}

//export gomaxGet
func gomaxGet(name *C.char) *C.t_class {
	return (*C.t_class)(classes[C.GoString(name)].ptr)
}

//export gomaxInit
func gomaxInit(name *C.char, obj unsafe.Pointer, argc int64, argv *C.t_atom) uintptr {
	atoms := decodeAtoms(argc, argv)
	return initializers[C.GoString(name)](Object{ptr: obj}, atoms)
}

//export gomaxMessage
func gomaxMessage(name *C.char, ptr uintptr, msg *C.char, inlet int64, argc int64, argv *C.t_atom) {
	atoms := decodeAtoms(argc, argv)
	handler := handlers[C.GoString(name)]
	if handler != nil {
		handler(ptr, C.GoString(msg), int(inlet), atoms)
	}
}

//export gomaxAssist
func gomaxAssist(name *C.char, ptr uintptr, io, i int64) *C.char {
	assist := assists[C.GoString(name)]
	if assist != nil {
		return C.CString(assist(ptr, io, i))
	}
	return C.CString("")
}

//export gomaxFree
func gomaxFree(name *C.char, ptr uintptr) {
	free := finalizers[C.GoString(name)]
	if free != nil {
		free(ptr)
	}
}

// Class is a Max object class.
type Class struct {
	ptr unsafe.Pointer
	reg bool
}

// NewClass will create a new class with the specified name using the provided callbacks to initialize and free objects.
func NewClass(name string, init func(Object, []Atom) uintptr, handler func(uintptr, string, int, []Atom), assist func(uintptr, int64, int64) string, free func(uintptr)) Class {
	// register methods
	initializers[name] = init
	handlers[name] = handler
	assists[name] = assist
	finalizers[name] = free

	// create class
	class := Class{ptr: unsafe.Pointer(C.maxgo_class_new(C.CString(name)))}

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
	C.maxgo_class_add_method((*C.t_class)(c.ptr), C.CString(name))
}

// Register will register the class if not already registered.
func (c Class) Register() {
	// check
	if c.reg {
		panic("maxgo: class already registered")
	}

	// register class
	C.class_register(C.CLASS_BOX, (*C.t_class)(c.ptr))
}

/* Objects */

// Object is single Max object.
type Object struct {
	ptr unsafe.Pointer
}

// Inlet is a single Max inlet.
type Inlet struct {
	ptr unsafe.Pointer
}

// Outlet is a single MAx outlet.
type Outlet struct {
	typ Type
	ptr unsafe.Pointer
}

// Inlet will declare an inlet.
func (o *Object) Inlet(typ Type) Inlet {
	switch typ {
	case Any:
		return Inlet{ptr: C.inlet_new(o.ptr, nil)}
	case Bang:
		return Inlet{ptr: C.inlet_new(o.ptr, C.CString("bang"))}
	case Int:
		return Inlet{ptr: C.intin(o.ptr, C.short(1))}
	case Float:
		return Inlet{ptr: C.floatin(o.ptr, C.short(1))}
	case List:
		return Inlet{ptr: C.inlet_new(o.ptr, C.CString("list"))}
	default:
		panic("invalid inlet type")
	}
}

// Outlet will declare an outlet.
func (o *Object) Outlet(typ Type) Outlet {
	switch typ {
	case Any:
		return Outlet{typ: typ, ptr: C.outlet_new(o.ptr, nil)}
	case Bang:
		return Outlet{typ: typ, ptr: C.bangout(o.ptr)}
	case Int:
		return Outlet{typ: typ, ptr: C.intout(o.ptr)}
	case Float:
		return Outlet{typ: typ, ptr: C.floatout(o.ptr)}
	case List:
		return Outlet{typ: typ, ptr: C.listout(o.ptr)}
	default:
		panic("invalid outlet type")
	}
}

// Any will send any message.
func (o Outlet) Any(msg string, atoms []Atom) {
	if o.typ == Any {
		argc, argv := encodeAtoms(atoms)
		C.outlet_anything(o.ptr, C.gensym(C.CString(msg)), C.short(argc), argv)
	} else {
		Error("any sent to outlet of type %s", o.typ)
	}
}

// Bang will send a bang.
func (o Outlet) Bang() {
	if o.typ == Bang || o.typ == Any {
		C.outlet_bang(o.ptr)
	} else {
		Error("bang sent to outlet of type %s", o.typ)
	}
}

// Int will send and int.
func (o Outlet) Int(n int64) {
	if o.typ == Int || o.typ == Any {
		C.outlet_int(o.ptr, C.longlong(n))
	} else {
		Error("int sent to outlet of type %s", o.typ)
	}
}

// Float will send a float.
func (o Outlet) Float(n float64) {
	if o.typ == Float || o.typ == Any {
		C.outlet_float(o.ptr, C.double(n))
	} else {
		Error("float sent to outlet of type %s", o.typ)
	}
}

// List will send a list.
func (o Outlet) List(atoms []Atom) {
	if o.typ == List || o.typ == Any {
		argc, argv := encodeAtoms(atoms)
		C.outlet_list(o.ptr, nil, C.short(argc), argv)
	} else {
		Error("list sent to outlet of type %s", o.typ)
	}
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
	// check length
	if len(atoms) == 0 {
		return 0, nil
	}

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
