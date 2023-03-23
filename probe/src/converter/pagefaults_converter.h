#ifndef SYSDIG_PAGEFAULTS_H
#define SYSDIG_PAGEFAULTS_H
#include "../cgo/kindling.h"
#include <cstdlib>
#include <fstream>
#include <iostream>
#include <thread>

class pagefaults_analyzer {
    unordered_map<int64_t, threadinfo_map_t::ptr_t> threadstable;
    unordered_map<int64_t, int64_t> maj_mp, min_mp;  // from pid to maj or min value
    sinsp *inspector;
    uint64_t lasttime_read_bpf_map;
    pagefault_data* read_from_bpf_map;
public:
    pagefaults_analyzer(){};
    pagefaults_analyzer(sinsp *inspector){
        this->inspector = inspector;
        lasttime_read_bpf_map = 0;
        threadstable = inspector->m_thread_manager->get_threads()->getThreadsTable();
        convert_threadstable();
    };
    /* convert process information to main thread information.
    The page fault data we want is limited to thread granularity.*/
    void convert_threadstable();
    // from /proc
    int get_pagefaults_from_threadstable(kindling_event_t_for_go evt[], int* evtlen, int *maxlen);
    // from eBPF Map
    int get_pagefaults_from_bpf_map(kindling_event_t_for_go evt[], int* evtlen, int *maxlen);
    void init_pagefault_kindling_event(kindling_event_t_for_go *ptr, sinsp_threadinfo* threadInfo = nullptr);
    ~pagefaults_analyzer(){};
};

#endif