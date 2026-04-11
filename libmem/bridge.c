#include "bridge.h"

/* Go callback declarations (defined in callbacks.go via //export) */
extern lm_bool_t goProcessCallback(lm_process_t *process, lm_void_t *arg);
extern lm_bool_t goThreadCallback(lm_thread_t *thread, lm_void_t *arg);
extern lm_bool_t goModuleCallback(lm_module_t *module, lm_void_t *arg);
extern lm_bool_t goSymbolCallback(lm_symbol_t *symbol, lm_void_t *arg);
extern lm_bool_t goSegmentCallback(lm_segment_t *segment, lm_void_t *arg);

lm_bool_t bridge_enum_processes(lm_void_t *arg) {
	return LM_EnumProcesses(goProcessCallback, arg);
}

lm_bool_t bridge_enum_threads(lm_void_t *arg) {
	return LM_EnumThreads(goThreadCallback, arg);
}

lm_bool_t bridge_enum_threads_ex(const lm_process_t *process, lm_void_t *arg) {
	return LM_EnumThreadsEx(process, goThreadCallback, arg);
}

lm_bool_t bridge_enum_modules(lm_void_t *arg) {
	return LM_EnumModules(goModuleCallback, arg);
}

lm_bool_t bridge_enum_modules_ex(const lm_process_t *process, lm_void_t *arg) {
	return LM_EnumModulesEx(process, goModuleCallback, arg);
}

lm_bool_t bridge_enum_symbols(const lm_module_t *module, lm_void_t *arg) {
	return LM_EnumSymbols(module, goSymbolCallback, arg);
}

lm_bool_t bridge_enum_symbols_demangled(const lm_module_t *module, lm_void_t *arg) {
	return LM_EnumSymbolsDemangled(module, goSymbolCallback, arg);
}

lm_bool_t bridge_enum_segments(lm_void_t *arg) {
	return LM_EnumSegments(goSegmentCallback, arg);
}

lm_bool_t bridge_enum_segments_ex(const lm_process_t *process, lm_void_t *arg) {
	return LM_EnumSegmentsEx(process, goSegmentCallback, arg);
}
