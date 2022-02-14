package http

import (
	"github.com/Kindling-project/kindling/collector/analyzer/tools"
	"strconv"

	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
)

/**
Status line
	HTTP-Version[HTTP/1.0 | HTTP/1.1]
	Blank
	Status-Code
	Blank
	Reason-Phrase
	\r\n
Response header
Response body
*/
func fastfailHttpResponse() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		if len(message.Data[message.Offset:]) < 14 {
			return true
		}
		offset, version := message.ReadUntilBlankWithLength(message.Offset, 9)
		if !httpVersoinList[string(version)] || message.Data[offset-1] != ' ' {
			return true
		}

		message.Offset = offset
		return false
	}
}

func parseHttpResponse() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		_, statusCode := message.ReadUntilBlankWithLength(message.Offset, 6)
		statusCodeI, err := strconv.ParseInt(string(statusCode), 10, 0)
		if err != nil {
			return false, true
		}

		if statusCodeI > 999 || statusCodeI < 99 {
			statusCodeI = 0
		}

		if !message.HasAttribute(constlabels.HttpApmTraceType) {
			headers := parseHeaders(message)
			traceType, traceId := tools.ParseTraceHeader(headers)
			if len(traceType) > 0 && len(traceId) > 0 {
				message.AddStringAttribute(constlabels.HttpApmTraceType, traceType)
				message.AddStringAttribute(constlabels.HttpApmTraceId, traceId)
			}
		}

		message.AddIntAttribute(constlabels.HttpStatusCode, statusCodeI)
		message.AddStringAttribute(constlabels.HttpResponsePayload, string(message.GetData(0, 80)))
		if statusCodeI >= 400 {
			message.AddBoolAttribute(constlabels.IsError, true)
			message.AddIntAttribute(constlabels.ErrorType, int64(constlabels.ProtocolError))
		}
		return true, true
	}
}

var httpVersoinList = map[string]bool{
	"HTTP/1.0": true,
	"HTTP/1.1": true,
}
