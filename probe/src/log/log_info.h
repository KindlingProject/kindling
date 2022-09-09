#ifndef KINDLING_PROBE_LOG_INFO_H
#define KINDLING_PROBE_LOG_INFO_H
#include "utils/window_list.h"
#include <string>
#include <list>
#include <vector>
#include <asm/types.h>

using std::string;
using std::list;
using std::vector;

class LogData {
  public:
    long ts_;
    string data_;

    LogData(long ts, string data) : ts_(ts), data_(data) {
    }

    ~LogData() {
      string().swap(data_);
    }
};

class LogDatas {
  public:
    LogDatas(int leftSize) : leftSize_(leftSize) {
    }
    ~LogDatas() {
      logs_.clear();
    }
    void Reset() {
      logs_.clear();
    }

    bool isOverLimit() {
      return leftSize_ < 0; 
    }

    void CollectLogs(void* data);
    string ToString();
  private:
    int leftSize_;
    list<string> logs_;
};

class LogCache {
  public:
    LogCache(int cache_second);
    ~LogCache();
    bool addLog(void* evt);
    string GetLogs(__u32 tid, vector<std::pair<uint64_t, uint64_t>> &periods, int maxLength);
    void ExpireCache(int seconds);
  private:
    WindowList<LogData>* logs_;
};

#endif //KINDLING_PROBE_LOG_INFO_H
