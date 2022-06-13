package dubbo2

import (
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
)

func fastfailDubbo2Request() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 16 || message.Data[0] != MagicHigh || message.Data[1] != MagicLow
	}
}

func parseDubbo2Request() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		contentKey := getContentKey(message.Data)
		if contentKey == "" {
			return false, true
		}

		message.AddStringAttribute(constlabels.ContentKey, contentKey)
		message.AddStringAttribute(constlabels.Dubbo2RequestPayload, getAsciiString(message.GetData(16, protocol.GetDubbo2PayLoadLength())))
		return true, true
	}
}

func getContentKey(requestData []byte) string {
	serialID := requestData[2] & SerialMask
	if serialID == Zero {
		return ""
	}
	if (requestData[2] & FlagEvent) != Zero {
		return "Heartbeat"
	}
	if (requestData[2] & FlagRequest) == Zero {
		// Invalid Data
		return ""
	}
	if (requestData[2] & FlagTwoWay) == Zero {
		// Ignore Oneway Data
		return "Oneway"
	}

	serializer := GetSerializer(serialID)
	if serializer == serialUnsupport {
		// Unsupport Serial. only support hessian and fastjson.
		return "UnSupportSerialFormat"
	}

	var (
		service string
		method  string
	)
	// version
	offset := serializer.eatString(requestData, 16)

	// service name
	offset, service = serializer.getStringValue(requestData, offset)

	// service version
	offset = serializer.eatString(requestData, offset)

	// method name
	_, method = serializer.getStringValue(requestData, offset)

	return service + "#" + method
}
