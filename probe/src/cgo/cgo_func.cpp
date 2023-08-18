//
// Created by jundi zhou on 2022/6/1.
//

#include "cgo_func.h"
#include "kindling.h"
#include "catch_sig.h"

int runForGo() { return init_probe(); }

int getEventsByInterval(int interval, void** kindlingEvent, void* count) { 
    return get_events_by_interval((uint64_t)interval, kindlingEvent, count); 
}

int startProfile() { return start_profile(); }
int stopProfile() { return stop_profile(); }

char* startAttachAgent(int pid) { return start_attach_agent(pid); }
char* stopAttachAgent(int pid) { return stop_attach_agent(pid); }

void suppressEventsCommForGo(char *comm) { suppress_events_comm(string(comm)); }
void subEventForGo(char* eventName, char* category, void *params) { sub_event(eventName, category, (event_params_for_subscribe *)params); }

void startProfileDebug(int pid, int tid) { start_profile_debug(pid, tid); }
void stopProfileDebug() { stop_profile_debug(); }

void getCaptureStatistics(struct capture_statistics_for_go* stats) { get_capture_statistics(stats); }
void catchSignalUp() { sig_set_up(); }

