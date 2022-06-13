package dubbo2

import (
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
)

func fastfailDubbo2Response() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 16 || message.Data[0] != MagicHigh || message.Data[1] != MagicLow
	}
}

func parseDubbo2Response() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		errorCode := getErrorCode(message.Data)
		if errorCode == -1 {
			return false, true
		}

		message.AddIntAttribute(constlabels.Dubbo2ErrorCode, errorCode)
		if errorCode > 20 {
			message.AddBoolAttribute(constlabels.IsError, true)
			message.AddIntAttribute(constlabels.ErrorType, int64(constlabels.ProtocolError))
		}
		message.AddStringAttribute(constlabels.Dubbo2ResponsePayload, getAsciiString(message.GetData(16, protocol.GetDubbo2PayLoadLength())))
		return true, true
	}
}

func getErrorCode(responseData []byte) int64 {
	SerialID := responseData[2] & SerialMask
	if SerialID == Zero {
		return -1
	}
	if (responseData[2] & FlagEvent) != Zero {
		return 20
	}
	if (responseData[2] & FlagRequest) != Zero {
		// Invalid Data
		return -1
	}

	return int64(responseData[3])
}
