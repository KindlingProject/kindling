#include "src/probe/publisher/publisher.h"
#include <unistd.h>
#include <sys/un.h>
#include <string>
#include <iostream>
#include <zmq.h>
#include <regex>
#include "tuples.h"
#include <dirent.h>
#include "src/probe/publisher/defination.h"
#include "src/probe/converter/sysdig_converter.h"
#include "src/probe/converter/uprobe_converter.h"

using namespace std;
using namespace kindling;

publisher::publisher(sinsp *inspector, uprobe_converter* uprobe_converter) {
    m_socket = NULL;
    m_selector = new selector(inspector);
    m_inspector = inspector;
    m_bind_address = new shared_unordered_map<string, Socket>;
    m_client_event_map = new shared_unordered_map<void *, vector<KindlingEventList *>>;
    uprobe_converter_ = uprobe_converter;
}

publisher::~publisher() {
    delete m_selector;
    delete m_bind_address;
    delete m_client_event_map;
}

void publisher::consume_sysdig_event(sinsp_evt *evt, int pid, converter *sysdigConverter) {
    if (!m_socket) {
        return;
    }

    // filter out pid in filter_pid
    for (int i : filter_pid) {
        if (i == pid) {
            return;
        }
    }
    // convert sysdig event to kindling event
    if (m_selector->select(evt->get_type(), ((sysdig_converter *) sysdigConverter)->get_kindling_category(evt))) {
        if (sysdigConverter->judge_max_size()) {
            return;
        }
        auto it = m_kindlingEventLists.find(sysdigConverter);
        KindlingEventList* kindlingEventList;
        if (it == m_kindlingEventLists.end()) {
            kindlingEventList = new KindlingEventList();
            m_kindlingEventLists[sysdigConverter] = kindlingEventList;
            m_ready[kindlingEventList] = false;
        } else {
            kindlingEventList = it->second;
        }
        sysdigConverter->convert(evt);
        // if send list was sent
        if (sysdigConverter->judge_batch_size() && !m_ready[kindlingEventList]) {
            m_kindlingEventLists[sysdigConverter] = sysdigConverter->swap_list(kindlingEventList);
            m_ready[kindlingEventList] = true;
        }
    }
}

Socket publisher::init_zeromq_rep_server() {
    void *context = zmq_ctx_new();
    void *socket = zmq_socket(context, ZMQ_REP);
    zmq_bind(socket, "ipc:///home/kindling/0");
    return socket;
}

Socket publisher::init_zeromq_push_server() {
    void *context = zmq_ctx_new();
    void *socket = zmq_socket(context, ZMQ_PUSH);
    return socket;
}

int publisher::start() {
    Socket socket = init_zeromq_rep_server();
    auto sub_server = thread(bind(&publisher::subscribe_server, this, socket));
    auto send_server = thread(bind(&publisher::send_server, this));
    sub_server.join();
    send_server.join();
    return 0;
}

void publisher::send_server(publisher *mpublisher) {
    cout << "Thread sender start" << endl;
    uint64_t total= 0;
    uint64_t msg_total_size = 0;
    while (true) {
        usleep(100000);
        for (auto list : mpublisher->m_kindlingEventLists) {
            auto pKindlingEventList = list.second;
            // flag == false
            if (!mpublisher->m_ready[pKindlingEventList]) {
                continue;
            }
            if (pKindlingEventList->kindling_event_list_size() > 0) {
                string msg;
                pKindlingEventList->SerializeToString(&msg);
                int num = pKindlingEventList->kindling_event_list_size();
                total = total + num;
                printf("Send %d kindling events, sending size: %.2f KB. Total count of kindling events: %lu\n", num, msg.length() / 1024.0, total);
//                cout << pKindlingEventList->Utf8DebugString() << endl;
                zmq_send(mpublisher->m_socket, msg.data(), msg.size(), ZMQ_DONTWAIT);
                pKindlingEventList->clear_kindling_event_list();
            }
            mpublisher->m_ready[pKindlingEventList] = false;
        }
    }
}

void publisher::subscribe_server(publisher *mpublisher, Socket subscribe_socket) {
    cout << "Subcribe server start" << endl;
    while (true) {
        char result[1000];
        memset(result, 0, 1000);
        zmq_recv(subscribe_socket, result, 1000, 0);
        string reason;
        mpublisher->subscribe(result, reason);
        zmq_send(subscribe_socket, reason.data(), 7, 0);
    }
}

void publisher::subscribe(string sub_event, string &reason) {
    SubEvent subEvent;
    void *socket;

    subEvent.ParseFromString(sub_event);
    cout << "subscribe info: " << subEvent.Utf8DebugString() << endl;
    string address = subEvent.address().data();

    // filter out subscriber
    pid_mutex_.lock();
    filter_pid.push_back(((int) subEvent.pid()));
    pid_mutex_.unlock();

    // subscribe
    auto ad_index = m_bind_address->find((char *) address.data());
    // if exists, delete first
    if (ad_index != m_bind_address->end()) {
        // TODO
    }
    // new socket and bind
    socket = init_zeromq_push_server();
    int rc = zmq_bind(socket, address.c_str());
    if (rc != 0) {
        reason = "sub address error";
        return;
    }
    // set selectors
    m_selector->parse(subEvent.labels());

    // bind
    m_socket = socket;
    m_bind_address->insert((char *) address.data(), socket);
}

// 将uprobe事件放入待发送map
px::Status publisher::consume_uprobe_data(uint64_t table_id, px::types::TabletID tablet_id,
                  std::unique_ptr<px::types::ColumnWrapperRecordBatch> record_batch) {

    if (record_batch->empty() || record_batch->at(0)->Size() == 0) {
        return px::Status::OK();
    }

    if (record_batch->size() != 22) {
        return px::Status::OK();
    }

    auto it = m_kindlingEventLists.find(uprobe_converter_);
    KindlingEventList* kindlingEventList;
    if (it == m_kindlingEventLists.end()) {
        kindlingEventList = new KindlingEventList();
        m_kindlingEventLists[uprobe_converter_] = kindlingEventList;
    } else {
        kindlingEventList = it->second;
    }

    auto batch_size = record_batch->at(0)->Size();
    for (size_t i = 0; i < batch_size; ++ i ) {
        int64_t ts = record_batch->at(px::stirling::kHTTPTimeIdx)->Get<px::types::Time64NSValue>(i).val;
        int32_t pid = record_batch->at(px::stirling::kHTTPUPIDIdx)->Get<px::types::UInt128Value>(i).High64();
        std::string remote_addr = record_batch->at(px::stirling::kHTTPRemoteAddrIdx)->Get<px::types::StringValue>(i);
        int64_t remote_port = record_batch->at(px::stirling::kHTTPRemotePortIdx)->Get<px::types::Int64Value>(i).val;
	std::string source_addr = record_batch->at(px::stirling::kHTTPSourceAddrIdx)->Get<px::types::StringValue>(i);
        int64_t source_port = record_batch->at(px::stirling::kHTTPSourcePortIdx)->Get<px::types::Int64Value>(i).val;
        int64_t trace_role = record_batch->at(px::stirling::kHTTPTraceRoleIdx)->Get<px::types::Int64Value>(i).val;
        int64_t major_version = record_batch->at(px::stirling::kHTTPMajorVersionIdx)->Get<px::types::Int64Value>(i).val;
        int64_t minor_version = record_batch->at(px::stirling::kHTTPMinorVersionIdx)->Get<px::types::Int64Value>(i).val;
        int64_t content_type = record_batch->at(px::stirling::kHTTPContentTypeIdx)->Get<px::types::Int64Value>(i).val;
        std::string req_headers = record_batch->at(px::stirling::kHTTPReqHeadersIdx)->Get<px::types::StringValue>(i);
        std::string req_method = record_batch->at(px::stirling::kHTTPReqMethodIdx)->Get<px::types::StringValue>(i);
        std::string req_path = record_batch->at(px::stirling::kHTTPReqPathIdx)->Get<px::types::StringValue>(i);
        std::string req_body = record_batch->at(px::stirling::kHTTPReqBodyIdx)->Get<px::types::StringValue>(i);
        int64_t req_body_size = record_batch->at(px::stirling::kHTTPReqBodySizeIdx)->Get<px::types::Int64Value>(i).val;
        std::string resp_headers = record_batch->at(px::stirling::kHTTPRespHeadersIdx)->Get<px::types::StringValue>(i);
        int64_t resp_status = record_batch->at(px::stirling::kHTTPRespStatusIdx)->Get<px::types::Int64Value>(i).val;
        std::string resp_body = record_batch->at(px::stirling::kHTTPRespBodyIdx)->Get<px::types::StringValue>(i);
        int64_t resp_body_size = record_batch->at(px::stirling::kHTTPRespBodySizeIdx)->Get<px::types::Int64Value>(i).val;
        int64_t latency = record_batch->at(px::stirling::kHTTPLatencyIdx)->Get<px::types::Int64Value>(i).val;
	
	VLOG(1) << absl::Substitute("[stirling][grpc] ts:$0 pid:$1 remote_addr:$2 remote_port:$3 trace_role:$4 source_addr:$5 req_method:$6 req_path:$7 latency:$8 source_port:$9",
                                    ts, pid, remote_addr, remote_port, trace_role, source_addr, req_method, req_path, latency, source_port);

        struct grpc_event_t gevt;
        gevt.timestamp = ts;
        gevt.pid = pid;
        gevt.remote_addr = remote_addr;
        gevt.remote_port = remote_port;
	gevt.source_addr = source_addr;
        gevt.source_port = source_port;
        gevt.trace_role = trace_role;
        gevt.req_headers = req_headers;
        gevt.req_method = req_method;
        gevt.req_path = req_path;
        gevt.req_body = req_body;
        gevt.req_body_size = req_body_size;
        gevt.resp_status = resp_status;
        gevt.resp_body = resp_body;
        gevt.resp_body_size = resp_body_size;
        gevt.latency = latency;

        auto tinfo = m_inspector->get_thread_ref(pid, true, true, true);
        if (tinfo) {
            gevt.container_id = tinfo->m_container_id;
        } else {
		    VLOG(1) << "[stirling] cannot find container_id for pid:" << pid << std::endl;
        }

        if (uprobe_converter_->judge_max_size() == false) {
            // convert to kindling event
            uprobe_converter_->convert(&gevt);
            // if send list was sent
            if (uprobe_converter_->judge_batch_size() && !m_ready[kindlingEventList]) {
                m_kindlingEventLists[uprobe_converter_] = uprobe_converter_->swap_list(kindlingEventList);
                m_ready[kindlingEventList] = true;
            }
        }
    }
    return px::Status::OK();
}

selector::selector(sinsp *inspector) {
    m_labels = new map<ppm_event_type, vector<Category>* >;
    for (auto e : kindling_to_sysdig) {
        m_events[e.event_name] = e.event_type;
    }
    for (auto c : category_map) {
        m_categories[c.cateogry_name] = c.category_value;
    }
    m_inspector = inspector;
}

bool selector::select(uint16_t type, Category category) {
    auto it = m_labels->find(static_cast<const ppm_event_type>(type));
    if (it != m_labels->end()) {
        if (it->second->size() == 0) {
            return true;
        }
        for (auto c : *it->second) {
            if (c == category) {
                return true;
            }
        }
    }
    return false;
}
Category selector::get_category(string category) {
    auto it = m_categories.find(category);
    if (it != m_categories.end()) {
        return it->second;
    } else {
         return CAT_NONE;
    }
}

void selector::parse(const google::protobuf::RepeatedPtrField<::kindling::Label> &labels) {
    for (auto label : labels) {
        auto it = m_events.find(label.name());
        if (it != m_events.end()) {
            auto v = m_labels->find(it->second);
            auto c = get_category(label.category());
            if (label.category() != "" && c == CAT_NONE) {
                cout << "Subscribe: Kindling event category err: " << label.category() << endl;
                continue;
            }
            if (v != m_labels->end()) {
               v->second->push_back(c);
            } else {
                auto categories = new vector<Category>;
                if (c != CAT_NONE) {
                    categories->push_back(c);
                }
                m_labels->insert(pair<ppm_event_type, vector<Category> *> (it->second, categories));
            }
            cout << "Subscribe info: type: " << it->second << " category: " << (label.category() != "" ? label.category() : "none") << endl;
        } else {
            cout << "Subscribe: Kindling event name err: " << label.name() << endl;
        }
    }
    // notify kernel, set eventmask
    for (auto it : *m_labels) {
        m_inspector->set_eventmask(it.first);
        if (!PPME_IS_ENTER(it.first)) {
            m_inspector->set_eventmask(it.first - 1);
        }
    }
}
