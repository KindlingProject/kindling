package kafka

import (
	"github.com/dxsup/kindling-collector/analyzer/network/protocol"
	"github.com/dxsup/kindling-collector/model/constlabels"
)

func fastfailRequestFetch() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.GetIntAttribute(constlabels.KafkaApi) != _apiFetch
	}
}

func parseRequestFetch() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		var (
			offset       int
			err          error
			topicNum     int32
			topicName    string
			partitionNum int32
			partition    int32
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

		if offset, err = message.ReadArraySize(offset, compact, &topicNum); err != nil || topicNum != 1 {
			return false, true
		}
		if offset, err = message.ReadString(offset, compact, &topicName); err != nil {
			return false, true
		}
		if offset, err = message.ReadArraySize(offset, compact, &partitionNum); err != nil || partitionNum != 1 {
			return false, true
		}
		if _, err = message.ReadInt32(offset, &partition); err != nil {
			return false, true
		}
		message.AddStringAttribute(constlabels.KafkaTopic, string(topicName))
		message.AddIntAttribute(constlabels.KafkaPartition, int64(partition))

		return true, true
	}
}
