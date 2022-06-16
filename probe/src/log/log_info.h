#ifndef KINDLING_PROBE_LOG_INFO_H
#define KINDLING_PROBE_LOG_INFO_H
#include "ring_buffer.h"
#include <string>
#include <list>
#include <vector>
#include <asm/types.h>

using std::string;
using std::list;
using std::vector;

class LogData {
  public:
    LogData();
    ~LogData();
    void setData(long ts, int64_t size, __u32 tid, char* data);
    long getTs();
    __u32 getTid();
    string getData();
  private:
    long ts_;
    int64_t size_;
    __u32 tid_;
    string data_;
};

class LogDatas {
  public:
    LogDatas(__u32 tid);
    ~LogDatas();
    void CollectLogs(void* data);
    void Reset();
    string ToString();
  private:
    __u32 tid_;
    list<string> logs_;
};

class LogCache {
  public:
    LogCache(int size, int cacheMs);
    ~LogCache();
    bool addLog(void* evt);
    string getLogs(__u32 tid, vector<std::pair<uint64_t, uint64_t>> &periods);
  private:
    BucketRingBuffers<LogData>* logs_;
};

#endif //KINDLING_PROBE_LOG_INFO_H