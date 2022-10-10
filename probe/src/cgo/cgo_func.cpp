//
// Created by jundi zhou on 2022/6/1.
//

#include "cgo_func.h"
#include "kindling.h"


int runForGo(){
	return init_probe();
}

int getKindlingEvent(void **kindlingEvent){
	return getEvent(kindlingEvent);
}


void subEventForGo(char* eventName, char* category){
	sub_event(eventName, category);
}

int startProfile() {
    start_profile();
}
int stopProfile() {
    stop_profile();
}