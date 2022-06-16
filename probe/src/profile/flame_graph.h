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

class SampleData {
  public:
    SampleData();
    string GetString();

    __u32 pid_;
    __u32 tid_;
    __u64 nr_;
    __u64 ips_[256];
  private:
    string addThreadInfo(string call_stack);
};

class AggregateData {
  public:
    AggregateData(__u32 tid);
    ~AggregateData();
    void Aggregate(void* data);
    void Reset();
    string ToString();

  private:
    __u32 tid_;
    map<string, int> agg_count_map_;
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