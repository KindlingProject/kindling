package redis

import (
	"github.com/dxsup/kindling-collector/analyzer/network/protocol"
	"github.com/dxsup/kindling-collector/model/constlabels"
)

/**
-Error message\r\n
*/
func fastfailRedisError() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.Data[message.Offset] != '-'
	}
}

func parseRedisError() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		offset, data := message.ReadUntilCRLF(message.Offset + 1)
		if data == nil {
			return false, true
		}

		message.Offset = offset
		if len(data) > 0 && !message.HasAttribute(constlabels.RedisErrMsg) {
			message.AddStringAttribute(constlabels.RedisErrMsg, string(data))
		}
		return true, message.IsComplete()
	}
}
