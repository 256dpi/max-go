#ifndef API
#define API 1

#ifdef MAC_VERSION
#include <Carbon/Carbon.h>
#endif

#include <ext.h>

void maxgo_log(const char *str);
void maxgo_error(const char *str);
void maxgo_alert(const char *str);

t_class *maxgo_class_new(const char *name);

void maxgo_class_add_method(t_class *class, const char *name);

#endif
