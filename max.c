#include "_cgo_export.h"

#include "max.h"

/* Basic */

void maxgo_log(char *str) {
  post(str);
  free(str);
}

void maxgo_error(char *str) {
  error(str);
  free(str);
}

void maxgo_alert(char *str) {
  ouchstring(str);
  free(str);
}

t_symbol *maxgo_gensym(char *str) {
  t_symbol *sym = gensym(str);
  free(str);
  return sym;
}

/* Initialization */

extern void ext_main(void *r) { maxgoMain(); }

/* Classes */

static t_class *class = NULL;

typedef struct {
  t_object obj;
  long inlet;
  void **proxies;
  int num_proxies;
  unsigned long long ref;
  void *clock;
} t_bridge;

static void bridge_tick(void *ptr) {
  // get bridge
  t_bridge *bridge = (t_bridge *)ptr;

  // handle all queued events
  for (;;) {
    // retrieve next event
    struct maxgoPop_return ret =
        maxgoPop(bridge->ref);  // (unsafe.Pointer, C.maxgo_type_e, *C.t_symbol, int64, *C.t_atom, bool)

    // check result
    if (ret.r0 == NULL) {
      return;
    }

    // call outlet
    switch (ret.r1) {
      case MAXGO_BANG:
        outlet_bang(ret.r0);
        break;
      case MAXGO_INT:
        outlet_int(ret.r0, atom_getlong(ret.r4));
        break;
      case MAXGO_FLOAT:
        outlet_float(ret.r0, atom_getfloat(ret.r4));
        break;
      case MAXGO_LIST:
        outlet_list(ret.r0, NULL, ret.r3, ret.r4);
        break;
      case MAXGO_ANY:
        outlet_anything(ret.r0, ret.r2, ret.r3, ret.r4);
        break;
    }

    // free atoms
    if (ret.r4 != NULL) {
      freebytes(ret.r4, ret.r3 * sizeof(t_atom));
    }

    // return if there are no more events
    if (!ret.r5) {
      return;
    }
  }
}

static void *bridge_new(t_symbol *name, long argc, t_atom *argv) {
  // allocate bridge
  t_bridge *bridge = object_alloc(class);

  // initialize object
  struct maxgoInit_return ret = maxgoInit(&bridge->obj, argc, argv);  // (uint64, int)

  // set reference
  bridge->ref = ret.r0;

  // check reference
  if (bridge->ref == 0) {
    return NULL;
  }

  // save number of proxies
  bridge->num_proxies = ret.r1;

  // allocate proxy list
  bridge->proxies = (void **)getbytes(bridge->num_proxies * sizeof(void *));

  // create proxies
  for (int i = 0; i < bridge->num_proxies; i++) {
    bridge->proxies[i] = proxy_new(&bridge->obj, bridge->num_proxies - i, &bridge->inlet);
  }

  // create clock
  bridge->clock = clock_new(bridge, (method)bridge_tick);

  return bridge;
}

static void bridge_bang(t_bridge *bridge) {
  // get inlet
  long inlet = proxy_getinlet(&bridge->obj);

  // handle message
  maxgoHandle(bridge->ref, "bang", inlet, 0, NULL);
}

static void bridge_int(t_bridge *bridge, long n) {
  // get inlet
  long inlet = proxy_getinlet(&bridge->obj);

  // prepare args
  t_atom args[1] = {0};
  atom_setlong(args, n);

  // handle message
  maxgoHandle(bridge->ref, "int", inlet, 1, args);
}

static void bridge_float(t_bridge *bridge, double n) {
  // get inlet
  long inlet = proxy_getinlet(&bridge->obj);

  // prepare args
  t_atom args[1] = {0};
  atom_setfloat(args, n);

  // handle message
  maxgoHandle(bridge->ref, "float", inlet, 1, args);
}

static void bridge_gimme(t_bridge *bridge, t_symbol *msg, long argc, t_atom *argv) {
  // get inlet
  long inlet = proxy_getinlet(&bridge->obj);

  // handle message
  maxgoHandle(bridge->ref, msg->s_name, inlet, argc, argv);
}

static void bridge_loadbang(t_bridge *bridge) {
  // handle message
  maxgoHandle(bridge->ref, "loadbang", 0, 0, NULL);
}

static void bridge_dblclick(t_bridge *bridge) {
  // handle message
  maxgoHandle(bridge->ref, "dblclick", 0, 0, NULL);
}

static void bridge_assist(t_bridge *bridge, void *b, long io, long i, char *buf) {
  // get info
  struct maxgoDescribe_return ret = maxgoDescribe(bridge->ref, io, i);  // (*C.char, bool)

  // copy label
  strncpy_zero(buf, ret.r0, 512);

  // free string
  free(ret.r0);
}

static void bridge_inletinfo(t_bridge *bridge, void *b, long i, char *v) {
  // get info
  struct maxgoDescribe_return ret = maxgoDescribe(bridge->ref, 1, i);  // (*C.char, bool)

  // set cold if not hot
  if (!ret.r1) {
    *v = 1;
  }

  // free string
  free(ret.r0);
}

static void bridge_free(t_bridge *bridge) {
  // free object
  maxgoFree(bridge->ref);

  // free proxies
  for (int i = 0; i < bridge->num_proxies; i++) {
    object_free(bridge->proxies[i]);
  }

  // free list
  freebytes(bridge->proxies, bridge->num_proxies * sizeof(void *));

  // free clock
  clock_unset(bridge->clock);
  freeobject((t_object *)bridge->clock);
}

void maxgo_init(char *name) {
  // check class
  if (class != NULL) {
    error("class has already been initialized");
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
  class_addmethod(class, (method)bridge_loadbang, "loadbang", 0);
  class_addmethod(class, (method)bridge_dblclick, "dblclick", 0);
  class_addmethod(class, (method)bridge_assist, "assist", A_CANT, 0);
  class_addmethod(class, (method)bridge_inletinfo, "inletinfo", A_CANT, 0);

  // register class
  class_register(CLASS_BOX, class);

  // free name
  free(name);
}

void maxgo_notify(void *ptr) {
  // get bridge
  t_bridge *bridge = (t_bridge *)ptr;

  // schedule clock
  clock_delay(bridge->clock, 0);
}

/* Threads */

void maxgo_yield(void *p, void *ref) {
  // yield back
  maxgoYield((unsigned long long)ref);
}

void maxgo_defer(unsigned long long ref) {
  // defer function call
  defer_low(NULL, (method)maxgo_yield, (void *)ref, 0, NULL);
}
