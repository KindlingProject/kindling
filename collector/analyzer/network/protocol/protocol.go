package protocol

const (
	HTTP      = "http"
	DNS       = "dns"
	KAFKA     = "kafka"
	MYSQL     = "mysql"
	REDIS     = "redis"
	NOSUPPORT = "NOSUPPORT"
)

var (
	http_payLoad_length int = 80
)

func SetHttpPayLoadLength(length int) {
	http_payLoad_length = length
}

func GetHttpPayLoadLength() int {
	return http_payLoad_length
}
