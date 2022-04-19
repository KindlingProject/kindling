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

func GetHttpPayLoadLength() int {
	if length, ok := payloadLength[HTTP]; ok {
		return length
	}
	return 0
}

func GetDubboPayLoadLength() int {
	if length, ok := payloadLength[DUBBO]; ok {
		return length
	}
	return 0
}
