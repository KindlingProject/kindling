package dubbo

import (
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
)

func fastfailDubboRequest() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 16 || message.Data[0] != MAGIC_HIGH || message.Data[1] != MAGIC_LOW
	}
}

func parseDubboRequest() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		contentKey := getContentKey(message.Data)
		if contentKey == "" {
			return false, true
		}

		message.AddStringAttribute(constlabels.ContentKey, contentKey)
		message.AddStringAttribute(constlabels.DubboRequestPayload, getAsciiString(message.GetData(16, protocol.GetDubboPayLoadLength())))
		return true, true
	}
}

func getContentKey(requestData []byte) string {
	serialID := requestData[2] & SERIAL_MASK
	if serialID == Zero {
		return ""
	}
	if (requestData[2] & FLAG_EVENT) != Zero {
		return "Heartbeat"
	}
	if (requestData[2] & FLAG_REQUEST) == Zero {
		// Invalid Data
		return ""
	}
	if (requestData[2] & FLAG_TWOWAY) == Zero {
		// Ignore Oneway Data
		return "Oneway"
	}

	serializer := GetSerializer(serialID)
	if serializer == serial_unsupport {
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
