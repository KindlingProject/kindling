//
// Created by 散养鸡 on 2022/4/29.
//

#include "cpu_converter.h"
#include <vector>
using namespace std;
using namespace kindling;

cpu_converter::cpu_converter(sinsp *inspector, Profiler *prof, LogCache *log) : m_inspector(inspector), m_profiler(prof), m_log(log) {
    file_cache = new event_cache(1);
    net_cache = new event_cache(2);
    epoll_cache = new epoll_event_cache(4);
}

cpu_converter::~cpu_converter() {
    delete file_cache;
    delete net_cache;
    delete epoll_cache;
}

bool cpu_converter::Cache(sinsp_evt *sevt) {
    info_base *info = nullptr;
    auto type = sevt->get_type();
    if (type == PPME_SYSCALL_EPOLLWAIT_X) {
        return epoll_cache->setInfo(sevt);
    }
    sinsp_evt::category cat;
    sevt->get_category(&cat);
    auto s_tinfo = sevt->get_thread_info();
    if (type == PPME_PROCEXIT_1_E || type == PPME_PROCEXIT_E) {
        net_cache->clearList(s_tinfo->m_tid);
        file_cache->clearList(s_tinfo->m_tid);
        epoll_cache->clearList(s_tinfo->m_tid);
    }
    if (!(cat.m_category == EC_IO_WRITE || cat.m_category == EC_IO_READ)) {
        return false;
    }
    auto s_fdinfo = sevt->get_fd_info();
    if (s_fdinfo == nullptr) {
        return false;
    }
    if (PPME_IS_ENTER(type)) {
        switch (s_fdinfo->m_type) {
            case SCAP_FD_FILE:
            case SCAP_FD_FILE_V2: {
                info = new file_info();
                info->start_time = sevt->get_ts();
                info->name = s_fdinfo->m_name;
                auto psize = sevt->get_param_value_raw("size");
                if (!psize || *(uint32_t *) psize->m_val <= 0) {
                    return false;
                }
                info->size = *(uint32_t *) psize->m_val;
                info->operation_type = (cat.m_category == EC_IO_READ) ? "read" : "write";
                break;
            }
            case SCAP_FD_IPV4_SOCK:
            case SCAP_FD_IPV4_SERVSOCK: {
                info = new net_info();
                info->start_time = sevt->get_ts();
                info->name = s_fdinfo->m_name;
                auto psize = sevt->get_param_value_raw("size");
                if (!psize || *(uint32_t *) psize->m_val <= 0) {
                    return false;
                }
                info->size = *(uint32_t *) psize->m_val;
                info->operation_type = (cat.m_category == EC_IO_READ) ? "read" : "write";

                epoll_cache->SetLastEpollCache(s_tinfo->m_tid, sevt->get_fd_num(), info);
                break;
            }
            default:
                return false;
        }
    } else {
        switch (s_fdinfo->m_type) {
            case SCAP_FD_FILE:
            case SCAP_FD_FILE_V2: {
                info = new file_info();
                break;
            }
            case SCAP_FD_IPV4_SOCK:
            case SCAP_FD_IPV4_SERVSOCK: {
                info = new net_info();
                break;
            }
            default:
                return false;
        }
        info->end_time = sevt->get_ts();
        info->exit = true;
    }

    info->event_type = static_cast<uint16_t>(type);

    switch (s_fdinfo->m_type) {
		case SCAP_FD_FILE:
        case SCAP_FD_FILE_V2:
            return file_cache->setInfo(s_tinfo->m_tid, info);
        case SCAP_FD_IPV4_SOCK:
        case SCAP_FD_IPV4_SERVSOCK: {
            return net_cache->setInfo(s_tinfo->m_tid, info);
        }
        default:
            return false;
    }
}

void cpu_converter::convert(void *evt)
{
    sinsp_evt *cpu_evt = static_cast<sinsp_evt*> (evt);
    uint64_t start_time = *reinterpret_cast<uint64_t*> (cpu_evt->get_param_value_raw("start_ts")->m_val);
    uint64_t end_time = *reinterpret_cast<uint64_t*> (cpu_evt->get_param_value_raw("end_ts")->m_val);
    
    // convert
    auto kevt = get_kindlingEventList()->add_kindling_event_list();
    init_kindling_event(kevt, cpu_evt);
    add_threadinfo(kevt, cpu_evt);
    add_cpu_data(kevt, cpu_evt);
}

void merge()
{
    return;
}
void split()
{
    return;
}

int cpu_converter::init_kindling_event(kindling::KindlingEvent* kevt, sinsp_evt *sevt)
{
    kevt->set_source(TRACEPOINT);
    kevt->set_name("cpu_event");

    return 0;
}

int cpu_converter::add_threadinfo(kindling::KindlingEvent* kevt, sinsp_evt *sevt)
{
    auto s_tinfo = sevt->get_thread_info();
    if (!s_tinfo) {
        return -1;
    }
    auto k_tinfo = kevt->mutable_ctx()->mutable_thread_info();
    k_tinfo->set_pid(s_tinfo->m_pid);
    k_tinfo->set_tid(s_tinfo->m_tid);
    k_tinfo->set_comm(s_tinfo->m_comm);
    // k_tinfo->set_container_id(s_tinfo->m_container_id);
    return 0;
}

int cpu_converter::add_cpu_data(KindlingEvent* kevt, sinsp_evt *sevt)
{
    uint64_t start_time = *reinterpret_cast<uint64_t*> (sevt->get_param_value_raw("start_ts")->m_val);
    uint64_t end_time = *reinterpret_cast<uint64_t*> (sevt->get_param_value_raw("end_ts")->m_val);
    uint32_t cnt = *reinterpret_cast<uint32_t*> (sevt->get_param_value_raw("cnt")->m_val);
    uint64_t *time_specs = reinterpret_cast<uint64_t *> (sevt->get_param_value_raw("time_specs")->m_val);
    uint64_t *runq_latency = reinterpret_cast<uint64_t *> (sevt->get_param_value_raw("runq_latency")->m_val);
    uint8_t *time_type = reinterpret_cast<uint8_t *> (sevt->get_param_value_raw("time_type")->m_val);
    cpu_data c_data;
    vector<pair<uint64_t, uint64_t>> on_time, off_time;
    vector<uint8_t> off_type;
    uint64_t start = start_time;
    for (int i = 0; i < cnt; i++) {
        if (time_type[i] == 0) {
            c_data.on_total_time += time_specs[i];
            on_time.push_back({start, start + time_specs[i] * 1000});
        } else {
            c_data.off_total_time += time_specs[i];
            c_data.runq_latency += (to_string(runq_latency[i/2]) + ",");
            off_time.push_back({start, start + time_specs[i] * 1000});
            off_type.push_back(time_type[i]);
        }
        start = start + time_specs[i] * 1000;
        c_data.time_specs += (to_string(time_specs[i]) + ",");
        c_data.time_type += (to_string(time_type[i]) +  ",");
    }
    // on_total_time
    auto off_attr = kevt->add_user_attributes();
    off_attr->set_key("on_total_time");
    off_attr->set_value(&c_data.on_total_time, 8);
    off_attr->set_value_type(UINT64);

    // off_total_time
    off_attr = kevt->add_user_attributes();
    off_attr->set_key("off_total_time");
    off_attr->set_value(&c_data.off_total_time, 8);
    off_attr->set_value_type(UINT64);

    // start_time
    off_attr = kevt->add_user_attributes();
    off_attr->set_key("start_time");
    off_attr->set_value(&c_data.start_time, 8);
    off_attr->set_value_type(UINT64);

    // end_time
    off_attr = kevt->add_user_attributes();
    off_attr->set_key("end_time");
    off_attr->set_value(&c_data.end_time, 8);
    off_attr->set_value_type(UINT64);

    // time_specs
    off_attr = kevt->add_user_attributes();
    off_attr->set_key("type_specs");
    off_attr->set_value(c_data.time_specs);
    off_attr->set_value_type(CHARBUF);

    // runq_latency
    off_attr = kevt->add_user_attributes();
    off_attr->set_key("runq_latency");
    off_attr->set_value(c_data.runq_latency);
    off_attr->set_value_type(CHARBUF);

    // time_type
    off_attr = kevt->add_user_attributes();
    off_attr->set_key("time_type");
    off_attr->set_value(c_data.time_type);
    off_attr->set_value_type(CHARBUF);

    // on_stack
    auto s_tinfo = sevt->get_thread_info();
    string data = m_profiler->GetOnCpuData(s_tinfo->m_tid, on_time);
    if (data != "") {
        // LOG(INFO) << "related stack: " << data;
        auto on_attr = kevt->add_user_attributes();
        on_attr->set_key("stack");
        on_attr->set_value(data);
        on_attr->set_value_type(CHARBUF);
    }
    auto log_msg = m_log->getLogs(s_tinfo->m_tid, on_time);
    if (log_msg != "") {
        // LOG(INFO) << "related log: " << log_msg;
        auto log_attr = kevt->add_user_attributes();
        log_attr->set_key("log");
        log_attr->set_value(log_msg);
        log_attr->set_value_type(CHARBUF);
    }

    string info = "";
    for (int i = 0; i < off_time.size(); i++) {
        switch (off_type[i]) {
            case 1: {
                info.append(file_cache->GetInfo(s_tinfo->m_tid, off_time[i], off_type[i]));
                break;
            }
            case 2: {
                info.append(net_cache->GetInfo(s_tinfo->m_tid, off_time[i], off_type[i]));
                break;
            }
            case 4: {
                info.append(epoll_cache->GetInfo(s_tinfo->m_tid, off_time[i], off_type[i]));
                break;
            }
        }
        info.append("|");
    }
    if (info.length() != off_time.size()) {
        LOG(INFO) << "related off info: " << info;
        auto file_attr = kevt->add_user_attributes();
        file_attr->set_key("off info");
        file_attr->set_value(info);
        file_attr->set_value_type(CHARBUF);
    }
    // merge();
    // analyse()
    return 0;
}