//
// Created by 散养鸡 on 2022/6/22.
//

#include "epoll_event_cache.h"
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
        for (auto efd : dynamic_cast<epoll_info*>(epoll_pair)->fds) {
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
