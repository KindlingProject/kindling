#include "symbol.h"
#include <linux/elf.h>

BPFSymbolTable::BPFSymbolTable() {
  uint32_t use_symbol_type = (1 << STT_FUNC) | (1 << STT_GNU_IFUNC);
  symbol_option_ = {.use_debug_file = true,
                    .check_debug_file_crc = true,
                    .lazy_symbolize = 1,
                    .use_symbol_type = use_symbol_type};
}

BPFSymbolTable::~BPFSymbolTable() {
  for (auto it : pid_sym_)
    bcc_free_symcache(it.second, it.first);
}

void BPFSymbolTable::FreeSymcache(int pid) {
  auto iter = pid_sym_.find(pid);
  if (iter != pid_sym_.end()) {
    bcc_free_symcache(iter->second, iter->first);
    pid_sym_.erase(iter);
  }
}

std::string BPFSymbolTable::GetSymbol(uint64_t addr, int pid) {
    if (pid < 0) {
        pid = -1;
    }

    if (pid_sym_.find(pid) == pid_sym_.end()) {
        pid_sym_[pid] = bcc_symcache_new(pid, &symbol_option_);
    }
    void* cache = pid_sym_[pid];
    bcc_symbol symbol;
    std::string res;
    if (bcc_symcache_resolve(cache, addr, &symbol) != 0) {
        res = std::string("[UNKNOWN]");
    } else {
        res = std::string(symbol.demangle_name);
        bcc_symbol_free_demangle_name(&symbol);
    }
    return res;
}