#include "_cgo_export.h"

#include "max.h"

/* Basic */

void maxgo_log(const char *str) { post(str); }

void maxgo_error(const char *str) { error(str); }

void maxgo_alert(const char *str) { ouchstring(str); }

/* Classes */

static t_class *class = NULL;

typedef struct {
  t_object obj;
  long inlet;
  void *proxy;
  unsigned long long ref;
} t_bridge;

static void *bridge_new(t_symbol *name, long argc, t_atom *argv) {
  // allocate bridge
  t_bridge *bridge = object_alloc(class);

  // initialize object
  bridge->ref = gomaxInit(&bridge->obj, argc, argv);

  // TODO: A proxy also allocates an inlet...

  // create proxy
  bridge->proxy = proxy_new(&bridge->obj, 1, &bridge->inlet);

  return bridge;
}

static void bridge_bang(t_bridge *bridge) {
  // get inlet
  long inlet = proxy_getinlet(&bridge->obj);

  // dispatch message
  gomaxMessage(bridge->ref, "bang", inlet, 0, NULL);
}

static void bridge_int(t_bridge *bridge, long n) {
  // get inlet
  long inlet = proxy_getinlet(&bridge->obj);

  // prepare args
  t_atom args[1] = {0};
  atom_setlong(args, n);

  // dispatch message
  gomaxMessage(bridge->ref, "int", inlet, 1, args);
}

static void bridge_float(t_bridge *bridge, double n) {
  // get inlet
  long inlet = proxy_getinlet(&bridge->obj);

  // prepare args
  t_atom args[1] = {0};
  atom_setfloat(args, n);

  // dispatch message
  gomaxMessage(bridge->ref, "float", inlet, 1, args);
}

static void bridge_gimme(t_bridge *bridge, t_symbol *msg, long argc, t_atom *argv) {
  // get inlet
  long inlet = proxy_getinlet(&bridge->obj);

  // dispatch message
  gomaxMessage(bridge->ref, msg->s_name, inlet, argc, argv);
}

static void bridge_dblclick(t_bridge *bridge) { gomaxMessage(bridge->ref, "dblclick", 0, 0, NULL); }

static void bridge_assist(t_bridge *bridge, void *b, long io, long i, char *buf) {
  const char *str = gomaxAssist(bridge->ref, io, i);
  strncpy_zero(buf, str, 512);
}

static void bridge_free(t_bridge *bridge) {
  // free object
  gomaxFree(bridge->ref);

  // free proxy
  object_free(bridge->proxy);
}

void maxgo_init(const char *name) {
  // check class
  if (class != NULL) {
    error("maxgo_init: has already been called");
    return;
  }

  // create class
  class = class_new(name, (method)bridge_new, (method)bridge_free, (long)sizeof(t_bridge), 0L, A_GIMME, 0);

  // add generic methods
  class_addmethod(class, (method)bridge_bang, "bang", 0);
  class_addmethod(class, (method)bridge_int, "int", A_LONG, 0);
  class_addmethod(class, (method)bridge_float, "float", A_FLOAT, 0);
  class_addmethod(class, (method)bridge_gimme, "list", A_GIMME, 0);
  class_addmethod(class, (method)bridge_gimme, "anything", A_GIMME, 0);
  class_addmethod(class, (method)bridge_dblclick, "dblclick", 0);
  class_addmethod(class, (method)bridge_assist, "assist", A_CANT, 0);

  // register class
  class_register(CLASS_BOX, class);
}
