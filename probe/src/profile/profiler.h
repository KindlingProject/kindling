#ifndef KINDLING_PROBE_PROFILE_PROFILER_H
#define KINDLING_PROBE_PROFILE_PROFILER_H
extern "C" {
#include "profile/perf/perf.h"
}
#include <string>
#include <vector>

class Profiler {
public:
    Profiler(int cache_keep_time, int perf_period_ms);
    ~Profiler();
    void Start();
    void Stop();
    void EnableAutoGet();
    void EnableFlameFile();
    void SetMaxDepth(int max_depth);
    void SetFilterThreshold(int filter_threshold);
    std::string GetOnCpuData(__u32 tid, std::vector<std::pair<uint64_t, uint64_t>> &periods);

private:
    struct perfData *perf_data_;
};

#endif //KINDLING_PROBE_PROFILE_PROFILER_