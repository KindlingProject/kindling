package protocol

const (
	HTTP      = "http"
	DNS       = "dns"
	KAFKA     = "kafka"
	MYSQL     = "mysql"
	REDIS     = "redis"
	DUBBO     = "dubbo"
	NOSUPPORT = "NOSUPPORT"
)

var payloadLength map[string]int = map[string]int{}

func SetPayLoadLength(protocol string, length int) {
	payloadLength[protocol] = length
}

func GetPayLoadLength(protocol string) int {
	if length, ok := payloadLength[protocol]; ok {
		return length
	}
	return 200
}

func GetHttpPayLoadLength() int {
	return GetPayLoadLength(HTTP)
}

func GetDubboPayLoadLength() int {
	return GetPayLoadLength(DUBBO)
}
