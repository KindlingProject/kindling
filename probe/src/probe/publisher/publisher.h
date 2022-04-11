#ifndef KINDLING_PROBE_PUBLISHER_H
#define KINDLING_PROBE_PUBLISHER_H
#include <vector>
#include "src/probe/converter/kindling_event.pb.h"
#include "src/probe/publisher/subscribe.pb.h"
#include "src/probe/converter/converter.h"
#include "src/probe/utils/shared_unordered_map.h"
#include "sinsp.h"

#include <src/common/base/base.h>
#include <src/shared/types/column_wrapper.h>
#include <src/shared/types/hash_utils.h>

using namespace std;
using namespace kindling;
typedef void * Socket;

class selector {
public:
    selector(sinsp *inspector);
    bool select(uint16_t type, Category category);
    void parse(const google::protobuf::RepeatedPtrField<::kindling::Label> &field);

private:
    map<ppm_event_type, vector<Category>*> *m_labels;
    map<string, ppm_event_type> m_events;
    map<string, Category> m_categories;
    Category get_category(string category);
    sinsp *m_inspector;
};

// publish kindling event
class publisher {
public:
    publisher(sinsp *);
    ~publisher();
    // return list, add for converter or clear for send
    vector<KindlingEventList*> get_kindlingEventLists(converter *cvter);

    void consume_sysdig_event(sinsp_evt *evt, int pid, converter *sysdigConverter);
    // run [thread] send, [thread] subscribe
    int start();
private:
    Socket init_zeromq_rep_server();
    Socket init_zeromq_push_server();
    // send threads: different content-keys correspond to sender thread.
    static void send_server(publisher *);
    // subscribe thread: accept subscribe request from clients
    static void subscribe_server(publisher *, Socket);

    void subscribe(string sub_event, string &reason);
    void unsubscribe(string sub_event);

    // single sender
    Socket m_socket;
    map<converter *, KindlingEventList *> m_kindlingEventLists;
    map<KindlingEventList *, bool> m_ready;

    shared_unordered_map<string, Socket> *m_bind_address;

    // multi sender
    // vector for multi event source, e.g. [0] for sysdig, [1] for pixie
    shared_unordered_map<Socket, vector<KindlingEventList *>> *m_client_event_map;
    // selectors
    selector *m_selector;


    sinsp *m_inspector;
    std::mutex pid_mutex_;
    vector<int> filter_pid;
};

#endif //KINDLING_PROBE_PUBLISHER_H
