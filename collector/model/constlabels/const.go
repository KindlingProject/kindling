package constlabels

const (
	NoError = iota
	ConnectFail
	NoResponse
	ProtocolError
)

const (
	Pid             = "pid"
	Protocol        = "protocol"
	IsError         = "is_error"
	ErrorType       = "error_type"
	IsSlow          = "is_slow"
	IsServer        = "is_server"
	ContainerId     = "container_id"
	SrcNode         = "src_node"
	SrcNodeIp       = "src_node_ip"
	SrcNamespace    = "src_namespace"
	SrcPod          = "src_pod"
	SrcWorkloadName = "src_workload_name"
	SrcWorkloadKind = "src_workload_kind"
	SrcService      = "src_service"
	SrcIp           = "src_ip"
	SrcPort         = "src_port"
	SrcContainerId  = "src_container_id"
	SrcContainer    = "src_container"
	DstNode         = "dst_node"
	DstNodeIp       = "dst_node_ip"
	DstNamespace    = "dst_namespace"
	DstPod          = "dst_pod"
	DstWorkloadName = "dst_workload_name"
	DstWorkloadKind = "dst_workload_kind"
	DstService      = "dst_service"
	DstIp           = "dst_ip"
	DstPort         = "dst_port"
	DnatIp          = "dnat_ip"
	DnatPort        = "dnat_port"
	DstContainerId  = "dst_container_id"
	DstContainer    = "dst_container"
	Node            = "node"
	Namespace       = "namespace"
	WorkloadKind    = "workload_kind"
	WorkloadName    = "workload_name"
	Service         = "service"
	Pod             = "pod"
	Container       = "container"
	Ip              = "ip"
	Port            = "port"

	RequestContent  = "request_content"
	ResponseContent = "response_content"
	StatusCode      = "status_code"

	Topic      = "topic"
	Operation  = "operation"
	ConsumerId = "consumer_id"

	RequestDurationStatus   = "request_duration_status"
	RequestReqxferStatus    = "request_reqxfer_status"
	RequestProcessingStatus = "request_processing_status"
	ResponseRspxferStatus   = "response_rspxfer_status"

	ExternalClusterNamespace = "EXTERNAL"
)

const (
	STR_EMPTY = ""
)
