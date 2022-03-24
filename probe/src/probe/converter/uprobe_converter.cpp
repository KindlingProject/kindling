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

    kevt->set_source(UPROBE);
    kevt->set_name("grpc_uprobe");
    kevt->set_category(CAT_NET);
    kevt->set_timestamp(gevt->timestamp);

    // pid
    auto pid_attr = kevt->add_user_attributes();
    pid_attr->set_key("pid");
    auto pid_value = new AnyValue();
    pid_value->set_uint_value(gevt->pid);
    pid_attr->set_allocated_value(pid_value);


    // fd fake~ @qianlu.kk
    auto fd_attr = kevt->add_user_attributes();
    fd_attr->set_key("fd");
    auto fd_value = new AnyValue();
    fd_value->set_uint_value(1001);
    fd_attr->set_allocated_value(fd_value);

    // remote_addr
    auto ra_attr = kevt->add_user_attributes();
    ra_attr->set_key("remote_addr");
    auto ra_value = new AnyValue();
    ra_value->set_string_value(gevt->remote_addr);
    ra_attr->set_allocated_value(ra_value);

    // remote_port
    auto rp_attr = kevt->add_user_attributes();
    rp_attr->set_key("remote_port");
    auto rp_value = new AnyValue();
    rp_value->set_uint_value(gevt->remote_port);
    rp_attr->set_allocated_value(rp_value);

    // trace_role
    auto tc_attr = kevt->add_user_attributes();
    tc_attr->set_key("trace_role");
    auto tc_value = new AnyValue();
    tc_value->set_uint_value(gevt->trace_role);
    tc_attr->set_allocated_value(tc_value);

    // req_headers
    auto rh_attr = kevt->add_user_attributes();
    rh_attr->set_key("req_headers");
    auto rh_value = new AnyValue();
    rh_value->set_string_value(gevt->req_headers);
    rh_attr->set_allocated_value(rh_value);

    // content_type
//    auto tgid_attr = kevt->add_user_attributes();
//    tgid_attr->set_key("tgid");
//    auto anyValue = new AnyValue();
//    anyValue->set_uint_value(1000);
//    tgid_attr->set_allocated_value(anyValue);

    // req_method
    auto req_method_attr = kevt->add_user_attributes();
    req_method_attr->set_key("req_method");
    auto req_method = new AnyValue();
    req_method->set_string_value(gevt->req_method);
    req_method_attr->set_allocated_value(req_method);

    // req_path
    auto req_path_attr = kevt->add_user_attributes();
    req_path_attr->set_key("req_path");
    auto req_path = new AnyValue();
    req_path->set_string_value(gevt->req_path);
    req_path_attr->set_allocated_value(req_path);

    // resp_status
    auto resp_status_attr = kevt->add_user_attributes();
    resp_status_attr->set_key("resp_status");
    auto resp_status = new AnyValue();
    resp_status->set_uint_value(gevt->resp_status);
    resp_status_attr->set_allocated_value(resp_status);

    // req_body_size
    auto req_size_attr = kevt->add_user_attributes();
    req_size_attr->set_key("req_body_size");
    auto req_body_size = new AnyValue();
    req_body_size->set_uint_value(gevt->req_body_size);
    req_size_attr->set_allocated_value(req_body_size);

    // req_body
    auto req_body_attr = kevt->add_user_attributes();
    req_body_attr->set_key("req_body");
    auto req_body = new AnyValue();
    req_body->set_bytes_value(gevt->req_body);
    req_body_attr->set_allocated_value(req_body);

    // resp_body_size
    auto resp_size_attr = kevt->add_user_attributes();
    resp_size_attr->set_key("resp_body_size");
    auto resp_body_size = new AnyValue();
    resp_body_size->set_uint_value(gevt->resp_body_size);
    resp_size_attr->set_allocated_value(resp_body_size);

    // resp_body
    auto resp_body_attr = kevt->add_user_attributes();
    resp_body_attr->set_key("resp_body");
    auto resp_body = new AnyValue();
    resp_body->set_bytes_value(gevt->resp_body);
    resp_body_attr->set_allocated_value(resp_body);

    // latency done.
    auto latency_attr = kevt->add_user_attributes();
    latency_attr->set_key("latency");
    auto latency_value = new AnyValue();
    latency_value->set_uint_value(gevt->latency);
    latency_attr->set_allocated_value(latency_value);

    // container_id
    auto cid_attr = kevt->add_user_attributes();
    cid_attr->set_key("containerid");
    auto cid_value = new AnyValue();
    cid_value->set_string_value(gevt->container_id);
    cid_attr->set_allocated_value(cid_value);

    return;
}