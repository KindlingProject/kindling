package constlabels

const (
	ContentKey = "content_key"

	HttpMethod          = "http_method"
	HttpUrl             = "http_url"
	HttpApmTraceType    = "trace_type"
	HttpApmTraceId      = "trace_id"
	HttpRequestPayload  = "request_payload"
	HttpResponsePayload = "response_payload"
	HttpStatusCode      = "http_status_code"

	DnsId     = "dns_id"
	DnsDomain = "dns_domain"
	DnsRcode  = "dns_rcode"
	DnsIp     = "dns_ip"

	Sql        = "sql"
	SqlErrCode = "sql_error_code"
	SqlErrMsg  = "sql_error_msg"

	RedisErrMsg = "redis_error_msg"

	KafkaApi           = "kafka_api"
	KafkaVersion       = "kafka_version"
	KafkaCorrelationId = "kafka_id"
	KafkaTopic         = "kafka_topic"
	KafkaPartition     = "kafka_partition"
	KafkaErrorCode     = "kafka_error_code"

	Dubbo2RequestPayload  = "request_payload"
	Dubbo2ResponsePayload = "response_payload"
	Dubbo2ErrorCode       = "dubbo_error_code"
)
