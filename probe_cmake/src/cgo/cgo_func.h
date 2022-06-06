#pragma once
#ifndef SYSDIG_KINDLING_H
#define SYSDIG_KINDLING_H
#include "sinsp.h"
#include "kindling_event.h"

void init_probe();
void do_inspect();
int getEvent(void **kindlingEvent);

#endif //SYSDIG_KINDLING_H