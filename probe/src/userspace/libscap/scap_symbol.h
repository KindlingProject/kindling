#ifndef SYSDIG_SCAP_SYMBOL_H
#define SYSDIG_SCAP_SYMBOL_H

#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <gelf.h>
#include <stdbool.h>

struct symbol {
	const char *name;
	const char *module;
	uint64_t offset;
};

struct symbol_option {
	int use_debug_file;
	int check_debug_file_crc;
	// Bitmask flags indicating what types of ELF symbols to use
	uint32_t use_symbol_type;
};

struct load_addr_t {
	uint64_t target_addr;
	uint64_t binary_addr;
};

// Symbol name, start address, length, payload
// Callback returning a negative value indicates to stop the iteration
typedef int (*elf_symcb)(const char *, uint64_t, void *);

// Segment virtual address, memory size, file offset, payload
// Callback returning a negative value indicates to stop the iteration
typedef int (*elf_load_sectioncb)(uint64_t, uint64_t, uint64_t, void *);

int resolve_symbol_name(const char *module, const char *symbol_name, uint64_t *res_addr);

// Iterate over symbol table of a binary module
// Parameter "option" points to a symbol_option struct to indicate whether
// and how to use debuginfo file, and what types of symbols to load.
// Returns -1 on error, and 0 on success or stopped by callback
int elf_foreach_sym(const char *path, elf_symcb callback, void *option, void *payload);

int elf_get_type(const char *path);

// Iterate over all executable load sections of an ELF
// Returns -1 on error, 1 if stopped by callback, and 0 on success
int elf_foreach_load_section(const char *path,
							 elf_load_sectioncb callback,
							 void *payload);
#endif //SYSDIG_SCAP_SYMBOL_H

