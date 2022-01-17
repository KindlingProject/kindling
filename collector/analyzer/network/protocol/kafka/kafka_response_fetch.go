package kafka

import (
	"github.com/dxsup/kindling-collector/analyzer/network/protocol"
	"github.com/dxsup/kindling-collector/model/constlabels"
)

func fastfailResponseFetch() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.GetIntAttribute(constlabels.KafkaApi) != _apiFetch
	}
}

func parseResponseFetch() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		var (
			offset       int
			err          error
			topicNum     int32
			topicName    string
			partitionNum int32
			partition    int32
			errorCode    int16
		)

		version := message.GetIntAttribute(constlabels.KafkaVersion)
		compact := version >= 12
		offset = message.Offset

		if version >= 1 {
			offset += 4 // throttle_time_ms
		}
		if version >= 7 {
			if offset, err = message.ReadInt16(offset, &errorCode); err != nil {
				return false, true
			}
			offset += 4 //session_id
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
		if offset, err = message.ReadInt32(offset, &partition); err != nil {
			return false, true
		}
		if version < 7 {
			if _, err = message.ReadInt16(offset, &errorCode); err != nil {
				return false, true
			}
		}
		message.AddStringAttribute(constlabels.KafkaTopic, topicName)
		message.AddIntAttribute(constlabels.KafkaPartition, int64(partition))
		message.AddIntAttribute(constlabels.KafkaErrorCode, int64(errorCode))
		return true, true
	}
}
