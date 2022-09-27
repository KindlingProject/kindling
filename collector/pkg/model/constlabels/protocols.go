package constlabels

const (
	ContentKey      = "content_key"
	RequestPayload  = "request_payload"
	ResponsePayload = "response_payload"

	HttpMethod       = "http_method"
	HttpUrl          = "http_url"
	HttpApmTraceType = "trace_type"
	HttpApmTraceId   = "trace_id"
	HttpStatusCode   = "http_status_code"

	DnsId     = "dns_id"
	DnsDomain = "dns_domain"
	DnsRcode  = "dns_rcode"
	DnsIp     = "dns_ip"

	Sql        = "sql"
	SqlErrCode = "sql_error_code"
	SqlErrMsg  = "sql_error_msg"

	RedisCommand = "redis_command"
	RedisErrMsg  = "redis_error_msg"

	KafkaApi           = "kafka_api"
	KafkaVersion       = "kafka_version"
	KafkaCorrelationId = "kafka_id"
	KafkaTopic         = "kafka_topic"
	KafkaErrorCode     = "kafka_error_code"

	DubboErrorCode = "dubbo_error_code"
)
