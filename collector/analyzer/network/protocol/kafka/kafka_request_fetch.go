package kafka

import (
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
)

func fastfailRequestFetch() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.GetIntAttribute(constlabels.KafkaApi) != _apiFetch
	}
}

func parseRequestFetch() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		var (
			offset    int
			err       error
			topicNum  int32
			topicName string
		)
		version := message.GetIntAttribute(constlabels.KafkaVersion)
		compact := version >= 12

		// replica_id, max_wait_ms, min_bytes,
		offset = message.Offset + 12
		if version >= 3 {
			offset += 4 // max_bytes
		}
		if version >= 4 {
			offset += 1 // isolation_level
		}
		if version >= 7 {
			offset += 8 // session_id, session_epoch
		}

		if offset, err = message.ReadArraySize(offset, compact, &topicNum); err != nil {
			return false, true
		}
		if topicNum > 0 {
			if _, err = message.ReadString(offset, compact, &topicName); err != nil {
				return false, true
			}
			/*
				TODO Get all topicNames(Ver 0~12)
				There is not enough cases to cover multi-topics, just read first topicName.
				Since Version 13, topicName will be repalced with topicId as uuid, this will not got.
			*/
			message.AddUtf8StringAttribute(constlabels.KafkaTopic, topicName)
		}

		return true, true
	}
}
