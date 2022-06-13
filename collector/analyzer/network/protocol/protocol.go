package protocol

const (
	HTTP      = "http"
	DNS       = "dns"
	KAFKA     = "kafka"
	MYSQL     = "mysql"
	REDIS     = "redis"
	DUBBO2    = "dubbo"
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
	return 80
}

func GetHttpPayLoadLength() int {
	return GetPayLoadLength(HTTP)
}

func GetDubbo2PayLoadLength() int {
	return GetPayLoadLength(DUBBO2)
}
