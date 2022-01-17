package constnames

const (
	TcpCloseEvent          string = "tcp_close"
	TcpRcvEstablishedEvent string = "tcp_rcv_established"
	TcpDropEvent           string = "tcp_drop"
	TcpRetransmitSkbEvent  string = "tcp_retransmit_skb"

	GrpcUprobeEvent          string = "grpc_uprobe"
	NetRequestGaugeGroupName        = "net_request_gauge_group"
	TcpGaugeGroupName               = "tcp_metric_gauge_group"
	NodeGaugeGroupName              = "node_metric_gauge_group"
)
