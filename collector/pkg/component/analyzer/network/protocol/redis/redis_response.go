package redis

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

/*
For Simple Strings the first byte of the reply is "+"
For Errors the first byte of the reply is "-"
For Integers the first byte of the reply is ":"
For Bulk Strings the first byte of the reply is "$"
For Arrays the first byte of the reply is "*"
*/
func fastfailResponse() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		keyword := message.Data[message.Offset]
		return keyword != '+' &&
			keyword != '-' &&
			keyword != '*' &&
			keyword != '$' &&
			keyword != ':'
	}
}

func parseResponse() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		message.AddByteArrayUtf8Attribute(constlabels.ResponsePayload, message.GetData(0, protocol.GetPayLoadLength(protocol.REDIS)))
		return true, false
	}
}
