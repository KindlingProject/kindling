package constvalues

const (
	RequestCount        = "request_count"
	RequestTotalTime    = "request_total_time"
	ConnectTime         = "connect_time"
	RequestSentTime     = "request_sent_time"
	WaitingTtfbTime     = "waiting_ttfb_time"
	ContentDownloadTime = "content_download_time"

	RequestIo  = "request_io"
	ResponseIo = "response_io"

	SpanInfo = "KSpanInfo"
)

const (
	TcpRttMetricName        = "kindling_tcp_rtt_microseconds"
	TcpRetransmitMetricName = "kindling_tcp_retransmit_total"
	TcpDropMetricName       = "kindling_tcp_packet_loss_total"
)

const (
	ProtocolHttp  = "http"
	ProtocolHttp2 = "http2"
	ProtocolGrpc  = "grpc"
	ProtocolDubbo = "dubbo"
	ProtocolDns   = "dns"
	ProtocolKafka = "kafka"
	ProtocolMysql = "mysql"
)
