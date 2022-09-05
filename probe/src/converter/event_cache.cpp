//
// Created by 散养鸡 on 2022/5/20.
//

#include <iostream>
#include "event_cache.h"

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
                    info_pair->size = info->size;
                    info_pair->end_time = info->end_time;
                    if(info_pair->name.empty()){
                        info_pair->name = info->name;
                    }
                    info_pair->latency = info->latency;
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
    int count=0;
    // clear: end_time  <  off.start
    while (f != list->end() && (*f)->end_time < period.first) {
    	count++;
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

string event_cache::GetOnInfo(uint32_t tid, pair<uint64_t, uint64_t> &period) {
    string result = "";
    // find
    auto it = cache.find(tid);
    if (it == cache.end()) {
        return result;
    }
    auto list = it->second;
    auto f = list->begin();
    // out of range: end_time < on_start_time
    while (f != list->end() && (*f)->end_time < period.first) {
        auto tmp = *f;
        f = list->erase(f);
        delete tmp;
    }
    while (f != list->end()) {
        if (period.first < (*f)->start_time && (*f)->end_time < period.second) {
            if (result != "") {
                result.append("#");
            }
            result.append((*f)->toString());
            f++;
        } else {
            break;
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
//        std::cout << "Clear tid " << tid << "current map size: " << cache.size() <<endl;
        cache.erase(it);
    }
    return true;
}