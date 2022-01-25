//
// Created by 散养鸡 on 2021/12/13.
//

#include "src/probe/publisher/publisher.h"
#include <unistd.h>
#include <sys/un.h>
#include <string>
#include <iostream>
#include <zmq.h>
#include <regex>
#include "src/userspace/libsinsp/tuples.h"
#include <dirent.h>
#include "src/probe/publisher/defination.h"
#include "src/probe/converter/sysdig_converter.h"
#include "src/probe/converter/uprobe_converter.h"

using namespace std;
using namespace kindling;
vector<Socket> *sub_event_list[PPM_EVENT_MAX];

publisher::publisher(sinsp *inspector, uprobe_converter* uprobe_converter) {
    m_socket = NULL;
    m_selector = new selector();
    m_inspector = inspector;
    m_bind_address = new shared_unordered_map<string, Socket>;
    m_client_event_map = new shared_unordered_map<void *, vector<KindlingEventList *>>;
    uprobe_converter_ = uprobe_converter;
}

publisher::~publisher() {
    delete m_bind_address;
    delete m_client_event_map;
}

void publisher::distribute_event(sinsp_evt *evt, int pid, converter *sysdigConverter) {
    if (!m_socket) {
        return;
    }

    sinsp_threadinfo *evThreadInfo = evt->get_thread_info();
    if (evThreadInfo->m_ptid == pid || evThreadInfo->m_pid == pid || evThreadInfo->m_pid == 0) {
        return;
    }

    for (int i : filter_pid) {
        if (i == evThreadInfo->m_pid) {
            return;
        }
    }

    bool f = m_selector->select(evt->get_type(), ((sysdig_converter *) sysdigConverter)->get_kindling_category(evt));
    if (f) {
        put_uds_map(sysdigConverter, evt);
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
    while (true) {
        usleep(100000);
        // 序列化
        mpublisher->event_mutex_.lock();
//        auto client_event = mpublisher->m_client_event_map->begin();
//        while (client_event != mpublisher->m_client_event_map->end()) {
//            string msg;
//            // TODO
//            auto pKindlingEventList = client_event->second[0];
//            auto client_socket = client_event->first;
//            pKindlingEventList->SerializeToString(&msg);
////            cout << pKindlingEventList->Utf8DebugString() << endl;
//            if (0 == pKindlingEventList->kindling_event_list_size()) {
//                client_event++;
//                continue;
//            }
//            zmq_send(client_socket, msg.data(), msg.size(), ZMQ_DONTWAIT);
//            pKindlingEventList->clear_kindling_event_list();
//            client_event++;
//        }

        for (auto list : mpublisher->m_kindlingEventLists) {
            auto pKindlingEventList = list.second;
            if (pKindlingEventList->kindling_event_list_size() > 0) {
                string msg;
                pKindlingEventList->SerializeToString(&msg);
//                cout << pKindlingEventList->Utf8DebugString() << endl;
                zmq_send(mpublisher->m_socket, msg.data(), msg.size(), ZMQ_DONTWAIT);
                pKindlingEventList->clear_kindling_event_list();
            }
        }
        mpublisher->event_mutex_.unlock();
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

//    // notify kernel probe and set sub_event_map
//    for (int i = 0; i < subEvent.event_type_size(); i++) {
//        uint16_t event_type = subEvent.event_type(i);
//        set_sysdig_event(event_type);
//        put_sub_event_map(event_type, socket);
//    }
}

// 将uprobe事件放入待发送map
px::Status publisher::consume_uprobe_data(uint64_t table_id, px::types::TabletID tablet_id,
                  std::unique_ptr<px::types::ColumnWrapperRecordBatch> record_batch) {

    // std::cout << "[qianlu][grpc] table_id:" << table_id << " tablet_id:" << tablet_id << std::endl;

    if (record_batch->empty() || record_batch->at(0)->Size() == 0) {
        // std::cout << "[qianlu][grpc] record_batch is empty. table_id:" << table_id << std::endl;
        return px::Status::OK();
    }

    if (record_batch->size() != 20) {
//        std::cout << "[qianlu] size not match" << " table_id:" << table_id << " tablet_id:" << tablet_id << "column size:" << record_batch->size() << std::endl;
        return px::Status::OK();
    }

    // std::cout << "[qianlu][grpc] record_batch cols is " << record_batch->size() << " samples:" << record_batch->at(0)->Size() << std::endl;

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
        // std::cout << "[qianlu][grpc] begin to process record " << i << " ... " << std::endl;
        int64_t ts = record_batch->at(px::stirling::kHTTPTimeIdx)->Get<px::types::Time64NSValue>(i).val;
        // std::cout << "ts:" << ts << " idx:" << px::stirling::kHTTPTimeIdx << std::endl;
        int32_t pid = record_batch->at(px::stirling::kHTTPUPIDIdx)->Get<px::types::UInt128Value>(i).High64();
        // std::cout << "pid:" << pid << " idx:" << px::stirling::kHTTPUPIDIdx << std::endl;
        std::string remote_addr = record_batch->at(px::stirling::kHTTPRemoteAddrIdx)->Get<px::types::StringValue>(i);
        // std::cout << "remote_addr:" << remote_addr << " idx:" << px::stirling::kHTTPRemoteAddrIdx << std::endl;
        int64_t remote_port = record_batch->at(px::stirling::kHTTPRemotePortIdx)->Get<px::types::Int64Value>(i).val;
        // std::cout << "remote_port:" << remote_port << " idx:" << px::stirling::kHTTPRemotePortIdx << std::endl;
        int64_t trace_role = record_batch->at(px::stirling::kHTTPTraceRoleIdx)->Get<px::types::Int64Value>(i).val;
        // std::cout << "trace_role:" << trace_role << " idx:" << px::stirling::kHTTPTraceRoleIdx << std::endl;
        int64_t major_version = record_batch->at(px::stirling::kHTTPMajorVersionIdx)->Get<px::types::Int64Value>(i).val;
        // std::cout << "major_version:" << major_version << " idx:" << px::stirling::kHTTPMajorVersionIdx << std::endl;
        int64_t minor_version = record_batch->at(px::stirling::kHTTPMinorVersionIdx)->Get<px::types::Int64Value>(i).val;
        // std::cout << "minor_version:" << minor_version << " idx:" << px::stirling::kHTTPMinorVersionIdx << std::endl;
        int64_t content_type = record_batch->at(px::stirling::kHTTPContentTypeIdx)->Get<px::types::Int64Value>(i).val;
        // std::cout << "content_type:" << content_type << " idx:" << px::stirling::kHTTPContentTypeIdx << std::endl;
        std::string req_headers = record_batch->at(px::stirling::kHTTPReqHeadersIdx)->Get<px::types::StringValue>(i);
        // std::cout << "req_headers:" << req_headers << " idx:" << px::stirling::kHTTPReqHeadersIdx << std::endl;
        std::string req_method = record_batch->at(px::stirling::kHTTPReqMethodIdx)->Get<px::types::StringValue>(i);
        // std::cout << "req_method:" << req_method << " idx:" << px::stirling::kHTTPReqMethodIdx << std::endl;
        std::string req_path = record_batch->at(px::stirling::kHTTPReqPathIdx)->Get<px::types::StringValue>(i);
        // std::cout << "req_path:" << req_path << std::endl;
        std::string req_body = record_batch->at(px::stirling::kHTTPReqBodyIdx)->Get<px::types::StringValue>(i);
        // std::cout << "req_body:" << req_body << std::endl;
        int64_t req_body_size = record_batch->at(px::stirling::kHTTPReqBodySizeIdx)->Get<px::types::Int64Value>(i).val;
        // std::cout << "req_body_size:" << req_body_size << std::endl;
        std::string resp_headers = record_batch->at(px::stirling::kHTTPRespHeadersIdx)->Get<px::types::StringValue>(i);
        // std::cout << "resp_headers:" << resp_headers << std::endl;
        int64_t resp_status = record_batch->at(px::stirling::kHTTPRespStatusIdx)->Get<px::types::Int64Value>(i).val;
        // std::cout << "resp_status:" << resp_status << std::endl;
        std::string resp_body = record_batch->at(px::stirling::kHTTPRespBodyIdx)->Get<px::types::StringValue>(i);
        // std::cout << "resp_body:" << resp_body << std::endl;
        int64_t resp_body_size = record_batch->at(px::stirling::kHTTPRespBodySizeIdx)->Get<px::types::Int64Value>(i).val;
        // std::cout << "resp_body_size:" << resp_body_size << std::endl;
        int64_t latency = record_batch->at(px::stirling::kHTTPLatencyIdx)->Get<px::types::Int64Value>(i).val;
        // std::cout << "latency:" << latency << std::endl;

        struct grpc_event_t gevt;
        gevt.timestamp = ts;
        gevt.pid = pid;
        gevt.remote_addr = remote_addr;
        gevt.remote_port = remote_port;
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
            // std::cout << "[qianlu] find container_id for pid:" << pid << " container_id:" << gevt.container_id << std::endl;
        } else {
            std::cout << "[qianlu] cannot find container_id for pid:" << pid << std::endl;
        }

        // convert数据
        event_mutex_.lock();
        KindlingEvent *kindlingEvent = kindlingEventList->add_kindling_event_list();
        uprobe_converter_->convert(kindlingEvent, &gevt);
        event_mutex_.unlock();
    }
    return px::Status::OK();
}

// 将事件放入待发送map
void publisher::put_uds_map(converter *sysdigConverter, sinsp_evt *evt) {
    auto it = m_kindlingEventLists.find(sysdigConverter);
    KindlingEventList* kindlingEventList;
    if (it == m_kindlingEventLists.end()) {
        kindlingEventList = new KindlingEventList();
        m_kindlingEventLists[sysdigConverter] = kindlingEventList;
    } else {
        kindlingEventList = it->second;
    }
    event_mutex_.lock();
    KindlingEvent *kindlingEvent = kindlingEventList->add_kindling_event_list();
    sysdigConverter->convert(kindlingEvent, evt);
    event_mutex_.unlock();
}

void publisher::put_sub_event_map(uint16_t sub_event, void *socket) {
    event_mutex_.lock();
    vector<void *> *client_sockets = sub_event_list[sub_event];
    if (client_sockets != nullptr) {
        bool exist_client = false;
        for (void *client : *client_sockets) {
            if (client == socket) {
                exist_client = true;
                break;
            }
        }
        if (!exist_client) {
            client_sockets->push_back(socket);
            cout << system("date") << " event未被订阅，插入map,fd: " << socket << "event: " << sub_event << endl;
        }
    } else {
        auto *clients = new vector<void *>;
        clients->push_back(socket);
        cout << system("date") << " event未被订阅，插入map,fd: " << socket << "event: " << sub_event << endl;
        sub_event_list[sub_event] = clients;
    }

    event_mutex_.unlock();
}

void publisher::delete_sub_event_map(uint16_t sub_event, void *socket) {
    event_mutex_.lock();
    vector<Socket> *client_sockets = sub_event_list[sub_event];
    if (client_sockets != nullptr) {
        vector<void *>::iterator it;
        for (it = client_sockets->begin(); it != client_sockets->end();) {
            if (*it == socket) {
                it = client_sockets->erase(it);
            } else {
                it++;
            }
        }
    }
    event_mutex_.unlock();
}

void publisher::set_sysdig_event(uint16_t type) {

}

selector::selector() {
    m_labels = new map<ppm_event_type, vector<Category>* >;
    for (auto e : kindling_to_sysdig) {
        m_events[e.event_name] = e.event_type;
    }
    for (auto c : category_map) {
        m_categories[c.cateogry_name] = c.category_value;
    }
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
}
