#include <stdio.h>
#include <stdlib.h>
#include <dlfcn.h>

void atom_getfloat(void) { printf("%s\n", __func__); }
void atom_getlong(void) { printf("%s\n", __func__); }
void atom_getsym(void) { printf("%s\n", __func__); }
void atom_setfloat(void) { printf("%s\n", __func__); }
void atom_setlong(void) { printf("%s\n", __func__); }
void atom_setsym(void) { printf("%s\n", __func__); }
void bangout(void) { printf("%s\n", __func__); }
void class_addmethod(void) { printf("%s\n", __func__); }
void class_dspinit(void) { printf("%s\n", __func__); }
void class_new(void) { printf("%s\n", __func__); }
void class_register(void) { printf("%s\n", __func__); }
void clock_delay(void) { printf("%s\n", __func__); }
void clock_new(void) { printf("%s\n", __func__); }
void defer_low(void) { printf("%s\n", __func__); }
void clock_unset(void) { printf("%s\n", __func__); }
void error(void) { printf("%s\n", __func__); }
void floatout(void) { printf("%s\n", __func__); }
void freeobject(void) { printf("%s\n", __func__); }
void gensym(void) { printf("%s\n", __func__); }
void intout(void) { printf("%s\n", __func__); }
void listout(void) { printf("%s\n", __func__); }
void object_alloc(void) { printf("%s\n", __func__); }
void object_free(void) { printf("%s\n", __func__); }
void object_method_imp(void) { printf("%s\n", __func__); }
void ouchstring(void) { printf("%s\n", __func__); }
void outlet_anything(void) { printf("%s\n", __func__); }
void outlet_bang(void) { printf("%s\n", __func__); }
void outlet_float(void) { printf("%s\n", __func__); }
void outlet_int(void) { printf("%s\n", __func__); }
void outlet_list(void) { printf("%s\n", __func__); }
void outlet_new(void) { printf("%s\n", __func__); }
void post(void) { printf("%s\n", __func__); }
void proxy_getinlet(void) { printf("%s\n", __func__); }
void proxy_new(void) { printf("%s\n", __func__); }
void strncpy_zero(void) { printf("%s\n", __func__); }
void sysmem_freeptr(void) { printf("%s\n", __func__); }
void sysmem_newptr(void) { printf("%s\n", __func__); }
void systhread_ismainthread(void) { printf("%s\n", __func__); }
void z_dsp_free(void) { printf("%s\n", __func__); }
void z_dsp_setup(void) { printf("%s\n", __func__); }

int main() {
    // open library
	void *handle = dlopen("../example/out/maxgo.mxo/MacOS/maxgo", RTLD_NOW);
	if (handle == NULL) {
		fprintf(stderr, "%s\n", dlerror());
		exit(1);
	}

    // get ext_main
	void (*ext_main)() = dlsym(handle, "ext_main");
    if (ext_main == NULL)  {
        fprintf(stderr, "%s\n", dlerror());
        exit(1);
    }

    // call ext_main
    printf("--- START ---\n");
    ext_main();
    printf("--- END ---\n");

	return 0;
}
