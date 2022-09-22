package redis

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
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
		message.AddByteArrayUtf8Attribute(constlabels.RequestPayload, message.GetData(0, protocol.GetPayLoadLength(protocol.REDIS)))
		return true, false
	}
}
