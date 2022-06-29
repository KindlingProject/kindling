package constvalues

const (
	RequestCount         = "request_count"
	RequestTotalTime     = "request_total_time"
	ConnectTime          = "connect_time"
	RequestSentTime      = "request_sent_time"
	WaitingTtfbTime      = "waiting_ttfb_time"
	ContentDownloadTime  = "content_download_time"
	RequestTimeHistogram = "request_time_histogram"

	RequestIo  = "request_io"
	ResponseIo = "response_io"

	SpanInfo = "KSpanInfo"
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
