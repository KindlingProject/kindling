package factory

import (
	"sync"

	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol/dns"
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol/generic"
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol/http"
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol/kafka"
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol/mysql"
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol/redis"
)

var (
	cache_port_parsers_map = make(map[uint32][]*protocol.ProtocolParser)
	mutex                  = sync.Mutex{}

	generic_parser *protocol.ProtocolParser = generic.NewGenericParser()
	http_parser    *protocol.ProtocolParser = http.NewHttpParser()
	kafka_parser   *protocol.ProtocolParser = kafka.NewKafkaParser()
	mysql_parser   *protocol.ProtocolParser = mysql.NewMysqlParser()
	redis_parser   *protocol.ProtocolParser = redis.NewRedisParser()
	dns_parser     *protocol.ProtocolParser = dns.NewDnsParser()
)

func GetParser(key string) *protocol.ProtocolParser {
	switch key {
	case protocol.HTTP:
		return http_parser
	case protocol.KAFKA:
		return kafka_parser
	case protocol.MYSQL:
		return mysql_parser
	case protocol.REDIS:
		return redis_parser
	case protocol.DNS:
		return dns_parser
	default:
		return nil
	}
}

func GetGenericParser() *protocol.ProtocolParser {
	return generic_parser
}

func GetCachedParsersByPort(port uint32) ([]*protocol.ProtocolParser, bool) {
	mutex.Lock()
	parser, ok := cache_port_parsers_map[port]
	mutex.Unlock()

	return parser, ok
}

func AddCachedParser(port uint32, parser *protocol.ProtocolParser) {
	mutex.Lock()
	if val := cache_port_parsers_map[port]; val == nil {
		parsers := make([]*protocol.ProtocolParser, 0)
		parsers = append(parsers, parser)
		cache_port_parsers_map[port] = parsers
	} else {
		exist := false
		for _, value := range val {
			if value == parser {
				exist = true
				break
			}
		}

		if !exist {
			// Make sure Generic is last
			if len(val) > 0 && val[len(val)-1] == generic_parser {
				parsers := append(val[0:len(val)-1], parser)
				parsers = append(parsers, generic_parser)
				cache_port_parsers_map[port] = parsers
			} else {
				parsers := append(val, parser)
				cache_port_parsers_map[port] = parsers
			}
		}
	}
	mutex.Unlock()
}

func RemoveCachedParser(port uint32, parser *protocol.ProtocolParser) {
	mutex.Lock()
	if val, ok := cache_port_parsers_map[port]; ok {
		for i, value := range val {
			if value == parser {
				val = append(val[:i], val[i+1:]...)
				cache_port_parsers_map[port] = val
			}
		}
	}
	mutex.Unlock()
}
