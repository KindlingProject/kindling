//
// Created by 散养鸡 on 2022/5/20.
//

#include <iostream>
#include "event_cache.h"
bool epoll_event_cache::SetLastEpollCache(uint32_t tid, int64_t fd, info_base *info) {
    // 获取当前线程上一个epoll，判断fd
    auto it = cache.find(tid);
    if (it == cache.end())
        return false;
    if (cache[tid]->size() == 0)
        return false;
    auto epoll_pair = cache[tid]->back();
    // 判断时间，1ms内认为可以关联
    if (info->start_time - epoll_pair->end_time < 1000000) { // 1ms
        for (auto efd : static_cast<epoll_info*>(epoll_pair)->fds) {
            if (fd == efd) {
                // 匹配fd，更新epoll 事件中的relate id信息为net.ts
                epoll_pair->relate_id = to_string(fd) + "-" + to_string(info->start_time);
                epoll_pair->name = info->name;
                epoll_pair->size = info->size;
                epoll_pair->operation_type = info->operation_type;
//                std::cout << "[epoll relate]" << epoll_pair->relate_id << " " << info->toStringTs() << std::endl;
//                std::cout << "current size: " << cache[tid]->size() << std::endl;
            }
        }
    }
    return true;
}
bool epoll_event_cache::setInfo(sinsp_evt *evt) {
    std::list<info_base*> *list;
    auto epoll_pair = new epoll_info();
    auto s_tinfo = evt->get_thread_info();
    if (s_tinfo->m_latency < 10000) // 10us
        return false;
    auto p = evt->get_param_value_raw("fds");
    if (!p)
        return false;
    auto payload = p->m_val;
    if (p->m_len < 2) {
        return false;
    }
    uint16_t nfds = *(uint16_t *)payload;
    if (nfds <= 0)
        return false;

    epoll_pair->fds = vector<int>(nfds);
    uint32_t pos = 2;
    for (int j = 0; j < nfds; j++)
    {
        if (pos + 10 > p->m_len) {
            break;
        }
        int64_t fd = *(int64_t *)(payload + pos);
        sinsp_fdinfo_t *fdinfo = s_tinfo->get_fd(fd);
	    if (fdinfo) { // only focus on net
            if (fdinfo->m_type == SCAP_FD_IPV4_SOCK || fdinfo->m_type == SCAP_FD_IPV4_SERVSOCK
                || fdinfo->m_type == SCAP_FD_IPV6_SOCK || fdinfo->m_type == SCAP_FD_IPV6_SERVSOCK) {
                epoll_pair->fds[j] = fd;
            }
	    }
        pos += 10;
    }
    epoll_pair->exit = true;
    epoll_pair->end_time = evt->get_ts();
    epoll_pair->start_time = epoll_pair->end_time - s_tinfo->m_latency;
    auto it = cache.find(s_tinfo->m_tid);
    if (it == cache.end()) {
        list = new std::list<info_base *>();
        cache[s_tinfo->m_tid] = list;
    } else {
        list = cache[s_tinfo->m_tid];
    }
    if (list->size() >= list_max_size) {
        auto tmp = list->front();
        list->pop_front();
        delete tmp;
    }
    list->push_back(epoll_pair);
    return true;
}

bool event_cache::setInfo(uint32_t tid, info_base *info) {
    //std::cout << "try to set: " << info->toStringTs() << std::endl;
    auto it = cache.find(tid);
    if (it == cache.end()) {
        if (info->exit == false) { // skip the first exit one
            auto list = new std::list<info_base *>();
            list->emplace_back(info);
            cache[tid] = list;
        }
    } else {
        auto list = it->second;
        info_base *info_pair = nullptr;
        if (list->size() > 0) {
            info_pair = list->back();
             // update exit event
            if (info_pair->exit == false) {
                // pair-event
                if (info->exit == true && info_pair->event_type == info->event_type - 1 && info->end_time - info_pair->start_time > threshold) {
                    info_pair->end_time = info->end_time;
                    info_pair->exit = true;
                } else { // lost exit event, delete
                    list->pop_back();
                    delete info_pair;
                }
            }
        }
        if (info->exit == false) {
            if (list->size() >= list_max_size) {
                auto tmp = list->front();
                list->pop_front();
                delete tmp;
            }
             list->emplace_back(info);
        } else {
            delete info;
        }
    }
    return true;
}
string event_cache::GetInfo(uint32_t tid, pair<uint64_t, uint64_t> &period, uint8_t off_type) {
    string result = "";
    // find
    auto it = cache.find(tid);
    if (it == cache.end()) {
        return result;
    }
    auto list = it->second;
    if (off_type != event_type) {
        return result;
    }
    auto f = list->begin();
    // clear: end_time  <  off.start
    while (f != list->end() && (*f)->end_time < period.first) {
        auto tmp = *f;
        f = list->erase(f);
        delete tmp;
    }
    // 搜索 start_time < off.start < end_time
    // 判断 if off.end < end_time && off.end - start_time > threshold -> result.append()
    if (f != list->end()) {
        if ((*f)->start_time < period.first && period.second < (*f)->end_time && period.second - (*f)->start_time > threshold) {
            result.append((*f)->toString());
        }
    }
    return result;
}

bool event_cache::setThreshold(uint64_t threshold_ms) {
    threshold = threshold_ms * 1000000;
    return true;
}

bool event_cache::clearList(uint32_t tid) {
    auto it = cache.find(tid);
    if (it != cache.end()) {
        if (it->second) {
            delete it->second;
        }
        std::cout << "Clear tid " << tid << "current map size: " << cache.size() <<endl;
        cache.erase(it);
    }
    return true;
}