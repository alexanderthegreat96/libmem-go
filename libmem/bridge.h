#ifndef LIBMEM_GO_BRIDGE_H
#define LIBMEM_GO_BRIDGE_H

#include <libmem/libmem.h>

/* Process enumeration bridge */
lm_bool_t bridge_enum_processes(lm_void_t *arg);

/* Thread enumeration bridges */
lm_bool_t bridge_enum_threads(lm_void_t *arg);
lm_bool_t bridge_enum_threads_ex(const lm_process_t *process, lm_void_t *arg);

/* Module enumeration bridges */
lm_bool_t bridge_enum_modules(lm_void_t *arg);
lm_bool_t bridge_enum_modules_ex(const lm_process_t *process, lm_void_t *arg);

/* Symbol enumeration bridges */
lm_bool_t bridge_enum_symbols(const lm_module_t *module, lm_void_t *arg);
lm_bool_t bridge_enum_symbols_demangled(const lm_module_t *module, lm_void_t *arg);

/* Segment enumeration bridges */
lm_bool_t bridge_enum_segments(lm_void_t *arg);
lm_bool_t bridge_enum_segments_ex(const lm_process_t *process, lm_void_t *arg);

#endif
