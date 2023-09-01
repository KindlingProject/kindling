//
// Created by jundi zhou on 2022/6/1.
//

#ifndef SYSDIG_CGO_FUNC_H
#define SYSDIG_CGO_FUNC_H

#ifdef __cplusplus
extern "C" {
#endif
int runForGo();
int getKindlingEvent();
void suppressEventsCommForGo(char *comm);
void suppressEventsThreadForGo(char *thread);
void subEventForGo(char* eventName, char* category, void* params);
int initKindlingEventForGo(int number, void *kindlingEvent);
int getEventsByInterval(int interval, void *kindlingEvent, void *count);
int startProfile();
int stopProfile();
char* startAttachAgent(int pid);
char* stopAttachAgent(int pid);
void startProfileDebug(int pid, int tid);
void stopProfileDebug();
void getCaptureStatistics();
void catchSignalUp();
void sampleUrl(char* pidUrl, int sampled);
#ifdef __cplusplus
}
#endif

#endif  // SYSDIG_CGO_FUNC_H
