#include "_cgo_export.h"

#include "max.h"

/* Basic */

void maxgo_log(const char *str) { post(str); }

void maxgo_error(const char *str) { error(str); }

void maxgo_alert(const char *str) { ouchstring(str); }

/* Classes */

typedef struct {
  t_object obj;
  t_symbol *name;
  t_class *class;
  long inlet;
  void *proxy;
  GoUintptr ref;
} t_bridge;

static void *bridge_new(t_symbol *name, long argc, t_atom *argv) {
  // get class
  t_class *class = gomaxGet(name->s_name);

  // allocate bridge
  t_bridge *bridge = object_alloc(class);

  // set info
  bridge->name = name;
  bridge->class = class;

  // initialize object
  bridge->ref = gomaxInit(name->s_name, &bridge->obj);

  // TODO: A proxy also allocates an inlet...

  // create proxy
  bridge->proxy = proxy_new(&bridge->obj, 1, &bridge->inlet);

  return bridge;
}

static void bridge_bang(t_bridge *bridge) {
  // get inlet
  long inlet = proxy_getinlet(&bridge->obj);

  // dispatch message
  gomaxMessage(bridge->name->s_name, bridge->ref, "$bang", inlet, 0, NULL);
}

static void bridge_int(t_bridge *bridge, long n) {
  // get inlet
  long inlet = proxy_getinlet(&bridge->obj);

  // prepare args
  t_atom args[1] = {0};
  atom_setlong(args, n);

  // dispatch message
  gomaxMessage(bridge->name->s_name, bridge->ref, "$int", inlet, 1, args);
}

static void bridge_float(t_bridge *bridge, double n) {
  // get inlet
  long inlet = proxy_getinlet(&bridge->obj);

  // prepare args
  t_atom args[1] = {0};
  atom_setfloat(args, n);

  // dispatch message
  gomaxMessage(bridge->name->s_name, bridge->ref, "$float", inlet, 1, args);
}

static void bridge_gimme(t_bridge *bridge, t_symbol *msg, long argc,
                         t_atom *argv) {
  // get inlet
  long inlet = proxy_getinlet(&bridge->obj);

  // dispatch message
  gomaxMessage(bridge->name->s_name, bridge->ref, msg->s_name, inlet, argc,
               argv);
}

static void bridge_dblclick(t_bridge *bridge) {
  gomaxMessage(bridge->name->s_name, bridge->ref, "$dblclick", 0, 0, NULL);
}

static void bridge_assist(t_bridge *bridge, void *b, long io, long i,
                          char *buf) {
  const char *str = gomaxAssist(bridge->name->s_name, bridge->ref, io, i);
  strncpy_zero(buf, str, 512);
}

static void bridge_free(t_bridge *bridge) {
  // free object
  gomaxFree(bridge->name->s_name, bridge->ref);

  // free proxy
  object_free(bridge->proxy);
}

t_class *maxgo_class_new(const char *name) {
  // create class
  t_class *class = class_new(name, (method)bridge_new, (method)bridge_free,
                             (long)sizeof(t_bridge), 0L, A_GIMME, 0);

  // add base methods for leftmost inlet
  class_addmethod(class, (method)bridge_bang, "bang", 0);
  class_addmethod(class, (method)bridge_int, "int", A_LONG, 0);
  class_addmethod(class, (method)bridge_float, "float", A_FLOAT, 0);
  class_addmethod(class, (method)bridge_gimme, "anything", A_GIMME, 0);
  class_addmethod(class, (method)bridge_dblclick, "dblclick", 0);
  class_addmethod(class, (method)bridge_assist, "assist", A_CANT, 0);

  // TODO: How to handle multi value messages that do not start with a symbol?
  //  -> Now they are dispatched after each other.

  return class;
}

void maxgo_class_add_method(t_class *class, const char *name) {
  class_addmethod(class, (method)bridge_gimme, name, A_GIMME, 0);
}

/*
void maxgo_main() {
  class_addmethod(c, (method)dummy_assist, "assist", A_CANT, 0);
  CLASS_ATTR_SYM(c, "name", 0, t_dummy, name);
}

void dummy_assist(t_dummy *x, void *b, long m, long a, char *s) {
  if (m == ASSIST_INLET) { // inlet
    sprintf(s, "I am inlet %ld", a);
  } else { // outlet
    sprintf(s, "I am outlet %ld", a);
  }
}
*/
