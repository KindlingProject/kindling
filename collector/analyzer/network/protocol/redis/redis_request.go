package redis

import (
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
)

/*
For Integers the first byte of the reply is ":"
For Bulk Strings the first byte of the reply is "$"
For Arrays the first byte of the reply is "*"
*/
func fastfailRedisRequest() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		keyword := message.Data[message.Offset]
		return keyword != '*' &&
			keyword != '$' &&
			keyword != ':'
	}
}

func parseRedisRequest() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		return true, false
	}
}
