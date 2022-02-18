//
// Created by jundi zhou on 2022/1/18.
//

#ifndef KINDLING_PROBE_MATCH_CONTAINER_H
#define KINDLING_PROBE_MATCH_CONTAINER_H

#include <string>
#include "sinsp.h"

struct cgroup_layout {
    const char* prefix;
    const char* suffix;
};

class match_container {
public:
    bool match_one_container_id(const std::string &cgroup, const std::string &prefix, const std::string &suffix, std::string &container_id);
    bool match_container_id(const std::string &cgroup, const cgroup_layout *layout,
                       std::string &container_id);
    bool matches_cgroups(const sinsp_threadinfo *tinfo, std::string &container_id);
    bool matches_cris_cgroups(const sinsp_threadinfo *tinfo, std::string &container_id, const cgroup_layout *layout);

};


#endif //KINDLING_PROBE_MATCH_CONTAINER_H
