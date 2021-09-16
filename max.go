package max

// #cgo CFLAGS: -I${SRCDIR}/lib/max -I${SRCDIR}/lib/msp
// #cgo windows CFLAGS: -DWIN_VERSION=1
// #cgo darwin CFLAGS: -DMAC_VERSION=1
// #cgo darwin LDFLAGS: -Wl,-undefined,dynamic_lookup
// #cgo windows LDFLAGS: -L${SRCDIR}/lib/max/x64 -L${SRCDIR}/lib/msp/x64 -lMaxAPI -lMaxAudio
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
	Bang   Type = "bang"
	Int    Type = "int"
	Float  Type = "float"
	List   Type = "list"
	Any    Type = "any"
	Signal Type = "signal"
)

func (t Type) enum() C.maxgo_type_e {
	switch t {
	case Bang:
		return C.MAXGO_BANG
	case Int:
		return C.MAXGO_INT
	case Float:
		return C.MAXGO_FLOAT
	case List:
		return C.MAXGO_LIST
	case Any:
		return C.MAXGO_ANY
	case Signal:
		return C.MAXGO_SIGNAL
	default:
		panic("invalid type")
	}
}

// Atom is a Max atom of type int64, float64 or string.
type Atom = interface{}

// Event describes an emitted event.
type Event struct {
	Outlet *Outlet
	Type   Type
	Msg    string
	Data   []Atom
}

/* Basic */

// Log will print a message to the max console.
func Log(format string, args ...interface{}) {
	C.maxgo_log(C.CString(fmt.Sprintf(format, args...))) // string freed by receiver
}

// Error will print an error to the max console.
func Error(format string, args ...interface{}) {
	C.maxgo_error(C.CString(fmt.Sprintf(format, args...))) // string freed by receiver
}

// Alert will show an alert dialog.
func Alert(format string, args ...interface{}) {
	C.maxgo_alert(C.CString(fmt.Sprintf(format, args...))) // string freed by receiver
}

// Pretty will pretty print and log the provided values.
func Pretty(a ...interface{}) {
	Log(pretty.Sprint(a...))
}

var symbols sync.Map

func gensym(str string) *C.t_symbol {
	// check cache
	val, ok := symbols.Load(str)
	if ok {
		return val.(*C.t_symbol)
	}

	// get and cache symbol
	sym := C.maxgo_gensym(C.CString(str)) // string freed by receiver
	symbols.Store(str, sym)

	return sym
}

/* Initialization */

// InitCallback is called to initialize objects.
type InitCallback func(obj *Object, atoms []Atom) bool

// HandleCallback is called to handle messages.
type HandleCallback func(obj *Object, inlet int, name string, atoms []Atom)

// ProcessCallback is called to process audio.
type ProcessCallback func(obj *Object, input, output []float64)

// FreeCallback is called to free objects.
type FreeCallback func(obj *Object)

var initCallback InitCallback
var handleCallback HandleCallback
var processCallback ProcessCallback
var freeCallback FreeCallback

var initMutex sync.Mutex
var initDone bool

//go:linkname mainMain main.main
func mainMain()

//export maxgoMain
func maxgoMain() {
	// call main
	mainMain()

	// acquire mutex
	initMutex.Lock()
	defer initMutex.Unlock()

	// check flag
	if !initDone {
		panic("not initialized")
	}
}

// Init will initialize the Max class with the specified name using the provided
// callbacks to initialize and free objects. This function must be called from
// the main packages main() function.
//
// The provided callbacks are called to initialize and object, handle messages,
// process audio and free the object when it is not used anymore. The callbacks
// are usually called on the Max main thread. However, the handler may be called
// from an unknown thread in parallel to the other callbacks.
func Init(name string, init InitCallback, handle HandleCallback, process ProcessCallback, free FreeCallback) {
	// ensure mutex
	initMutex.Lock()
	defer initMutex.Unlock()

	// check flag
	if initDone {
		panic("already initialized")
	}

	// set callbacks
	initCallback = init
	handleCallback = handle
	processCallback = process
	freeCallback = free

	// initialize
	C.maxgo_init(C.CString(name)) // string freed by receiver

	// set flag
	initDone = true
}

/* Classes */

var counter uint64

var objects = map[uint64]*Object{}
var objectsMutex sync.Mutex

//export maxgoInit
func maxgoInit(ptr unsafe.Pointer, argc int64, argv *C.t_atom) (uint64, int, int) {
	// decode atoms
	atoms := decodeAtoms(argc, argv)

	// get ref
	ref := atomic.AddUint64(&counter, 1)

	// prepare object
	obj := &Object{
		ref:   ref,
		ptr:   ptr,
		queue: make(chan Event, 100),
	}

	// store object
	objectsMutex.Lock()
	objects[ref] = obj
	objectsMutex.Unlock()

	// call init callback
	ok := initCallback(obj, atoms)
	if !ok {
		return 0, 0, 0
	}

	// determine required proxies and signals
	var proxies int
	var signals int
	for _, inlet := range obj.in {
		if inlet.Type() == Signal {
			signals++
		} else {
			proxies++
		}
	}
	if signals == 0 && proxies > 0 {
		proxies--
	}

	// create outlets in reverse order
	for i := len(obj.out) - 1; i >= 0; i-- {
		outlet := obj.out[i]
		switch outlet.typ {
		case Bang:
			outlet.ptr = C.bangout(obj.ptr)
		case Int:
			outlet.ptr = C.intout(obj.ptr)
		case Float:
			outlet.ptr = C.floatout(obj.ptr)
		case List:
			outlet.ptr = C.listout(obj.ptr)
		case Any:
			outlet.ptr = C.outlet_new(obj.ptr, nil)
		case Signal:
			str := C.CString("signal")
			outlet.ptr = C.outlet_new(obj.ptr, str)
			C.free(unsafe.Pointer(str))
		default:
			panic("invalid outlet type")
		}
	}

	return ref, proxies, signals
}

//export maxgoHandle
func maxgoHandle(ref uint64, msg *C.char, inlet int64, argc int64, argv *C.t_atom) {
	// get object
	objectsMutex.Lock()
	obj, ok := objects[ref]
	objectsMutex.Unlock()
	if !ok {
		return
	}

	// decode atoms
	atoms := decodeAtoms(argc, argv)

	// get name
	name := C.GoString(msg)

	// check inlet
	if inlet >= 0 {
		// get inlet
		in := obj.in[inlet]
		if in == nil {
			return
		}

		// check signal
		if in.typ == Signal {
			Error("message received on signal inlet %d", inlet)
			return
		}

		// check name
		if in.typ != Any && Type(name) != in.typ {
			Error("invalid message received on inlet %d", inlet)
			return
		}

		// check atoms
		if in.typ == Bang && len(atoms) != 0 || (in.typ == Int || in.typ == Float) && len(atoms) != 1 {
			Error("unexpected input received on inlet %d", inlet)
			return
		}

		// check types
		switch in.typ {
		case Int:
			if _, ok := atoms[0].(int64); !ok {
				Error("invalid input received on inlet %d", inlet)
				return
			}
		case Float:
			if _, ok := atoms[0].(float64); !ok {
				Error("invalid input received on inlet %d", inlet)
				return
			}
		}
	}

	// run callback if available
	if handleCallback != nil {
		handleCallback(obj, int(inlet), name, atoms)
	}
}

//export maxgoProcess
func maxgoProcess(ref uint64, in, out *C.double, samples C.long) {
	// get object
	objectsMutex.Lock()
	obj, ok := objects[ref]
	objectsMutex.Unlock()
	if !ok {
		return
	}

	// prepare input and output
	var input []float64
	var output []float64

	// convert input
	if in != nil {
		input = []float64{}
		sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&input))
		sliceHeader.Cap = int(samples)
		sliceHeader.Len = int(samples)
		sliceHeader.Data = uintptr(unsafe.Pointer(in))
	}

	// convert output
	if out != nil {
		output = []float64{}
		sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&output))
		sliceHeader.Cap = int(samples)
		sliceHeader.Len = int(samples)
		sliceHeader.Data = uintptr(unsafe.Pointer(out))
	}

	// run callback if available
	if processCallback != nil {
		processCallback(obj, input, output)
	}
}

//export maxgoPop
func maxgoPop(ref uint64) (unsafe.Pointer, C.maxgo_type_e, *C.t_symbol, int64, *C.t_atom, bool) {
	// get object
	objectsMutex.Lock()
	obj, ok := objects[ref]
	objectsMutex.Unlock()
	if !ok {
		return nil, 0, nil, 0, nil, false
	}

	// get event
	var evt Event
	select {
	case evt = <-obj.queue:
	default:
		return nil, 0, nil, 0, nil, false
	}

	// encode atoms
	argc, argv := encodeAtoms(evt.Data)

	// get symbol if available
	var sym *C.t_symbol
	if evt.Type == Any {
		sym = gensym(evt.Msg)
	}

	// determine if there are more events
	more := len(obj.queue) > 0

	return evt.Outlet.ptr, evt.Type.enum(), sym, argc, argv, more
}

//export maxgoDescribe
func maxgoDescribe(ref uint64, io, i int64) (*C.char, bool) {
	// get object
	objectsMutex.Lock()
	obj, ok := objects[ref]
	objectsMutex.Unlock()
	if !ok {
		return nil, false
	}

	// return label
	if io == 1 {
		if int(i) < len(obj.in) {
			label := fmt.Sprintf("%s (%s)", obj.in[i].label, obj.in[i].typ)
			return C.CString(label), obj.in[i].hot // string freed by receiver
		}
	} else {
		if int(i) < len(obj.out) {
			label := fmt.Sprintf("%s (%s)", obj.out[i].label, obj.out[i].typ)
			return C.CString(label), false // string freed by receiver
		}
	}

	return nil, false
}

//export maxgoFree
func maxgoFree(ref uint64) {
	// get and delete object
	objectsMutex.Lock()
	obj, ok := objects[ref]
	delete(objects, ref)
	objectsMutex.Unlock()
	if !ok {
		return
	}

	// run callback if available
	if freeCallback != nil {
		freeCallback(obj)
	}
}

/* Objects */

// Object is single Max object.
type Object struct {
	ref   uint64
	ptr   unsafe.Pointer
	in    []*Inlet
	out   []*Outlet
	queue chan Event
}

// Push will add the provided events to the objects queue.
func (o *Object) Push(events ...Event) {
	// queue events
	for _, evt := range events {
		select {
		case o.queue <- evt:
		default:
			Error("dropped event due to full queue")
		}
	}

	// notify
	C.maxgo_notify(o.ptr)
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
	// check signal
	if typ == Signal && len(o.in) > 0 {
		panic("signal only supported as the first inlet")
	}

	// create inlet
	inlet := &Inlet{typ: typ, label: label, hot: hot}

	// store inlet
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
	obj   *Object
	typ   Type
	label string
	ptr   unsafe.Pointer
}

// Outlet will declare an outlet.
func (o *Object) Outlet(typ Type, label string) *Outlet {
	// check signal
	if typ == Signal && len(o.out) > 0 {
		panic("signal only supported as the first outlet")
	}

	// create outlet
	outlet := &Outlet{obj: o, typ: typ, label: label}

	// store outlet
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

// Bang will send a bang.
func (o *Outlet) Bang() {
	if o.typ == Bang || o.typ == Any {
		o.obj.Push(Event{Outlet: o, Type: Bang})
	} else {
		Error("bang sent to outlet of type %s", o.typ)
	}
}

// Int will send and int.
func (o *Outlet) Int(n int64) {
	if o.typ == Int || o.typ == Any {
		o.obj.Push(Event{Outlet: o, Type: Int, Data: []Atom{n}})
	} else {
		Error("int sent to outlet of type %s", o.typ)
	}
}

// Float will send a float.
func (o *Outlet) Float(n float64) {
	if o.typ == Float || o.typ == Any {
		o.obj.Push(Event{Outlet: o, Type: Float, Data: []Atom{n}})
	} else {
		Error("float sent to outlet of type %s", o.typ)
	}
}

// List will send a list.
func (o *Outlet) List(atoms []Atom) {
	if o.typ == List || o.typ == Any {
		o.obj.Push(Event{Outlet: o, Type: List, Data: atoms})
	} else {
		Error("list sent to outlet of type %s", o.typ)
	}
}

// Any will send any message.
func (o *Outlet) Any(msg string, atoms []Atom) {
	if o.typ == Any {
		o.obj.Push(Event{Outlet: o, Type: Any, Msg: msg, Data: atoms})
	} else {
		Error("any sent to outlet of type %s", o.typ)
	}
}

/* Threads */

var queue = map[uint64]func(){}
var queueMutex sync.Mutex

// IsMainThread will return if the Max main thead is executing.
func IsMainThread() bool {
	return C.systhread_ismainthread() == 1
}

//export maxgoYield
func maxgoYield(ref uint64) {
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
	// check empty
	if argc == 0 {
		return nil
	}

	// cast to slice
	var list []C.t_atom
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&list))
	sliceHeader.Cap = int(argc)
	sliceHeader.Len = int(argc)
	sliceHeader.Data = uintptr(unsafe.Pointer(argv))

	// allocate result
	atoms := make([]interface{}, len(list))

	// add atoms
	for i, item := range list {
		switch item.a_type {
		case C.A_LONG:
			atoms[i] = int64(C.atom_getlong(&item))
		case C.A_FLOAT:
			atoms[i] = float64(C.atom_getfloat(&item))
		case C.A_SYM:
			atoms[i] = C.GoString(C.atom_getsym(&item).s_name)
		default:
			atoms[i] = nil
		}
	}

	return atoms
}

// the receiver must arrange for the returned non-nil array to be freed
func encodeAtoms(atoms []Atom) (int64, *C.t_atom) {
	// check length
	if len(atoms) == 0 {
		return 0, nil
	}

	// allocate atom array
	array := (*C.t_atom)(unsafe.Pointer(C.getbytes(C.t_getbytes_size(len(atoms) * C.sizeof_t_atom))))

	// cast to slice
	var slice []C.t_atom
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	sliceHeader.Cap = len(atoms)
	sliceHeader.Len = len(atoms)
	sliceHeader.Data = uintptr(unsafe.Pointer(array))

	// set atoms
	for i, atom := range atoms {
		switch atom := atom.(type) {
		case int64:
			C.atom_setlong(&slice[i], C.t_atom_long(atom))
		case float64:
			C.atom_setfloat(&slice[i], C.double(atom))
		case string:
			C.atom_setsym(&slice[i], gensym(atom))
		}
	}

	return int64(len(atoms)), array
}
