#ifndef TUSK_OPENLIB_UNIX_H_
#define TUSK_OPENLIB_UNIX_H_

#ifndef _WIN32

#ifdef __cplusplus
extern "C"
{
#endif

#include <dlfcn.h>
#include "syscall.h"

    struct TUSK_LIB
    {
        void *module;
    };

    struct TUSK_CPROC
    {
        int (*proc)();
    };

    static inline void closelib(struct TUSK_LIB lib)
    {
        dlclose(lib.module);
    }

    static inline struct TUSK_LIB loadlib(char *name)
    {

        struct TUSK_LIB lib;
        lib.module = dlopen(name, RTLD_NOW);

        return lib;
    }

    static inline struct TUSK_CPROC loadproc(struct TUSK_LIB lib, char *proc)
    {
        struct TUSK_CPROC cproc;
        cproc.proc = dlsym(lib.module, proc);
        return cproc;
    }

    static inline long long int callproc(struct TUSK_CPROC proc, sysproto, void *a20) //add an extra argument
    {
        return proc.proc(a0,
                         a1,
                         a2,
                         a3,
                         a4,
                         a5,
                         a6,
                         a7,
                         a8,
                         a9,
                         a10,
                         a11,
                         a12,
                         a13,
                         a14,
                         a15,
                         a16,
                         a17,
                         a18,
                         a19,
                         a20);
    }

#ifdef __cplusplus
}
#endif

#endif

#endif