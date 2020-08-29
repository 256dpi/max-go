#ifndef API
#define API 1

#ifdef MAC_VERSION
#include <Carbon/Carbon.h>
#endif

#include <ext.h>

void maxgo_log(const char *str);
void maxgo_error(const char *str);
void maxgo_alert(const char *str);

void maxgo_init(const char *name);

#endif
