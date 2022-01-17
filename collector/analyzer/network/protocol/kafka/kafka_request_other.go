package kafka

import (
	"github.com/dxsup/kindling-collector/analyzer/network/protocol"
	"github.com/dxsup/kindling-collector/model/constlabels"
)

func fastfailRequestOther() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.GetIntAttribute(constlabels.KafkaApi) <= _apiFetch
	}
}

func parseRequestOther() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		return true, true
	}
}
