package dubbo

import (
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
)

func fastfailDubboResponse() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 16 || message.Data[0] != MAGIC_HIGH || message.Data[1] != MAGIC_LOW
	}
}

func parseDubboResponse() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		errorCode := getErrorCode(message.Data)
		if errorCode == -1 {
			return false, true
		}

		message.AddIntAttribute(constlabels.DubboErrorCode, errorCode)
		if errorCode > 20 {
			message.AddBoolAttribute(constlabels.IsError, true)
			message.AddIntAttribute(constlabels.ErrorType, int64(constlabels.ProtocolError))
		}
		message.AddStringAttribute(constlabels.DubboResponsePayload, getAsciiString(message.GetData(16, protocol.GetDubboPayLoadLength())))
		return true, true
	}
}

func getErrorCode(responseData []byte) int64 {
	SerialID := responseData[2] & SERIAL_MASK
	if SerialID == Zero {
		return -1
	}
	if (responseData[2] & FLAG_EVENT) != Zero {
		return 20
	}
	if (responseData[2] & FLAG_REQUEST) != Zero {
		// Invalid Data
		return -1
	}

	return int64(responseData[3])
}
