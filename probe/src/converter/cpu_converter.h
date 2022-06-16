#ifndef CPU_CONVERTER_H
#define CPU_CONVERTER_H
//#include "profile/profiler.h"
#include "log/log_info.h"
#include "event_cache.h"
#include <string>
#include <map>
#include "sinsp.h"
#include "cgo/kindling.h"
class cpu_data {
public:
    uint64_t start_time;
    uint64_t end_time;
    uint64_t on_total_time;
    uint64_t off_total_time;
    string time_specs;
    string runq_latency;
    string time_type;
    uint32_t tid;
};

class cpu_converter
{
public:
    cpu_converter(sinsp *inspector);
//    cpu_converter(sinsp *inspector, Profiler *prof, LogCache *log);
    ~cpu_converter();
    int convert(kindling_event_t_for_go *p_kindling_event, sinsp_evt *evt);
    bool Cache(sinsp_evt *evt);
private:
    int init_kindling_event(kindling_event_t_for_go *p_kindling_event, sinsp_evt *sevt);
    int add_threadinfo(kindling_event_t_for_go *p_kindling_event, sinsp_evt *sevt);
    int add_cpu_data(kindling_event_t_for_go *p_kindling_event, sinsp_evt *sevt);

    int32_t set_boot_time(uint64_t *boot_time);

    sinsp *m_inspector;
//    Profiler *m_profiler;
    LogCache *m_log;
    uint64_t sample_interval;
    event_cache *file_cache;
    event_cache *net_cache;
    epoll_event_cache *epoll_cache;
};

#endif //CPU_CONVERTER_H
