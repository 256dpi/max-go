package max

// #cgo CFLAGS: -I${SRCDIR}/../lib
// #cgo windows CFLAGS: -DWIN_VERSION=1
// #cgo darwin CFLAGS: -DMAC_VERSION=1
// #cgo darwin LDFLAGS: -Wl,-undefined,dynamic_lookup
// #cgo windows LDFLAGS: -L${SRCDIR}/../lib/x64 -lMaxAPI
// #include "max.h"
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/kr/pretty"
)

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

var objects = map[uint64]*Object{}

var initCallback func(*Object, []Atom) uint64
var handlerCallback func(uint64, string, int, []Atom)
var freeCallback func(uint64)

//export gomaxInit
func gomaxInit(ptr unsafe.Pointer, argc int64, argv *C.t_atom) (int, uint64) {
	// decode atoms
	atoms := decodeAtoms(argc, argv)

	// prepare object
	obj := &Object{ptr: ptr}

	// call init callback
	ref := initCallback(obj, atoms)

	// store object
	objects[ref] = obj

	// determine required proxies
	var proxies int
	if len(obj.in) > 0 {
		proxies = len(obj.in) - 1
	}

	return proxies, ref
}

//export gomaxMessage
func gomaxMessage(ref uint64, msg *C.char, inlet int64, argc int64, argv *C.t_atom) {
	// decode atoms
	atoms := decodeAtoms(argc, argv)

	// call handle if available
	if handlerCallback != nil {
		handlerCallback(ref, C.GoString(msg), int(inlet), atoms)
	}
}

//export gomaxInfo
func gomaxInfo(ref uint64, io, i int64) (*C.char, bool) {
	// get object
	obj, ok := objects[ref]
	if !ok {
		return nil, false
	}

	// return label
	if io == 1 {
		if int(i) < len(obj.in) {
			return C.CString(obj.in[i].label), obj.in[i].hot
		}
	} else {
		if int(i) < len(obj.out) {
			return C.CString(obj.out[i].label), false
		}
	}

	return nil, false
}

//export gomaxFree
func gomaxFree(ref uint64) {
	// call handler if available
	if freeCallback != nil {
		freeCallback(ref)
	}

	// delete object
	delete(objects, ref)
}

// Init will initialize the Max class with the specified name using the provided
// callbacks to initialize and free objects. This function must be called from
// the main packages init() function as the main() function is never called by a
// Max external.
func Init(name string, init func(*Object, []Atom) uint64, handler func(uint64, string, int, []Atom), free func(uint64)) {
	// set callbacks
	initCallback = init
	handlerCallback = handler
	freeCallback = free

	// initialize
	C.maxgo_init(C.CString(name))
}

/* Objects */

// Object is single Max object.
type Object struct {
	ptr unsafe.Pointer
	in  []*Inlet
	out []*Outlet
}

// Inlet is a single Max inlet.
type Inlet struct {
	typ   Type
	label string
	hot   bool
}

// Inlet will declare an inlet. If no inlets are added to an object it will have
// a default inlet to receive messages.
func (o *Object) Inlet(typ Type, label string, hot bool) *Inlet {
	// prepare
	inlet := &Inlet{typ: typ, label: label, hot: hot}

	// store
	o.in = append(o.in, inlet)

	return inlet
}

// Type will return the inlets type.
func (i *Inlet) Type() Type {
	return i.typ
}

// Label will return the inlets label.
func (i *Inlet) Label() string {
	return i.label
}

// Outlet is a single MAx outlet.
type Outlet struct {
	typ   Type
	ptr   unsafe.Pointer
	label string
}

// Outlet will declare an outlet.
func (o *Object) Outlet(typ Type, label string) *Outlet {
	// create outlet
	var ptr unsafe.Pointer
	switch typ {
	case Any:
		ptr = C.outlet_new(o.ptr, nil)
	case Bang:
		ptr = C.bangout(o.ptr)
	case Int:
		ptr = C.intout(o.ptr)
	case Float:
		ptr = C.floatout(o.ptr)
	case List:
		ptr = C.listout(o.ptr)
	default:
		panic("maxgo: invalid outlet type")
	}

	// prepare
	outlet := &Outlet{typ: typ, ptr: ptr, label: label}

	// store
	o.out = append(o.out, outlet)

	return outlet
}

// Type will return the outlets type.
func (o *Outlet) Type() Type {
	return o.typ
}

// Label will return the outlets label.
func (o *Outlet) Label() string {
	return o.label
}

// Any will send any message.
func (o *Outlet) Any(msg string, atoms []Atom) {
	if o.typ == Any {
		argc, argv := encodeAtoms(atoms)
		C.outlet_anything(o.ptr, C.gensym(C.CString(msg)), C.short(argc), argv)
	} else {
		Error("any sent to outlet of type %s", o.typ)
	}
}

// Bang will send a bang.
func (o *Outlet) Bang() {
	if o.typ == Bang || o.typ == Any {
		C.outlet_bang(o.ptr)
	} else {
		Error("bang sent to outlet of type %s", o.typ)
	}
}

// Int will send and int.
func (o *Outlet) Int(n int64) {
	if o.typ == Int || o.typ == Any {
		C.outlet_int(o.ptr, C.longlong(n))
	} else {
		Error("int sent to outlet of type %s", o.typ)
	}
}

// Float will send a float.
func (o *Outlet) Float(n float64) {
	if o.typ == Float || o.typ == Any {
		C.outlet_float(o.ptr, C.double(n))
	} else {
		Error("float sent to outlet of type %s", o.typ)
	}
}

// List will send a list.
func (o *Outlet) List(atoms []Atom) {
	if o.typ == List || o.typ == Any {
		argc, argv := encodeAtoms(atoms)
		C.outlet_list(o.ptr, nil, C.short(argc), argv)
	} else {
		Error("list sent to outlet of type %s", o.typ)
	}
}

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
