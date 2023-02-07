//
// Created by jundi zhou on 2022/6/1.
//

#include "cgo_func.h"
#include "kindling.h"
#include "catch_sig.h"

int runForGo() { return init_probe(); }

int getKindlingEvent(void** kindlingEvent) { return getEvent(kindlingEvent); }


int startProfile() { return start_profile(); }
int stopProfile() { return stop_profile(); }

char* startAttachAgent(int pid) { return start_attach_agent(pid); }
char* stopAttachAgent(int pid) { return stop_attach_agent(pid); }

void subEventForGo(char* eventName, char* category, void *params) { sub_event(eventName, category, (event_params_for_subscribe *)params); }
void startProfileDebug(int pid, int tid) { start_profile_debug(pid, tid); }

void stopProfileDebug() { stop_profile_debug(); }

void getCaptureStatistics() { get_capture_statistics(); }
void catchSignalUp() { sig_set_up(); }