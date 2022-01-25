//
// Created by 散养鸡 on 2021/12/13.
//

#ifndef HCMINE_PUBLISHER_H
#define HCMINE_PUBLISHER_H
#include <vector>
#include "src/probe/converter/kindling_event.pb.h"
#include "src/probe/publisher/subscribe.pb.h"
#include "src/probe/converter/converter.h"
#include "src/probe/utils/shared_unordered_map.h"
#include "src/userspace/libsinsp/sinsp.h"
#include "src/probe/converter/uprobe_converter.h"

#include <src/common/base/base.h>
#include <src/shared/types/column_wrapper.h>
#include <src/shared/types/hash_utils.h>
#include <src/stirling/source_connectors/socket_tracer/http_table.h>

using namespace std;
using namespace kindling;
typedef void * Socket;

class selector {
public:
    selector();
    bool select(uint16_t type, Category category);
    void parse(const google::protobuf::RepeatedPtrField<::kindling::Label> &field);

private:
    map<ppm_event_type, vector<Category>*> *m_labels;
    map<string, ppm_event_type> m_events;
    map<string, Category> m_categories;
    Category get_category(string category);
};

// publish kindling event
class publisher {
public:
    publisher(sinsp *, uprobe_converter*);
    ~publisher();
    // return list, add for converter or clear for send
    vector<KindlingEventList*> get_kindlingEventLists(converter *cvter);

    void distribute_event(sinsp_evt *evt, int pid, converter *sysdigConverter);
    px::Status consume_uprobe_data(uint64_t table_id, px::types::TabletID tablet_id,
            std::unique_ptr<px::types::ColumnWrapperRecordBatch> record_batch);
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

    shared_unordered_map<string, Socket> *m_bind_address;

    // vector for multi event source, e.g. [0] for sysdig, [1] for pixie
    shared_unordered_map<Socket, vector<KindlingEventList *>> *m_client_event_map;
    // selectors
    selector *m_selector;

    // used for sysdig
    void put_uds_map(converter *converter, sinsp_evt* evt);
    void put_sub_event_map(uint16_t sub_event, void *socket);
    void delete_sub_event_map(uint16_t sub_event, void *socket);
    void set_sysdig_event(uint16_t type);

    uprobe_converter* uprobe_converter_;
    sinsp *m_inspector;
    std::mutex event_mutex_;
    std::mutex pid_mutex_;
    vector<int> filter_pid;
};

#endif //HCMINE_PUBLISHER_H
