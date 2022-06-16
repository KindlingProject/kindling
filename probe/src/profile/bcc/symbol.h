#pragma once

#include <string>
#include <map>
#include <bcc_syms.h>

class BPFSymbolTable {
 public:
  BPFSymbolTable();
  ~BPFSymbolTable();

  void FreeSymcache(int pid);
  std::string GetSymbol(uint64_t addr, int pid);

 private:
  bcc_symbol_option symbol_option_;
  std::map<int, void*> pid_sym_;
};
