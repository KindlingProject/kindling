#ifndef KINDLING_PROBE_PROFILE_FLAMEGRAPH_H
#define KINDLING_PROBE_PROFILE_FLAMEGRAPH_H
extern "C" {
#include "profile/perf/perf.h"
}
#include "utils/window_list.h"
#include <string>
#include <list>
#include <map>
#include <vector>

using std::list;
using std::map;
using std::vector;
using std::string;

enum FrameTypeId {
  FRAME_INTERPRETED  = 0,
  FRAME_JIT_COMPILED = 1,
  FRAME_INLINED      = 2,
  FRAME_NATIVE       = 3,
  FRAME_CPP          = 4,
  FRAME_KERNEL       = 5,
  FRAME_C1_COMPILED  = 6,
};

enum SymbolType {
  TYPE_KERNEL = 0,
  TYPE_USER = 1,
  TYPE_JVM = 2,
};

static bool endsWith(const string& s, const char* suffix, size_t suffixlen) {
    size_t len = s.length();
    return len >= suffixlen && s.compare(len - suffixlen, suffixlen, suffix) == 0; 
}

class FlameSymbolData {
  public:
    FlameSymbolData() {
    }

    ~FlameSymbolData() {
    }

    string symbol_;
    int type_;
};

class FlameSymbolDatas {
  public:
    vector<FlameSymbolData*> symbol_datas_;
    int symbol_from_;
    int symbol_to_;
    int max_depth_;

    FlameSymbolDatas(int max_depth) : symbol_from_(0), symbol_to_(0), max_depth_(max_depth) {
      symbol_datas_ = vector<FlameSymbolData*>();
    }

    ~FlameSymbolDatas() {
      for (auto itr = symbol_datas_.begin(); itr != symbol_datas_.end(); itr++) {
        if (NULL != *itr) {
          delete *itr;
          *itr = NULL;
        }
      }
      symbol_datas_.clear();
    }

    FlameSymbolData* getOrCreateSymbolData(int depth) {
      if (symbol_datas_.size() == 0) {
        for (int i = 0; i < max_depth_; i++) {
          symbol_datas_.push_back(new FlameSymbolData());
        }
      }
      return symbol_datas_.at(depth);
    }

    bool addPerfSymbol(int depth, __u64 ip, int pid, bool user);
    void addProfileSymbol(int depth, string symbol);
};

class SampleData {
  public:
    SampleData() {
    }
    ~SampleData() {
    }

    long ts_;
    __u32 pid_;
    __u64 nr_;
    __u64 ips_[256];
  
    void collectStacks(FlameSymbolDatas *symbolDatas);
};

class ProfileData {
  public:
    ProfileData() {
    }

    ProfileData(long ts, int depth, bool finish, string stack) : ts_(ts), depth_(depth), finish_(finish), stack_(stack) {
    }
    uint64_t ts_;
    __u32 tid_;
    int depth_;
    bool finish_;
    string stack_;

    bool collectStacks(FlameSymbolDatas *symbolDatas);
};

class Node {
  public:
    map<string, Node*> children_;
    __u64 total_;
    __u64 self_;
    int type_;

    Node() : children_(), total_(0), self_(0), type_(0) {
    }

    ~Node() {
      for (auto itr = children_.begin(); itr != children_.end();) {
        auto node = itr->second;
        if (node) {
          delete node;
          node = NULL;
          children_.erase(itr++);
        }
      }
      children_.clear();
    }

    Node* addChild(const string& key, int type) {
        total_ += 1;
        Node* node = children_[key];
        if (node == NULL) {
          node = children_[key] = new Node();
          node->type_ = type;
        }
        return node;
    }

    int getColor(string& name) {
      if (type_ == TYPE_KERNEL) {
        return FRAME_KERNEL;
      } else if (type_ == TYPE_USER) {
        return FRAME_JIT_COMPILED;
      } else if (type_ == TYPE_JVM) {
        if (endsWith(name, "_[j]", 4)) {
          name = name.substr(0, name.length() - 4);
          return FRAME_JIT_COMPILED;
        } else if (endsWith(name, "_[i]", 4)) {
          name = name.substr(0, name.length() - 4);
          return FRAME_INLINED;
        } else if (endsWith(name, "_[k]", 4)) {
          name = name.substr(0, name.length() - 4);
          return FRAME_KERNEL;
        } else if (name.find("::") != std::string::npos || name.compare(0, 2, "-[") == 0 || name.compare(0, 2, "+[") == 0) {
          return FRAME_CPP;
        } else if (((int)name.find('/') > 0 && name[0] != '[')
                || ((int)name.find('.') > 0 && name[0] >= 'A' && name[0] <= 'Z')) {
          return FRAME_JIT_COMPILED;
        } else {
          return FRAME_NATIVE;
        }
      } else {
        return FRAME_JIT_COMPILED;
      }
    }

    void addLeaf() {
        total_ += 1;
        self_ += 1;
    }
};

class AggregateData {
  public:
    FlameSymbolDatas *symbol_datas_;

    AggregateData(int max_depth) : func_name_map_(), func_names_() {
      symbol_datas_ = new FlameSymbolDatas(max_depth);
      root_ = new Node();
    }

    ~AggregateData() {
      func_name_map_.clear();
      func_names_.clear();
      delete root_;
      delete symbol_datas_;
    }
    Node* root() {
        return root_;
    }
    void Reset() {
      if (root_->total_ > 0) {
        delete root_;
        root_ = new Node();
      }
    }
    void Aggregate();
    void DumpFrameDatas(string& frameDatas, bool file);
    void DumpFuncNames(string& frameDatas);

  private:
    Node *root_;
    map<string, int> func_name_map_;
    vector<string> func_names_;

    int getFuncId(const string& funcName);
    void dumpFrameData(string& frameDatas, const string& name, Node* node, int depth, __u64 x, bool seperator, bool file);
};

class FlameGraph {
 public:
  FlameGraph(int cache_second, int perf_period_ms);
  ~FlameGraph();
  void SetMaxDepth(int max_depth);
  void ExpireCache(int seconds);
  void RecordSampleData(struct sample_type_data *sample_data);
  void RecordProfileData(uint64_t time, __u32 tid, int depth, bool finish, string stack);
  string GetOnCpuData(__u32 tid, vector<std::pair<uint64_t, uint64_t>> &periods);

 private:
  int size_;
  __u64 perf_threshold_ns_;
  WindowList<SampleData> *sample_datas_;
  WindowList<ProfileData> *profile_datas_;
  FILE *collect_file_;
};
#endif //KINDLING_PROBE_PROFILE_FLAMEGRAPH_H