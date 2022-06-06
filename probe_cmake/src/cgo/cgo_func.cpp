#include <iostream>
#include <cstdlib>
#include "cgo_func.h"
#include "scap_open_exception.h"
#include "sinsp_capture_interrupt_exception.h"
#include "util/util.h"

static sinsp *inspector = nullptr;

int cnt = 0;

void init_probe(){
	bool bpf = false;
	string bpf_probe;
	inspector = new sinsp();
	string output_format;
	output_format = "*%evt.num %evt.outputtime %evt.cpu %container.name (%container.id) %proc.name (%thread.tid:%thread.vtid) %evt.dir %evt.type %evt.info";
	cout<<"1"<<endl;
	try {
		inspector = new sinsp();
		inspector->set_hostname_and_port_resolution_mode(false);
		inspector->set_snaplen(80);

		inspector->suppress_events_comm("containerd");
		inspector->suppress_events_comm("dockerd");
		inspector->suppress_events_comm("containerd-shim");
		inspector->suppress_events_comm("kindling-collector");
		inspector->suppress_events_comm("sshd");
		sinsp_evt_formatter formatter(inspector, output_format);
		const char *probe = scap_get_bpf_probe_from_env();
		if (probe) {
			bpf = true;
			bpf_probe = probe;
		}

		bool open_success = true;

		try {
			inspector->open("");
			inspector->clear_eventmask();
			inspector->set_eventmask(PPME_SYSCALL_WRITEV_X);
			inspector->set_eventmask(PPME_SYSCALL_WRITEV_X-1);
			inspector->set_eventmask(PPME_SYSCALL_WRITE_X);
			inspector->set_eventmask(PPME_SYSCALL_WRITE_E);
			inspector->set_eventmask(PPME_SYSCALL_READ_X);
			inspector->set_eventmask(PPME_SYSCALL_READ_E);
		}
		catch (const sinsp_exception &e) {
			open_success = false;
			cout<<"open failed"<<endl;
		}

		//
		// Starting the live capture failed, try to load the driver with
		// modprobe.
		//
		cout<<"open"<<endl;
//		thread catch_signal(sigsetup);

		//TerminationHandler::set_sinsp(inspector);
		//do_inspect();
		sleep(100000000000000);
		cout<<"not close"<<endl;
		inspector->close();
	}
	catch (const exception &e) {
		fprintf(stderr, "kindling probe init err: %s", e.what());
	}
	//delete inspector;
}


int getEvent(void **pp_kindling_event){
	int32_t res;
	sinsp_evt *ev;
	string line;
	string output_format;
	output_format = "*%evt.num %evt.outputtime %evt.cpu %container.name (%container.id) %proc.name (%thread.tid:%thread.vtid) %evt.dir %evt.type %evt.info";
	//sinsp_evt_formatter formatter(inspector, output_format);
	res = inspector->next(&ev);
	if (res == SCAP_TIMEOUT) {
		return 0;
	} else if (res != SCAP_SUCCESS) {
		return 0;
	}
	if (!inspector->is_debug_enabled() &&
	    ev->get_category() & EC_INTERNAL) {
		return 0;
	}
	auto threadInfo = ev->get_thread_info();
	if (threadInfo == nullptr) {
		return 0;
	}
	auto category = ev->get_category();
	if (category & EC_IO_BASE) {
		auto pres = ev->get_param_value_raw("res");
		if (pres && *(int64_t *) pres->m_val <= 0) {
			return 0;
		}
	}
	kindling_event_t_for_go *p_kindling_event;
	if (nullptr == *pp_kindling_event) {
		*pp_kindling_event = (kindling_event_t_for_go *) malloc(sizeof(kindling_event_t_for_go));
		p_kindling_event = (kindling_event_t_for_go *) *pp_kindling_event;

		p_kindling_event->name = (char *) malloc(sizeof(char) * 1024);
		p_kindling_event->tinfo.comm = (char *) malloc(sizeof(char) * 256);
		p_kindling_event->tinfo.containerId = (char *) malloc(sizeof(char) * 256);
		p_kindling_event->fdInfo.filename = (char *) malloc(sizeof(char) * 1024);
		p_kindling_event->fdInfo.directory = (char *) malloc(sizeof(char) * 1024);

		for(int i=0;i<8;i++){
			p_kindling_event->userAttributes[i].key = (char *) malloc(sizeof(char) * 128);
			p_kindling_event->userAttributes[i].value = (char*) malloc(sizeof(char) * 1024);
		}
	}
	p_kindling_event = (kindling_event_t_for_go *) *pp_kindling_event;

	sinsp_fdinfo_t* fdInfo = ev->get_fd_info();
	p_kindling_event->timestamp = ev->get_ts();
	p_kindling_event->tinfo.pid = threadInfo->m_pid;
	p_kindling_event->tinfo.tid = threadInfo->m_tid;
	p_kindling_event->tinfo.uid = threadInfo->m_uid;
	p_kindling_event->tinfo.gid = threadInfo->m_gid;
	p_kindling_event->fdInfo.num = ev->get_fd_num();
	if(nullptr != fdInfo)
	{
		p_kindling_event->fdInfo.fdType = fdInfo->m_type;

		switch(fdInfo->m_type)
		{
		case SCAP_FD_FILE:
		case SCAP_FD_FILE_V2:
		{

			string name = fdInfo->m_name;
			size_t pos = name.rfind('/');
			if(pos != string::npos)
			{
				if(pos < name.size() - 1)
				{
					string fileName = name.substr(pos + 1, string::npos);
					memcpy(p_kindling_event->fdInfo.filename, fileName.data(), fileName.length());
					if(pos != 0)
					{

						name.resize(pos);

						kindling_strcpy(p_kindling_event->fdInfo.directory, (char *)name.data(), (int)strlen((char *)name.data()));
					}
					else
					{
						kindling_strcpy(p_kindling_event->fdInfo.directory, "/", (int)strlen("/"));
					}
				}
			}
			break;
		}
		case SCAP_FD_IPV4_SOCK:
		case SCAP_FD_IPV4_SERVSOCK:
			p_kindling_event->fdInfo.protocol = fdInfo->get_l4proto();
			p_kindling_event->fdInfo.role = fdInfo->is_role_server();
			p_kindling_event->fdInfo.sip = fdInfo->m_sockinfo.m_ipv4info.m_fields.m_sip;
			p_kindling_event->fdInfo.dip = fdInfo->m_sockinfo.m_ipv4info.m_fields.m_dip;
			p_kindling_event->fdInfo.sport = fdInfo->m_sockinfo.m_ipv4info.m_fields.m_sport;
			p_kindling_event->fdInfo.dport = fdInfo->m_sockinfo.m_ipv4info.m_fields.m_dport;
			break;
		case SCAP_FD_UNIX_SOCK:
			p_kindling_event->fdInfo.source = fdInfo->m_sockinfo.m_unixinfo.m_fields.m_source;
			p_kindling_event->fdInfo.destination = fdInfo->m_sockinfo.m_unixinfo.m_fields.m_dest;
			break;
		default:
			break;
		}
	}

	switch (ev->get_type()) {
	case PPME_TCP_RCV_ESTABLISHED_E:

	case PPME_TCP_SEND_RESET_E:
	case PPME_TCP_RECEIVE_RESET_E: {
		auto pTuple = ev->get_param_value_raw("tuple");
		break;
	}
	default:
		for (auto i = 0; i < ev->get_num_params(); i++) {
			if(i>=8){
				break;
			}
			kindling_strcpy(p_kindling_event->userAttributes[i].key, (char*)ev->get_param_name(i), (int)strlen((char*)ev->get_param_name(i)));
			memcpy(p_kindling_event->userAttributes[i].value, ev->get_param(i)->m_val, ev->get_param(i)->m_len);
			p_kindling_event->userAttributes[i].valueType = ev->get_param_info(i)->type;
		}
	}

	kindling_strcpy(p_kindling_event->name, (char*)ev->get_name(), (int)strlen((char*)ev->get_name()));
	kindling_strcpy(p_kindling_event->tinfo.comm, (char*)threadInfo->m_comm.data(), (int)strlen((char*)threadInfo->m_comm.data()));;
	kindling_strcpy(p_kindling_event->tinfo.containerId, (char*)threadInfo->m_container_id.data(), 12);
	return 1;

}