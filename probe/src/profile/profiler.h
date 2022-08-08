#ifndef KINDLING_PROBE_PROFILE_PROFILER_H
#define KINDLING_PROBE_PROFILE_PROFILER_H
extern "C" {
#include "profile/perf/perf.h"
}
#include <string>
#include <vector>

class Profiler {
public:
    Profiler(int size, int cache_keep_ms, int perf_period_ms);
    ~Profiler();
    void SetMaxDepth(int max_depth);
    void Start();
    void Stop();
    void RecordProfileData(uint64_t time, __u32 pid, __u32 tid, int depth, bool finish, std::string stack);
    std::string GetOnCpuData(__u32 pid, __u32 tid, std::vector<std::pair<uint64_t, uint64_t>> &periods);

private:
    struct perfData *perf_data_;
};

#endif //KINDLING_PROBE_PROFILE_PROFILER_