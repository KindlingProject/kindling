//
// Created by 千陆 on 2021/12/30.
//

#include "src/probe/converter/uprobe_converter.h"

uprobe_converter::uprobe_converter() : converter(100, INT_MAX) {}
uprobe_converter::uprobe_converter(int batch_size, int max_size) : converter(batch_size, max_size) {}
uprobe_converter::~uprobe_converter() {}

void uprobe_converter::convert(void *evt) {
//    std::unique_ptr<px::types::ColumnWrapperRecordBatch> record_batch =
//            static_cast<std::unique_ptr<px::types::ColumnWrapperRecordBatch>>(evt);
    auto kevt = get_kindlingEventList()->add_kindling_event_list();
    struct grpc_event_t* gevt = static_cast<struct grpc_event_t*>(evt);

    VLOG(1) << absl::Substitute("[grpc] entering uprobe converter ... ");

    kevt->set_source(UPROBE);
    kevt->set_name("grpc_uprobe");
    kevt->set_category(CAT_NET);
    kevt->set_timestamp(gevt->timestamp);

    // pid
    auto pid_attr = kevt->add_user_attributes();
    pid_attr->set_key("pid");
    pid_attr->set_value_type(INT32);
    pid_attr->set_value(&gevt->pid, 4);

    // remote_addr
    auto ra_attr = kevt->add_user_attributes();
    ra_attr->set_key("remote_addr");
    ra_attr->set_value_type(CHARBUF);
    ra_attr->set_value(gevt->remote_addr);

    // remote_port
    auto rp_attr = kevt->add_user_attributes();
    rp_attr->set_key("remote_port");
    rp_attr->set_value_type(INT64);
    rp_attr->set_value(&gevt->remote_port, 8);

    // source_addr
    auto sa_attr = kevt->add_user_attributes();
    sa_attr->set_key("source_addr");
    sa_attr->set_value_type(CHARBUF);
    sa_attr->set_value(gevt->source_addr);

    // source_port
    auto sp_attr = kevt->add_user_attributes();
    sp_attr->set_key("source_port");
    sp_attr->set_value_type(INT64);
    sp_attr->set_value(&gevt->source_port, 8);

    // trace_role
    auto tc_attr = kevt->add_user_attributes();
    tc_attr->set_key("trace_role");
    tc_attr->set_value_type(INT64);
    tc_attr->set_value(&gevt->trace_role, 8);

    // req_headers
    auto rh_attr = kevt->add_user_attributes();
    rh_attr->set_key("req_headers");
    rh_attr->set_value_type(CHARBUF);
    rh_attr->set_value(gevt->req_headers);

    // req_method
    auto req_method_attr = kevt->add_user_attributes();
    req_method_attr->set_key("req_method");
    req_method_attr->set_value_type(CHARBUF);
    req_method_attr->set_value(gevt->req_method);

    // req_path
    auto req_path_attr = kevt->add_user_attributes();
    req_path_attr->set_key("req_path");
    req_path_attr->set_value_type(CHARBUF);
    req_path_attr->set_value(gevt->req_path);

    // resp_status
    auto resp_status_attr = kevt->add_user_attributes();
    resp_status_attr->set_key("resp_status");
    resp_status_attr->set_value_type(INT64);
    resp_status_attr->set_value(&gevt->resp_status, 8);

    // req_body_size
    auto req_size_attr = kevt->add_user_attributes();
    req_size_attr->set_key("req_body_size");
    req_size_attr->set_value_type(INT64);
    req_size_attr->set_value(&gevt->req_body_size, 8);

    // req_body
    auto req_body_attr = kevt->add_user_attributes();
    req_body_attr->set_key("req_body");
    req_body_attr->set_value_type(BYTEBUF);
    req_body_attr->set_value(gevt->req_body);

    // resp_body_size
    auto resp_size_attr = kevt->add_user_attributes();
    resp_size_attr->set_key("resp_body_size");
    resp_size_attr->set_value_type(INT64);
    resp_size_attr->set_value(&gevt->resp_body_size, 8);

    // resp_body
    auto resp_body_attr = kevt->add_user_attributes();
    resp_body_attr->set_key("resp_body");
    resp_body_attr->set_value_type(BYTEBUF);
    resp_body_attr->set_value(gevt->resp_body);

    // latency done.
    auto latency_attr = kevt->add_user_attributes();
    latency_attr->set_key("latency");
    latency_attr->set_value_type(INT64);
    latency_attr->set_value(&gevt->latency, 8);

    // container_id
    auto cid_attr = kevt->add_user_attributes();
    cid_attr->set_key("containerid");
    cid_attr->set_value_type(CHARBUF);
    cid_attr->set_value(gevt->container_id);
    
    VLOG(1) << absl::Substitute("[grpc] uprobe convert end ");

    return;
}
