#include "../cgo/kindling.h"
#include "../cgo/utils.h"
#include "pagefaults_converter.h"
/* convert process information to main thread information.
  The page fault data we want is limited to thread granularity.*/
void pagefaults_analyzer::convert_threadstable() {
  for (auto e : threadstable) {
    sinsp_threadinfo* tmp = e.second.get();
    if (tmp->m_pid == tmp->m_tid) continue;
    maj_mp[tmp->m_pid] += tmp->m_pfmajor;
    min_mp[tmp->m_pid] += tmp->m_pfminor;
  }

  for (auto e : min_mp) {
    auto tmp = threadstable.find(e.first);
    sinsp_threadinfo* temp = inspector->build_threadinfo();
    temp->m_pid = temp->m_tid = e.first;
    temp->m_pfminor = tmp->second->m_pfminor - e.second;
    temp->m_pfmajor = tmp->second->m_pfmajor - maj_mp[e.first];
    threadstable[temp->m_tid] = threadinfo_map_t::ptr_t(temp);
  }

  cout << "total number of page fault threads initialized is " << threadstable.size() << endl;
}

void pagefaults_analyzer::init_pagefault_kindling_event(kindling_event_t_for_go* p_kindling_event,
                                                        sinsp_threadinfo* threadInfo) {
  p_kindling_event->name = (char*)malloc(sizeof(char) * 1024);
  p_kindling_event->context.tinfo.comm = (char*)malloc(sizeof(char) * 256);
  p_kindling_event->context.tinfo.containerId = (char*)malloc(sizeof(char) * 256);
  p_kindling_event->context.fdInfo.filename = (char*)malloc(sizeof(char) * 1024);
  p_kindling_event->context.fdInfo.directory = (char*)malloc(sizeof(char) * 1024);

  chrono::nanoseconds ns = std::chrono::duration_cast<std::chrono::nanoseconds>(
      std::chrono::system_clock::now().time_since_epoch());
  p_kindling_event->timestamp = ns.count();

  if (threadInfo) {
    p_kindling_event->context.tinfo.pid = threadInfo->m_pid;
    p_kindling_event->context.tinfo.tid = threadInfo->m_tid;
    p_kindling_event->context.tinfo.uid = threadInfo->m_uid;
    p_kindling_event->context.tinfo.gid = threadInfo->m_gid;
    strcpy(p_kindling_event->context.tinfo.comm, (char*)threadInfo->m_comm.data());
    strcpy(p_kindling_event->context.tinfo.containerId, (char*)threadInfo->m_container_id.data());
  }

  for (int i = 0; i < 2; i++) {
    p_kindling_event->userAttributes[i].key = (char*)malloc(sizeof(char) * 128);
    p_kindling_event->userAttributes[i].value = (char*)malloc(sizeof(char) * 1024);
  }
}

int pagefaults_analyzer::get_pagefaults_from_threadstable(kindling_event_t_for_go evt[],
                                                          int* evtlen, int* maxlen) {
  int evtcnt = 0;

  for (auto& e : threadstable) {
    sinsp_threadinfo* threadInfo = e.second.get();
    init_pagefault_kindling_event(&evt[evtcnt], threadInfo);
    strcpy(evt[evtcnt].name, "page_fault");
    int userAttNumber = 0;
    KeyValue kindling_event_params[2] = {
        {(char*)("pgft_maj"), (char*)(&threadInfo->m_pfmajor), 8, UINT64},
        {(char*)("pgft_min"), (char*)(&threadInfo->m_pfminor), 8, UINT64},
    };
    fill_kindling_event_param(&evt[evtcnt], kindling_event_params, 2, userAttNumber);
    evt[evtcnt].paramsNumber = userAttNumber;
    evtcnt++;
    if (evtcnt >= *maxlen) break;
  }
  *evtlen = evtcnt;
  return 0;
}

int pagefaults_analyzer::get_pagefaults_from_bpf_map(kindling_event_t_for_go evt[], int* evtlen,
                                                     int* maxlen) {
  pagefault_data* read_from_bpf_map = new pagefault_data[*maxlen];
  int rescnt = 0;
  int evtcnt = 0;
  chrono::nanoseconds ns = std::chrono::duration_cast<std::chrono::nanoseconds>(
      std::chrono::system_clock::now().time_since_epoch());
  uint64_t cur = ns.count();
  int32_t ret = inspector->get_page_faults_from_map(lasttime_read_bpf_map, cur, read_from_bpf_map,
                                                    &rescnt, *maxlen);
  lasttime_read_bpf_map = cur;
  if (ret == SCAP_FAILURE) {
    delete[] read_from_bpf_map;
    return -1;
  }

  for (int i = 0; i < rescnt; i++) {
    sinsp_threadinfo* threadInfo = inspector->m_thread_manager->get_threads()->get(read_from_bpf_map[i].tid);
    init_pagefault_kindling_event(&evt[evtcnt], threadInfo);
    strcpy(evt[evtcnt].name, "page_fault");
    evt[evtcnt].context.tinfo.pid = read_from_bpf_map[i].pid;
    evt[evtcnt].context.tinfo.tid = read_from_bpf_map[i].tid;
    evt[evtcnt].timestamp = read_from_bpf_map[i].timestamp;

    int userAttNumber = 0;

    KeyValue kindling_event_params[2] = {
        {(char*)("pgft_maj"), (char*)(&read_from_bpf_map[i].maj_flt), 8, UINT64},
        {(char*)("pgft_min"), (char*)(&read_from_bpf_map[i].min_flt), 8, UINT64},
    };

    fill_kindling_event_param(&evt[evtcnt], kindling_event_params, 2, userAttNumber);
    evt[evtcnt].paramsNumber = userAttNumber;
    evtcnt++;
    if (evtcnt >= *maxlen) break;
  }
  *evtlen = evtcnt;
  delete[] read_from_bpf_map;
  return ret;
}
