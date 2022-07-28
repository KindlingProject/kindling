#ifndef KINDLING_PROBE_PROFILE_FLAMEGRAPH_H
#define KINDLING_PROBE_PROFILE_FLAMEGRAPH_H
extern "C" {
#include "profile/perf/perf.h"
}
#include "utils/ring_buffer.h"
#include <string>
#include <list>
#include <map>
#include <vector>

using std::list;
using std::map;
using std::vector;
using std::string;

static int getPerfColor(bool user) {
  if (user) {
    return 1;
  } else {
    return 5;
  }
}

class FlameSymbolData {
  public:
    FlameSymbolData() {
    }

    string symbol_;
    int color_;
};

class FlameSymbolDatas {
  public:
    vector<FlameSymbolData*> symbol_datas_;
    int symbol_size_;
    int max_depth_;

    FlameSymbolDatas(int max_depth) : symbol_size_(0), max_depth_(max_depth) {
      symbol_datas_ = vector<FlameSymbolData*>();
      for (int i = 0; i < max_depth; i++) {
        symbol_datas_.push_back(new FlameSymbolData());
      }
    }

    bool addPerfSymbol(int depth, __u64 ip, int pid, bool user);
};

class SampleData {
  public:
    SampleData() {
    }

    __u32 pid_;
    __u32 tid_;
    __u64 nr_;
    __u64 ips_[256];
  
    void collectStacks(FlameSymbolDatas *symbolDatas);
};

class Node {
  public:
    map<string, Node> children_;
    __u64 total_;
    __u64 self_;
    __u32 color_;

    Node() : children_(), total_(0), self_(0), color_(0) {
    }

    Node* addChild(const string& key, int color) {
        total_ += 1;
        Node* node = &children_[key];
        if (node->total_ == 0) {
          node->color_ = color;
        }
        return node;
    }

    void addLeaf() {
        total_ += 1;
        self_ += 1;
    }
};

class AggregateData {
  public:
    AggregateData(__u32 tid, int max_depth) : tid_(tid), root_(), func_name_map_(), func_names_() {
      root_.color_ = getPerfColor(false);
      symbol_datas_ = new FlameSymbolDatas(max_depth);
    }
    ~AggregateData() {
      func_name_map_.clear();
      func_names_.clear();
      delete symbol_datas_;
    }
    Node* root() {
        return &root_;
    }
    void Reset() {
      if (root_.total_ > 0) {
        root_ = Node();
        root_.color_ = getPerfColor(false);
      }
    }
    void Aggregate(void* data);
    void DumpFrameDatas(string& frameDatas, bool file);
    void DumpFuncNames(string& frameDatas);

  private:
    __u32 tid_;
    Node root_;
    map<string, int> func_name_map_;
    vector<string> func_names_;
    FlameSymbolDatas *symbol_datas_;

    int getFuncId(const string& funcName);
    void dumpFrameData(string& frameDatas, const string& name, const Node& node, int depth, __u64 x, bool seperator, bool file);
};

class FlameGraph {
 public:
  FlameGraph(int cache_keep_time, int perf_period);
  ~FlameGraph();
  void EnableAutoGet();
  void EnableFlameFile();
  void SetMaxDepth(int max_depth);
  void SetFilterThreshold(int filter_threshold);
  void RecordSampleData(struct sample_type_data *sample_data);
  void CollectData();
  string GetOnCpuData(__u32 tid, vector<std::pair<uint64_t, uint64_t>> &periods);

 private:
  void resetLogFile();

  __u64 cache_keep_time_;
  __u64 perf_period_ns_;
  __u64 perf_threshold_ns_;
  bool write_flame_graph_;
  __u64 last_sample_time_;
  __u64 last_collect_time_;
  BucketRingBuffers<SampleData> *sample_datas_;
  FILE *collect_file_;
};
#endif //KINDLING_PROBE_PROFILE_FLAMEGRAPH_H