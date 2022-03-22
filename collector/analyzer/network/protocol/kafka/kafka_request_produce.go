package kafka

import (
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
)

func fastfailRequestProduce() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.GetIntAttribute(constlabels.KafkaApi) != _apiProduce
	}
}

func parseRequestProduce() protocol.ParsePkgFn {
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
		compact := version >= 9
		offset = message.Offset

		if version >= 3 {
			var transactionId string
			if offset, err = message.ReadNullableString(offset, compact, &transactionId); err != nil {
				//TODO maybe pass this error to caller if we need output the exact error msg to log
				return false, true
			}
		}
		// acks, timeout_ms
		offset += 6
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

		message.AddUtf8StringAttribute(constlabels.KafkaTopic, topicName)
		message.AddIntAttribute(constlabels.KafkaPartition, int64(partition))
		return true, true
	}
}
