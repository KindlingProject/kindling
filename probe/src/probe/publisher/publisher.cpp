#include "src/probe/publisher/publisher.h"
#include <unistd.h>
#include <sys/un.h>
#include <string>
#include <cstring>
#include <iostream>
#include <zmq.h>
#include <regex>
#include "tuples.h"
#include <dirent.h>
#include "src/probe/publisher/defination.h"
#include "src/probe/converter/sysdig_converter.h"

using namespace std;
using namespace kindling;

publisher::publisher(sinsp *inspector) {
    m_socket = NULL;
    m_selector = new selector(inspector);
    m_inspector = inspector;
    m_bind_address = new shared_unordered_map<string, Socket>;
    m_client_event_map = new shared_unordered_map<void *, vector<KindlingEventList *>>;
}

publisher::~publisher() {
    delete m_selector;
    delete m_bind_address;
    delete m_client_event_map;
}

bool filterSwitch(char *val, int threshold){
    int num = atoi(val);
    return num <= threshold;
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

        if(evt->get_type() == PPME_SCHEDSWITCH_6_E){
            if(filterSwitch(evt->get_param(1)->m_val, 0)){ //filter major
                return;
            }
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

        if (sysdigConverter->judge_max_size()) {
            // check if the send list has sent
            if (m_ready[kindlingEventList]) {
                // drop event
                return;
            }
            swap_list(sysdigConverter, kindlingEventList);
        }

        sysdigConverter->convert(evt);
        // if send list was sent
        if (sysdigConverter->judge_batch_size() && !m_ready[kindlingEventList]) {
            swap_list(sysdigConverter, kindlingEventList);
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
    LOG(INFO) << "Thread sender start";
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
                LOG(INFO) << "Send " << num << " kindling events, sending size: " << setprecision(2) <<
                    msg.length() / 1024.0 <<" KB. Total count of kindling events: " << total;
//                cout << pKindlingEventList->Utf8DebugString();
                zmq_send(mpublisher->m_socket, msg.data(), msg.size(), ZMQ_DONTWAIT);
                pKindlingEventList->clear_kindling_event_list();
            }
            mpublisher->m_ready[pKindlingEventList] = false;
        }
    }
}

void publisher::subscribe_server(publisher *mpublisher, Socket subscribe_socket) {
    LOG(INFO) << "Subcribe server start";
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
    LOG(INFO) << "subscribe info: " << subEvent.Utf8DebugString();
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

void publisher::swap_list(converter *cvt, KindlingEventList* kindlingEventList) {
    kindlingEventList = cvt->swap_list(kindlingEventList);
    m_kindlingEventLists[cvt] = kindlingEventList;
    m_ready[kindlingEventList] = true;
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
                LOG(INFO) << "Subscribe: Kindling event category err: " << label.category();
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
            LOG(INFO) << "Subscribe info: type: " << it->second << " category: " << (label.category() != "" ? label.category() : "none");
        } else {
            LOG(INFO) << "Subscribe: Kindling event name err: " << label.name();
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
