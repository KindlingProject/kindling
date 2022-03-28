#include "src/probe/converter/sysdig_converter.h"
#include <map>
#include <iostream>

using namespace std;
using namespace kindling;


sysdig_converter::sysdig_converter(sinsp *inspector) : converter(100, INT_MAX), m_inspector(inspector) {}

sysdig_converter::sysdig_converter(sinsp *inspector, int batch_size, int max_size) : converter(batch_size, max_size), m_inspector(inspector) {}

sysdig_converter::~sysdig_converter() {}

void sysdig_converter::convert(void *evt) {
    auto kevt = get_kindlingEventList()->add_kindling_event_list();
    sinsp_evt *sevt = (sinsp_evt *) evt;

    init_kindling_event(kevt, sevt);
    add_native_attributes(kevt, sevt);
    add_user_attributes(kevt, sevt);
    add_fdinfo(kevt, sevt);
    add_threadinfo(kevt, sevt);
}

// set source, name, timestamp, category according to list
int sysdig_converter::init_kindling_event(kindling::KindlingEvent *kevt, sinsp_evt *sevt) {
    kevt->set_source(get_kindling_source(sevt->get_type()));
    kevt->set_name(get_kindling_name(sevt));
    kevt->set_category(get_kindling_category(sevt));
    kevt->set_timestamp(sevt->get_ts());
    return 0;
}

int sysdig_converter::add_native_attributes(kindling::KindlingEvent *kevt, sinsp_evt *sevt) {
    // TODO
    return 0;
}


int sysdig_converter::add_user_attributes(kindling::KindlingEvent *kevt, sinsp_evt *sevt) {
    auto s_tinfo = sevt->get_thread_info();
    if (!s_tinfo) {
        return -1;
    }

    if (kevt->source() == SYSCALL_EXIT) {
        // set latency
        auto latency_attr = kevt->add_user_attributes();
        auto latency = s_tinfo->m_latency;
        latency_attr->set_key("latency");
        latency_attr->set_value_type(UINT64);
        latency_attr->set_value(&latency, 8);
    }
    // set params
    switch (sevt->get_type()) {
        case PPME_TCP_RCV_ESTABLISHED_E:
        case PPME_TCP_CLOSE_E: {
            auto pTuple = sevt->get_param_value_raw("tuple");
            setTuple(kevt, pTuple);

            auto pRtt = sevt->get_param_value_raw("srtt");
            if (pRtt != NULL) {
                auto attr = kevt->add_user_attributes();
                attr->set_key("rtt");
                attr->set_value(pRtt->m_val, pRtt->m_len);
                attr->set_value_type(UINT32);
            }
            break;
        }
        case PPME_TCP_DROP_E:
        case PPME_TCP_RETRANCESMIT_SKB_E: {
            auto pTuple = sevt->get_param_value_raw("tuple");
            setTuple(kevt, pTuple);
            break;
        }
        default:
            for (auto i = 0; i < sevt->get_num_params(); i++) {
                auto attr = kevt->add_user_attributes();
                attr->set_key(sevt->get_param_name(i));
                attr->set_value(sevt->get_param(i)->m_val, sevt->get_param(i)->m_len);
                attr->set_value_type(get_type(sevt->get_param_info(i)->type));
            }
    }
    return 0;
}

int sysdig_converter::add_fdinfo(kindling::KindlingEvent *kevt, sinsp_evt *sevt) {
    auto s_fdinfo = sevt->get_fd_info();
    if (!s_fdinfo) {
        return -1;
    }
    auto k_fdinfo = kevt->mutable_ctx()->mutable_fd_info();
    k_fdinfo->set_num(sevt->get_fd_num());
    // set type one-one relation
    k_fdinfo->set_type_fd(FDType(s_fdinfo->m_type));
    switch (s_fdinfo->m_type) {
        case SCAP_FD_FILE:
        case SCAP_FD_FILE_V2: {
            string name = s_fdinfo->m_name;
            size_t pos = name.rfind('/');
            if (pos != string::npos) {
                if (pos < name.size() - 1) {
                    k_fdinfo->set_filename(name.substr(pos + 1, string::npos));
                    if (pos != 0) {
                        name.resize(pos);
                        k_fdinfo->set_directory(name);
                    } else {
                        k_fdinfo->set_directory("/");
                    }
                }
            }
            break;
        }
        case SCAP_FD_IPV4_SOCK:
        case SCAP_FD_IPV4_SERVSOCK:
            k_fdinfo->set_protocol(get_protocol(s_fdinfo->get_l4proto()));
            k_fdinfo->set_role(s_fdinfo->is_role_server());
            k_fdinfo->add_sip(s_fdinfo->m_sockinfo.m_ipv4info.m_fields.m_sip);
            k_fdinfo->add_dip(s_fdinfo->m_sockinfo.m_ipv4info.m_fields.m_dip);
            k_fdinfo->set_sport(s_fdinfo->m_sockinfo.m_ipv4info.m_fields.m_sport);
            k_fdinfo->set_dport(s_fdinfo->m_sockinfo.m_ipv4info.m_fields.m_dport);
            break;
        case SCAP_FD_UNIX_SOCK:
            k_fdinfo->set_source(s_fdinfo->m_sockinfo.m_unixinfo.m_fields.m_source);
            k_fdinfo->set_destination(s_fdinfo->m_sockinfo.m_unixinfo.m_fields.m_dest);
            break;
        default:
            break;
    }
    return 0;
}

int sysdig_converter::add_threadinfo(kindling::KindlingEvent *kevt, sinsp_evt *sevt) {
    auto s_tinfo = sevt->get_thread_info();
    if (!s_tinfo) {
        return -1;
    }
    auto k_tinfo = kevt->mutable_ctx()->mutable_thread_info();
    k_tinfo->set_pid(s_tinfo->m_pid);
    k_tinfo->set_tid(s_tinfo->m_tid);
    k_tinfo->set_uid(s_tinfo->m_uid);
    k_tinfo->set_gid(s_tinfo->m_gid);
    k_tinfo->set_comm(s_tinfo->m_comm);
    k_tinfo->set_container_id(s_tinfo->m_container_id);
    return 0;
}

Source sysdig_converter::get_kindling_source(uint16_t etype) {
    if (PPME_IS_ENTER(etype)) {
        switch (etype) {
            case PPME_PROCEXIT_E:
            case PPME_SCHEDSWITCH_6_E:
            case PPME_SYSDIGEVENT_E:
            case PPME_CONTAINER_E:
            case PPME_PROCINFO_E:
            case PPME_SCHEDSWITCH_1_E:
            case PPME_DROP_E:
            case PPME_PROCEXIT_1_E:
            case PPME_CPU_HOTPLUG_E:
            case PPME_K8S_E:
            case PPME_TRACER_E:
            case PPME_MESOS_E:
            case PPME_CONTAINER_JSON_E:
            case PPME_NOTIFICATION_E:
            case PPME_INFRASTRUCTURE_EVENT_E:
            case PPME_PAGE_FAULT_E:
                return SOURCE_UNKNOWN;
            case PPME_TCP_RCV_ESTABLISHED_E:
            case PPME_TCP_CLOSE_E:
            case PPME_TCP_DROP_E:
            case PPME_TCP_RETRANCESMIT_SKB_E:
                return KRPOBE;
                // TODO add cases of tracepoint, kprobe, uprobe
            default:
                return SYSCALL_ENTER;
        }
    } else {
        switch (etype) {
            case PPME_CONTAINER_X:
            case PPME_PROCINFO_X:
            case PPME_SCHEDSWITCH_1_X:
            case PPME_DROP_X:
            case PPME_CPU_HOTPLUG_X:
            case PPME_K8S_X:
            case PPME_TRACER_X:
            case PPME_MESOS_X:
            case PPME_CONTAINER_JSON_X:
            case PPME_NOTIFICATION_X:
            case PPME_INFRASTRUCTURE_EVENT_X:
            case PPME_PAGE_FAULT_X:
                return SOURCE_UNKNOWN;
                // TODO add cases of tracepoint, kprobe, uprobe
            default:
                return SYSCALL_EXIT;
        }
    }
}

Category sysdig_converter::get_kindling_category(sinsp_evt *sEvt) {
    sinsp_evt::category cat;
    sEvt->get_category(&cat);
    switch (cat.m_category) {
        case EC_OTHER:
            return CAT_OTHER;
        case EC_FILE:
            return CAT_FILE;
        case EC_NET:
            return CAT_NET;
        case EC_IPC:
            return CAT_IPC;
        case EC_MEMORY:
            return CAT_MEMORY;
        case EC_PROCESS:
            return CAT_PROCESS;
        case EC_SLEEP:
            return CAT_SLEEP;
        case EC_SYSTEM:
            return CAT_SYSTEM;
        case EC_SIGNAL:
            return CAT_SIGNAL;
        case EC_USER:
            return CAT_USER;
        case EC_TIME:
            return CAT_TIME;
        case EC_IO_READ:
        case EC_IO_WRITE:
        case EC_IO_OTHER: {
            switch (cat.m_subcategory) {
                case sinsp_evt::SC_FILE:
                    return CAT_FILE;
                case sinsp_evt::SC_NET:
                    return CAT_NET;
                case sinsp_evt::SC_IPC:
                    return CAT_IPC;
                default:
                    return CAT_OTHER;
            }
        }
        default:
            return CAT_OTHER;
    }
}

L4Proto sysdig_converter::get_protocol(scap_l4_proto proto) {
    switch (proto) {
        case SCAP_L4_TCP:
            return TCP;
        case SCAP_L4_UDP:
            return UDP;
        case SCAP_L4_ICMP:
            return ICMP;
        case SCAP_L4_RAW:
            return RAW;
        default:
            return UNKNOWN;
    }
}

ValueType sysdig_converter::get_type(ppm_param_type type) {
    switch (type) {
        case PT_INT8:
            return INT8;
        case PT_INT16:
            return INT16;
        case PT_INT32:
            return INT32;
        case PT_INT64:
        case PT_FD:
        case PT_PID:
        case PT_ERRNO:
            return INT64;
        case PT_FLAGS8:
        case PT_UINT8:
        case PT_SIGTYPE:
            return UINT8;
        case PT_FLAGS16:
        case PT_UINT16:
        case PT_SYSCALLID:
            return UINT16;
        case PT_UINT32:
        case PT_FLAGS32:
        case PT_MODE:
        case PT_UID:
        case PT_GID:
        case PT_BOOL:
        case PT_SIGSET:
            return UINT32;
        case PT_UINT64:
        case PT_RELTIME:
        case PT_ABSTIME:
            return UINT64;
        case PT_CHARBUF:
        case PT_FSPATH:
            return CHARBUF;
        case PT_BYTEBUF:
            return BYTEBUF;
        case PT_DOUBLE:
            return DOUBLE;
        case PT_SOCKADDR:
        case PT_SOCKTUPLE:
        case PT_FDLIST:
        default:
            return BYTEBUF;
    }
}

string sysdig_converter::get_kindling_name(sinsp_evt *pEvt) {
    // TODO
    return pEvt->get_name();
}

int sysdig_converter::setTuple(kindling::KindlingEvent* kevt, const sinsp_evt_param *pTuple) {
    if (NULL != pTuple) {
        auto tuple = pTuple->m_val;
        if (tuple[0] == PPM_AF_INET) {
            if (pTuple->m_len == 1 + 4 + 2 + 4 + 2) {
                auto sip = kevt->add_user_attributes();
                sip->set_key("sip");
                sip->set_value(tuple+1, 4);
                sip->set_value_type(UINT32);

                auto sport = kevt->add_user_attributes();
                sport->set_key("sport");
                sport->set_value(tuple+5, 2);
                sport->set_value_type(UINT16);

                auto dip = kevt->add_user_attributes();
                dip->set_key("dip");
                dip->set_value(tuple+7, 4);
                dip->set_value_type(UINT32);

                auto dport = kevt->add_user_attributes();
                dport->set_key("dport");
                dport->set_value(tuple+11, 2);
                dport->set_value_type(UINT16);
            }
        }
    }
    return 0;
}
