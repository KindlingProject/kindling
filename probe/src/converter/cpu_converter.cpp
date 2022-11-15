//
// Created by 散养鸡 on 2022/4/29.
//

#include "cpu_converter.h"
#include <vector>
#include "iostream"
#include <fstream>


fstream debug_file;
using namespace std;

cpu_converter::cpu_converter(sinsp *inspector) : m_inspector(inspector) {
}


cpu_converter::~cpu_converter() {
}


int cpu_converter::convert(kindling_event_t_for_go *p_kindling_event, sinsp_evt *cpu_evt, vector<QObject *> qls, bool is_profile_debug, int64_t debug_pid, int64_t debug_tid) {
    // convert
    init_kindling_event(p_kindling_event, cpu_evt);
    add_threadinfo(p_kindling_event, cpu_evt);
    add_cpu_data(p_kindling_event, cpu_evt, qls, is_profile_debug, debug_pid, debug_tid);
    return 1;
}


int cpu_converter::init_kindling_event(kindling_event_t_for_go *p_kindling_event, sinsp_evt *sevt) {
	strcpy(p_kindling_event->name, "cpu_event");

	return 0;
}

int cpu_converter::add_threadinfo(kindling_event_t_for_go *p_kindling_event, sinsp_evt *evt) {
	auto threadInfo = evt->get_thread_info();
	if (!threadInfo) {
		return -1;
	}
	p_kindling_event->context.tinfo.pid = threadInfo->m_pid;
	p_kindling_event->context.tinfo.tid = threadInfo->m_tid;
	strcpy(p_kindling_event->context.tinfo.containerId, (char *) threadInfo->m_container_id.data());

	return 0;
}

int cpu_converter::add_cpu_data(kindling_event_t_for_go *p_kindling_event, sinsp_evt *sevt, vector<QObject *> qls, bool is_profile_debug, int64_t debug_pid, int64_t debug_tid) {
    uint64_t start_time = *reinterpret_cast<uint64_t *> (sevt->get_param_value_raw("start_ts")->m_val);
    uint64_t end_time = *reinterpret_cast<uint64_t *> (sevt->get_param_value_raw("end_ts")->m_val);
    uint32_t cnt = *reinterpret_cast<uint32_t *> (sevt->get_param_value_raw("cnt")->m_val);
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
			on_time.emplace_back(start, start + time_specs[i]);
		} else {
			c_data.off_total_time += time_specs[i];
			c_data.runq_latency += (to_string(runq_latency[i / 2]) + ",");
			off_time.emplace_back(start, start + time_specs[i]);

			off_type.emplace_back(time_type[i]);
		}
		start = start + time_specs[i];
		c_data.time_specs += (to_string(time_specs[i]) + ",");
		c_data.time_type += (to_string(time_type[i]) + ",");
	}

	uint16_t userAttNumber = 0;
	// on_total_time
	strcpy(p_kindling_event->userAttributes[userAttNumber].key, "on_total_time");
	memcpy(p_kindling_event->userAttributes[userAttNumber].value, &c_data.on_total_time, 8);
	p_kindling_event->userAttributes[userAttNumber].valueType = UINT64;
	p_kindling_event->userAttributes[userAttNumber].len = 8;
	userAttNumber++;

	// off_total_time
	strcpy(p_kindling_event->userAttributes[userAttNumber].key, "off_total_time");
	memcpy(p_kindling_event->userAttributes[userAttNumber].value, &c_data.off_total_time, 8);
	p_kindling_event->userAttributes[userAttNumber].valueType = UINT64;
	p_kindling_event->userAttributes[userAttNumber].len = 8;
	userAttNumber++;

	// start_time
	strcpy(p_kindling_event->userAttributes[userAttNumber].key, "start_time");
	memcpy(p_kindling_event->userAttributes[userAttNumber].value, &start_time, 8);
	p_kindling_event->userAttributes[userAttNumber].valueType = UINT64;
	p_kindling_event->userAttributes[userAttNumber].len = 8;
	userAttNumber++;

	// end_time
	strcpy(p_kindling_event->userAttributes[userAttNumber].key, "end_time");
	memcpy(p_kindling_event->userAttributes[userAttNumber].value, &end_time, 8);
	p_kindling_event->userAttributes[userAttNumber].valueType = UINT64;
	p_kindling_event->userAttributes[userAttNumber].len = 8;
	userAttNumber++;

	// time_specs
	strcpy(p_kindling_event->userAttributes[userAttNumber].key, "type_specs");
	memcpy(p_kindling_event->userAttributes[userAttNumber].value, c_data.time_specs.data(), c_data.time_specs.length());
	p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
	p_kindling_event->userAttributes[userAttNumber].len = c_data.time_specs.length();
	userAttNumber++;

	// runq_latency
	strcpy(p_kindling_event->userAttributes[userAttNumber].key, "runq_latency");
	memcpy(p_kindling_event->userAttributes[userAttNumber].value, c_data.runq_latency.data(),
		   c_data.runq_latency.length());
	p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
	p_kindling_event->userAttributes[userAttNumber].len = c_data.runq_latency.length();
	userAttNumber++;

    // time_type
    strcpy(p_kindling_event->userAttributes[userAttNumber].key, "time_type");
    memcpy(p_kindling_event->userAttributes[userAttNumber].value, c_data.time_type.data(), c_data.time_type.length());
    p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
    p_kindling_event->userAttributes[userAttNumber].len = c_data.time_type.length();
    userAttNumber++;

    // on_stack
    auto s_tinfo = sevt->get_thread_info();
    string data = "";
    for (auto it = qls.begin(); it != qls.end(); it++) {
        KindlingInterface *plugin = qobject_cast<KindlingInterface *>(*it);
        if (plugin) {
            data.append(plugin->getCache(s_tinfo->m_tid, on_time, off_type, false, false, true));
        }

    }
    if (data != "") {
        strcpy(p_kindling_event->userAttributes[userAttNumber].key, "stack");
        memcpy(p_kindling_event->userAttributes[userAttNumber].value, data.data(), data.length());
        p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
        p_kindling_event->userAttributes[userAttNumber].len = data.length();
        userAttNumber++;
    }

    string log_msg = "";
    for (auto it = qls.begin(); it != qls.end(); it++) {
        KindlingInterface *plugin = qobject_cast<KindlingInterface *>(*it);
        if (plugin) {
            log_msg.append(plugin->getCache(s_tinfo->m_tid, on_time, off_type, false, true, false));
        }

    }
    if (log_msg != "") {
        strcpy(p_kindling_event->userAttributes[userAttNumber].key, "log");
        memcpy(p_kindling_event->userAttributes[userAttNumber].value, log_msg.data(), log_msg.length());
        p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
        p_kindling_event->userAttributes[userAttNumber].len = log_msg.length();
        userAttNumber++;
    }

    string on_info = "";
    for (auto it = qls.begin(); it != qls.end(); it++) {
        KindlingInterface *plugin = qobject_cast<KindlingInterface *>(*it);
        if (plugin) {
            on_info.append(plugin->getCache(s_tinfo->m_tid, on_time, off_type, false, false, false));
        }

    }

    strcpy(p_kindling_event->userAttributes[userAttNumber].key, "on_info");
    memcpy(p_kindling_event->userAttributes[userAttNumber].value, on_info.data(), on_info.length());
    p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
    p_kindling_event->userAttributes[userAttNumber].len = on_info.length();
    userAttNumber++;

    string info = "";
    for (auto it = qls.begin(); it != qls.end(); it++) {
        KindlingInterface *plugin = qobject_cast<KindlingInterface *>(*it);
        if (plugin) {
            info.append(plugin->getCache(s_tinfo->m_tid, off_time, off_type, true, false, false));
        }

    }
    strcpy(p_kindling_event->userAttributes[userAttNumber].key, "off_info");
    memcpy(p_kindling_event->userAttributes[userAttNumber].value, info.data(), info.length());
    p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
    p_kindling_event->userAttributes[userAttNumber].len = info.length();
    userAttNumber++;
    p_kindling_event->paramsNumber = userAttNumber;
    cout<<"aaaaaaa"<<endl;

    if(is_profile_debug && s_tinfo->m_tid == debug_tid && s_tinfo->m_pid == debug_pid){
        if (!debug_file.is_open())
        {
            debug_file.open("profile_debug_cpu.log", ios::app|ios::out);
        }else {
            for(int i= 0;i<on_time.size();i++){
                string debug_info = to_string(on_time[i].first) + "  -  " + to_string(on_time[i].second) + "on, " + " onnumber: " + to_string(i+1) +"oninfo:"+on_info;
                debug_file << debug_info<<"\n";
            }
            for(int i= 0;i<off_time.size();i++){
                string debug_info = to_string(off_time[i].first) + "  -  " + to_string(off_time[i].second) + "off, " + " ofnumber: " + to_string(i+1) +"offinfo:"+info;
                debug_file << debug_info<<"\n";
            }

        }
    }
   // printf("name: %s thread: %s(%d) userattNumber: %d\n", p_kindling_event->name, p_kindling_event->context.tinfo.comm,
  //         p_kindling_event->context.tinfo.tid, userAttNumber);

//    printf("time: %lu, %lu, %lu, %lu\n", start_time, end_time, c_data.on_total_time, c_data.off_total_time);
//    printf("user attributes: \n");
//    for (int i = 5; i < userAttNumber; i++) {
//        char* tmp;
//        memcpy(tmp, p_kindling_event->userAttributes[i].value, p_kindling_event->userAttributes[i].len);
//        printf("%s: %s\n", p_kindling_event->userAttributes[i].key, tmp);
//    }
//    printf("oninfo: %s\n", on_info.data());
//    printf("offinfo: %s\n", info.data());
//    printf("stack: %s\n", data.data());
//    printf("log: %s\n", log_msg.data());

    return 0;
}