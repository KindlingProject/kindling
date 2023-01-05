//
// Created by jundi zhou on 2022/6/1.
//

#ifndef SYSDIG_CGO_FUNC_H
#define SYSDIG_CGO_FUNC_H

#ifdef __cplusplus
extern "C" {
#endif
int runForGo();
int getKindlingEvent(void** kindlingEvent);
void subEventForGo(char* eventName, char* category, void* params);
int startProfile();
int stopProfile();
int startAttachAgent(int pid);
int stopAttachAgent(int pid);
void startProfileDebug(int pid, int tid);
void stopProfileDebug();
void getCaptureStatistics();
void catchSignalUp();
#ifdef __cplusplus
}
#endif

#endif  // SYSDIG_CGO_FUNC_H
