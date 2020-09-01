#ifndef API
#define API 1

#ifdef MAC_VERSION
#include <Carbon/Carbon.h>
#endif

#include <ext.h>

typedef enum {
  MAXGO_BANG = 0,
  MAXGO_INT,
  MAXGO_FLOAT,
  MAXGO_LIST,
  MAXGO_ANY,
} maxgo_type_e;

void maxgo_log(char *str);
void maxgo_error(char *str);
void maxgo_alert(char *str);
t_symbol *maxgo_gensym(char *name);
void maxgo_init(char *name);
void maxgo_notify(void *ptr);
void maxgo_defer(unsigned long long ref);

#endif
