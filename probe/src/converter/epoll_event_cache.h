//
// Created by 散养鸡 on 2022/6/22.
//

#ifndef KINDLING_PROBE_EPOLL_CACHE_H
#define KINDLING_PROBE_EPOLL_CACHE_H
#include "event_cache.h"
class epoll_info : public info_base {
public:
    epoll_info() {}
    ~epoll_info() {}
    vector<int> fds;
    string toString() {
        return "net@" + operation_type + "@" + name + "@" + to_string(size) + "@" + relate_id;
    }
};

class epoll_event_cache : public event_cache {
public:
    epoll_event_cache(uint8_t type) : event_cache(type) {}
    bool setInfo(sinsp_evt *evt);
    bool SetLastEpollCache(uint32_t tid, int64_t fd, info_base *info);
};
#endif //KINDLING_PROBE_EPOLL_CACHE_H
