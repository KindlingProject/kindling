#include "converter.h"
#include <iostream>

converter::converter() {
    m_kindlingEventList = new kindling::KindlingEventList();
}

converter::converter(uint64_t batch_size, uint64_t max_size) : batch_size(batch_size), max_size(max_size) {
    m_kindlingEventList = new kindling::KindlingEventList();
}

converter::~converter() {
    delete m_kindlingEventList;
}

void converter::convert(void *evt) {}

bool converter::judge_max_size() {
    return m_kindlingEventList->kindling_event_list_size() >= max_size;
}

kindling::KindlingEventList *converter::get_kindlingEventList() {
    return m_kindlingEventList;
}

kindling::KindlingEventList * converter::swap_list(kindling::KindlingEventList *list) {
    std::swap(m_kindlingEventList, list);
    return list;
}

bool converter::judge_batch_size() {
    return m_kindlingEventList->kindling_event_list_size() >= batch_size;
}

uint64_t converter::current_list_size() {
    return m_kindlingEventList->kindling_event_list_size();
}
