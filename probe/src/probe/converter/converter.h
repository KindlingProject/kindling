#ifndef KINDLING_PROBE_CONVERTER_H
#define KINDLING_PROBE_CONVERTER_H
#include <string>
#include "src/probe/converter/kindling_event.pb.h"
#include <src/common/base/base.h>

class converter {
public:
    converter();
    converter(uint64_t batch_size, uint64_t max_size);
    virtual ~converter();
	// source evt -> kindling evt
	virtual void convert(void * evt);
    bool judge_batch_size();
	bool judge_max_size();
	uint64_t current_list_size();
    kindling::KindlingEventList* swap_list(kindling::KindlingEventList *);
    kindling::KindlingEventList* get_kindlingEventList();
private:
    kindling::KindlingEventList *m_kindlingEventList;
    uint64_t batch_size;
    uint64_t max_size;
};
#endif //KINDLING_PROBE_CONVERTER_H
