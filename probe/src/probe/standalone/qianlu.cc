#include <iostream>
#include <vector>
#include <src/stirling/stirling.h>
#include <src/common/base/base.h>
#include <src/stirling/source_connectors/socket_tracer/http_table.h>
#include "src/probe/utils/termination_handler.h"

px::Status consume_uprobe_data(uint64_t table_id, px::types::TabletID tablet_id,
                                          std::unique_ptr<px::types::ColumnWrapperRecordBatch> record_batch) {

    std::cout << "[qianlu][grpc] table_id:" << table_id << " tablet_id:" << tablet_id << std::endl;

    if (record_batch->empty() || record_batch->at(0)->Size() == 0) {
        std::cout << "[qianlu][grpc] record_batch is empty. table_id:" << table_id << std::endl;
        return px::Status::OK();
    }

    std::cout << "[qianlu][grpc] record_batch cols is " << record_batch->size() << " samples:" << record_batch->at(0)->Size() << std::endl;
//
//    auto it = m_kindlingEventLists.find(uprobe_converter_);
//    KindlingEventList* kindlingEventList;
//    if (it == m_kindlingEventLists.end()) {
//        kindlingEventList = new KindlingEventList();
//        m_kindlingEventLists[uprobe_converter_] = kindlingEventList;
//    } else {
//        kindlingEventList = it->second;
//    }
    if (record_batch->size() != 20) {
	std::cout << "[qianlu] size not match" << " table_id:" << table_id << " tablet_id:" << tablet_id << std::endl;
	return px::Status::OK();
    }
    int ids = 0;
    for (const auto& col : *record_batch) {
	auto received_type = col->data_type();
	std::cout << "ids:" << ids ++ << " types:" << px::types::ToString(received_type) << std::endl;
    }

    auto batch_size = record_batch->at(0)->Size();
    for (size_t i = 0; i < batch_size; ++ i ) {
        std::cout << "[qianlu][grpc] begin to process record " << i << " ... " << std::endl;
        int64_t ts = record_batch->at(px::stirling::kHTTPTimeIdx)->Get<px::types::Time64NSValue>(i).val;
	std::cout << "ts:" << ts << " idx:" << px::stirling::kHTTPTimeIdx << std::endl;
        int32_t pid = record_batch->at(px::stirling::kHTTPUPIDIdx)->Get<px::types::UInt128Value>(i).High64();
	std::cout << "pid:" << pid << " idx:" << px::stirling::kHTTPUPIDIdx << std::endl;
        std::string remote_addr = record_batch->at(px::stirling::kHTTPRemoteAddrIdx)->Get<px::types::StringValue>(i);
	std::cout << "remote_addr:" << remote_addr << " idx:" << px::stirling::kHTTPRemoteAddrIdx << std::endl;
        int64_t remote_port = record_batch->at(px::stirling::kHTTPRemotePortIdx)->Get<px::types::Int64Value>(i).val;
	std::cout << "remote_port:" << remote_port << " idx:" << px::stirling::kHTTPRemotePortIdx << std::endl;
        int64_t trace_role = record_batch->at(px::stirling::kHTTPTraceRoleIdx)->Get<px::types::Int64Value>(i).val;
	std::cout << "trace_role:" << trace_role << " idx:" << px::stirling::kHTTPTraceRoleIdx << std::endl;
	int64_t major_version = record_batch->at(px::stirling::kHTTPMajorVersionIdx)->Get<px::types::Int64Value>(i).val;
	std::cout << "major_version:" << major_version << " idx:" << px::stirling::kHTTPMajorVersionIdx << std::endl;
	int64_t minor_version = record_batch->at(px::stirling::kHTTPMinorVersionIdx)->Get<px::types::Int64Value>(i).val;
	std::cout << "minor_version:" << minor_version << " idx:" << px::stirling::kHTTPMinorVersionIdx << std::endl;
	int64_t content_type = record_batch->at(px::stirling::kHTTPContentTypeIdx)->Get<px::types::Int64Value>(i).val;
	std::cout << "content_type:" << content_type << " idx:" << px::stirling::kHTTPContentTypeIdx << std::endl;
        std::string req_headers = record_batch->at(px::stirling::kHTTPReqHeadersIdx)->Get<px::types::StringValue>(i);
	std::cout << "req_headers:" << req_headers << " idx:" << px::stirling::kHTTPReqHeadersIdx << std::endl;
        std::string req_method = record_batch->at(px::stirling::kHTTPReqMethodIdx)->Get<px::types::StringValue>(i);
	std::cout << "req_method:" << req_method << " idx:" << px::stirling::kHTTPReqMethodIdx << std::endl;
        std::string req_path = record_batch->at(px::stirling::kHTTPReqPathIdx)->Get<px::types::StringValue>(i);
	std::cout << "req_path:" << req_path << std::endl;
        std::string req_body = record_batch->at(px::stirling::kHTTPReqBodyIdx)->Get<px::types::StringValue>(i);
	std::cout << "req_body:" << req_body << std::endl;
        int64_t req_body_size = record_batch->at(px::stirling::kHTTPReqBodySizeIdx)->Get<px::types::Int64Value>(i).val;
	std::cout << "req_body_size:" << req_body_size << std::endl;
	std::string resp_headers = record_batch->at(px::stirling::kHTTPRespHeadersIdx)->Get<px::types::StringValue>(i);
	std::cout << "resp_headers:" << resp_headers << std::endl;
        int64_t resp_status = record_batch->at(px::stirling::kHTTPRespStatusIdx)->Get<px::types::Int64Value>(i).val;
	std::cout << "resp_status:" << resp_status << std::endl;
        std::string resp_body = record_batch->at(px::stirling::kHTTPRespBodyIdx)->Get<px::types::StringValue>(i);
	std::cout << "resp_body:" << resp_body << std::endl;
        int64_t resp_body_size = record_batch->at(px::stirling::kHTTPRespBodySizeIdx)->Get<px::types::Int64Value>(i).val;
	std::cout << "resp_body_size:" << resp_body_size << std::endl;
        int64_t latency = record_batch->at(px::stirling::kHTTPLatencyIdx)->Get<px::types::Int64Value>(i).val;
	std::cout << "latency:" << latency << std::endl;

        std::cout << "[qianlu][grpc] ts:" << ts << " pid:" << pid << " remote_addr:" << remote_addr << " remote_port:" << remote_port << " trace_role:" << trace_role << std::endl;

//        struct grpc_event_t gevt;
//        gevt.timestamp = ts;
//        gevt.pid = pid;
//        gevt.remote_addr = remote_addr;
//        gevt.remote_port = remote_port;
//        gevt.trace_role = trace_role;
//        gevt.req_headers = req_headers;
//        gevt.req_method = req_method;
//        gevt.req_path = req_path;
//        gevt.req_body = req_body;
//        gevt.req_body_size = req_body_size;
//        gevt.resp_status = resp_status;
//        gevt.resp_body = resp_body;
//        gevt.resp_body_size = resp_body_size;
//        gevt.latency = latency;

//        auto tinfo = m_inspector->get_thread_ref(pid, true, true, true);
//        if (tinfo) {
//            gevt.container_id = tinfo->m_container_id;
//            std::cout << "[qianlu] find container_id for pid:" << pid << " container_id:" << gevt.container_id << std::endl;
//        } else {
//            std::cout << "[qianlu] cannot find container_id for pid:" << pid << std::endl;
//        }

        // convert数据
//        event_mutex_.lock();
//        KindlingEvent *kindlingEvent = kindlingEventList->add_kindling_event_list();
//        uprobe_converter_->convert(kindlingEvent, &gevt);
//        event_mutex_.unlock();
    }
    return px::Status::OK();
}

int main(int argc, char** argv) {
    px::EnvironmentGuard env_guard(&argc, argv);

    TerminationHandler::InstallSignalHandlers();

    // init bcc & stirling
    auto stirling = px::stirling::Stirling::Create(px::stirling::CreateSourceRegistry(px::stirling::GetSourceNamesForGroup(px::stirling::SourceConnectorGroup::kTracers))
                                                           .ConsumeValueOrDie());
    TerminationHandler::set_stirling(stirling.get());
    std::cout << "hello, qianlu!" << std::endl;
    stirling->RegisterDataPushCallback(consume_uprobe_data);
    std::cout << "register data push callback done." << std::endl;
    auto status = stirling->RunAsThread();
    std::cout << status.ok() << "begin to run core" << std::endl;

    while (true) {
        sleep(100);
    }

    stirling->Stop();
    return 0;
}
