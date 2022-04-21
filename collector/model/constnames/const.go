package constnames

const (
	ReadEvent     = "read"
	WriteEvent    = "write"
	ReadvEvent    = "readv"
	WritevEvent   = "writev"
	SendToEvent   = "sendto"
	RecvFromEvent = "recvfrom"
	SendMsgEvent  = "sendmsg"
	RecvMsgEvent  = "recvmsg"

	TcpCloseEvent          = "tcp_close"
	TcpRcvEstablishedEvent = "tcp_rcv_established"
	TcpDropEvent           = "tcp_drop"
	TcpRetransmitSkbEvent  = "tcp_retransmit_skb"
	OtherEvent             = "other"

	GrpcUprobeEvent = "grpc_uprobe"
	// NetRequestGaugeGroupName is used for gaugeGroup generated from networkAnalyzer.
	NetRequestGaugeGroupName = "net_request_gauge_group"
	// SingleNetRequestGaugeGroup stands for the gaugeGroup with abnormal status.
	SingleNetRequestGaugeGroup = "single_net_request_gauge_group"
	// AggregatedNetRequestGaugeGroup stands for the gaugeGroup after aggregation.
	AggregatedNetRequestGaugeGroup = "aggregated_net_request_gauge_group"

	TcpGaugeGroupName  = "tcp_metric_gauge_group"
	NodeGaugeGroupName = "node_metric_gauge_group"
)
