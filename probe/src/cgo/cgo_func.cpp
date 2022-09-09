//
// Created by jundi zhou on 2022/6/1.
//

#include "cgo_func.h"
#include "kindling.h"


void runForGo(){
	init_probe();
}

void startPerf() {
	start_perf();
}

void stopPerf() {
	stop_perf();
}

void expireWindowCache() {
	exipre_window_cache();
}

int getKindlingEvent(void **kindlingEvent){
	return getEvent(kindlingEvent);
}


void subEventForGo(char* eventName, char* category){
	sub_event(eventName, category);
}
