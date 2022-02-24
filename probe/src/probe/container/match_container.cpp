//
// Created by jundi zhou on 2022/1/18.
//

#include "match_container.h"


const size_t CONTAINER_ID_LENGTH = 64;
const size_t REPORTED_CONTAINER_ID_LENGTH = 12;
const char* CONTAINER_ID_VALID_CHARACTERS = "0123456789abcdefABCDEF";


constexpr const cgroup_layout DOCKER_CGROUP_LAYOUT[] = {
{"/", ""}, // non-systemd docker
{"/docker-", ".scope"}, // systemd docker
{nullptr, nullptr}
};

constexpr const cgroup_layout CRI_CGROUP_LAYOUT[] = {
{"/", ""}, // non-systemd containerd
{"/crio-", ""}, // non-systemd cri-o
{"/cri-containerd-", ".scope"}, // systemd containerd
{"/crio-", ".scope"}, // systemd cri-o
{":cri-containerd:", ""}, // containerd without "SystemdCgroup = true"
{nullptr, nullptr}
};


bool match_container::match_one_container_id(const std::string &cgroup, const std::string &prefix, const std::string &suffix, std::string &container_id)
{
    size_t start_pos = cgroup.rfind(prefix);
    if (start_pos == std::string::npos)
    {
        return false;
    }
    start_pos += prefix.size();

    size_t end_pos = cgroup.rfind(suffix);
    if (end_pos == std::string::npos)
    {
        return false;
    }

    if (end_pos - start_pos != CONTAINER_ID_LENGTH)
    {
        return false;
    }

    size_t invalid_ch_pos = cgroup.find_first_not_of(CONTAINER_ID_VALID_CHARACTERS, start_pos);
    if (invalid_ch_pos < CONTAINER_ID_LENGTH)
    {
        return false;
    }

    container_id = cgroup.substr(start_pos, REPORTED_CONTAINER_ID_LENGTH);
    return true;
}

bool match_container::match_container_id(const std::string &cgroup, const cgroup_layout *layout,
                        std::string &container_id)
{
    for(size_t i = 0; layout[i].prefix && layout[i].suffix; ++i)
    {
        if(match_one_container_id(cgroup, layout[i].prefix, layout[i].suffix, container_id))
        {
            return true;
        }
    }

    return false;
}


bool match_container::matches_cris_cgroups(const sinsp_threadinfo *tinfo, std::string &container_id, const cgroup_layout *layout)
{

    for(const auto &it : tinfo->m_cgroups)
    {
        if(match_container_id(it.second, layout, container_id))
        {
            return true;
        }
    }

    return false;
}

bool match_container::matches_cgroups(const sinsp_threadinfo *tinfo, std::string &container_id)
{
    //todo: we use dockerd as cri more than conatinerd in china,We will change his order in the future
    if(matches_cris_cgroups(tinfo, container_id, DOCKER_CGROUP_LAYOUT)){
        return true;
    }
    if(matches_cris_cgroups(tinfo, container_id, CRI_CGROUP_LAYOUT)){
        return true;
    }

    return false;
}


