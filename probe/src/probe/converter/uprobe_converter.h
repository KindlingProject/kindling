//
// Created by 千陆 on 2021/12/30.
//

#ifndef KINDLING_PROBE_UPROBE_CONVERTER_H
#define KINDLING_PROBE_UPROBE_CONVERTER_H

#include "src/probe/converter/converter.h"
#include "src/userspace/libsinsp/sinsp.h"

#include <iostream>
#include <vector>

using namespace kindling;


struct grpc_event_t {
    int64_t timestamp;
    int32_t pid;
//    int64_t fd;
    std::string remote_addr;
    int64_t remote_port;
    int64_t trace_role;
    std::string req_headers;
    std::string req_method;
    std::string req_path;
    int64_t resp_status;
    int64_t req_body_size;
    std::string req_body;
    int64_t resp_body_size;
    std::string resp_body;
    int64_t latency;
    std::string container_id;
};

class uprobe_converter : public converter {
public:
    void convert(void *evt);
    uprobe_converter();
    ~uprobe_converter();

private:
    sinsp *m_inspector;
};


#endif //KINDLING_PROBE_UPROBE_CONVERTER_H
