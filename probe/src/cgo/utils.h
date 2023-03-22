// Created by Daxin Wang on 2023/3/17.
#ifndef KINDLING_UTILS_H
#define KINDLING_UTILS_H

#include "kindling.h"

void fill_kindling_event_param(kindling_event_t_for_go* p_kindling_event, KeyValue raw_params[],
                               int raw_param_len, int& userAttNumber);

void printCurrentTime();

//
// Get the string representation of a ppm_event_type
//
std::string get_event_type(uint16_t type);

#endif
