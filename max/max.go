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
	"sync"
	"sync/atomic"
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

var counter uint64
var objects = map[uint64]*Object{}

var initCallback func(*Object, []Atom)
var handlerCallback func(*Object, string, int, []Atom)
var freeCallback func(*Object)

//export gomaxInit
func gomaxInit(ptr unsafe.Pointer, argc int64, argv *C.t_atom) (int, uint64) {
	// decode atoms
	atoms := decodeAtoms(argc, argv)

	// get ref
	ref := atomic.AddUint64(&counter, 1)

	// prepare object
	obj := &Object{ref: ref, ptr: ptr}

	// store object
	objects[ref] = obj

	// call init callback
	initCallback(obj, atoms)

	// determine required proxies
	var proxies int
	if len(obj.in) > 0 {
		proxies = len(obj.in) - 1
	}

	// create outlets in reverse order
	for i := len(obj.out) - 1; i >= 0; i-- {
		outlet := obj.out[i]
		switch outlet.typ {
		case Any:
			outlet.ptr = C.outlet_new(obj.ptr, nil)
		case Bang:
			outlet.ptr = C.bangout(obj.ptr)
		case Int:
			outlet.ptr = C.intout(obj.ptr)
		case Float:
			outlet.ptr = C.floatout(obj.ptr)
		case List:
			outlet.ptr = C.listout(obj.ptr)
		default:
			panic("maxgo: invalid outlet type")
		}
	}

	return proxies, ref
}

//export gomaxMessage
func gomaxMessage(ref uint64, msg *C.char, inlet int64, argc int64, argv *C.t_atom) {
	// get object
	obj, ok := objects[ref]
	if !ok {
		return
	}

	// decode atoms
	atoms := decodeAtoms(argc, argv)

	// TODO: Check types?

	// call handle if available
	if handlerCallback != nil {
		handlerCallback(obj, C.GoString(msg), int(inlet), atoms)
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
	// get object
	obj, ok := objects[ref]
	if !ok {
		return
	}

	// call handler if available
	if freeCallback != nil {
		freeCallback(obj)
	}

	// delete object
	delete(objects, ref)
}

// Init will initialize the Max class with the specified name using the provided
// callbacks to initialize and free objects. This function must be called from
// the main packages init() function as the main() function is never called by a
// Max external.
func Init(name string, init func(*Object, []Atom), handler func(*Object, string, int, []Atom), free func(*Object)) {
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
	ref uint64
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
	inlet := &Inlet{typ: typ, label: label, hot: hot}
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
	label string
	ptr   unsafe.Pointer
}

// Outlet will declare an outlet.
func (o *Object) Outlet(typ Type, label string) *Outlet {
	outlet := &Outlet{typ: typ, label: label}
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

var queue = map[uint64]func(){}
var queueMutex sync.Mutex

// IsMainThread will return if the Max main thead is executing.
func IsMainThread() bool {
	return C.systhread_ismainthread() == 1
}

// IsTimerThread will return if the Max timer thread is executing.
func IsTimerThread() bool {
	return C.systhread_istimerthread() == 1
}

//export gomaxYield
func gomaxYield(ref uint64) {
	// get function
	queueMutex.Lock()
	fn := queue[ref]
	delete(queue, ref)
	queueMutex.Unlock()

	// execute function
	fn()
}

// Defer will run the provided function on the Max main thread.
func Defer(fn func()) {
	// get reference
	ref := atomic.AddUint64(&counter, 1)

	// store function
	queueMutex.Lock()
	queue[ref] = fn
	queueMutex.Unlock()

	// defer call
	C.maxgo_defer(C.ulonglong(ref))
}

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
