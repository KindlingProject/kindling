package http

import (
	"strings"

	"github.com/dxsup/kindling-collector/analyzer/network/protocol"
	"github.com/dxsup/kindling-collector/analyzer/network/protocol/http/tools"
	"github.com/dxsup/kindling-collector/model/constlabels"
)

var httpMerger = tools.NewHttpMergeCache()

func fastfailHttpRequest() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data[message.Offset:]) < 14
	}
}

/*
Request line
    Method [GET/POST/PUT/DELETE/HEAD/TRACE/OPTIONS/CONNECT]
	Blank
	Request-URI [eg. /xxx/yyy?parm0=aaa&param1=bbb]
	Blank
	HTTP-Version [HTTP/1.0 | HTTP/1.2]
	\r\n
Request header
Request body
*/
func parseHttpRequest() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		offset, method := message.ReadUntilBlankWithLength(message.Offset, 8)

		if !httpMethodsList[string(method)] || message.Data[offset-1] != ' ' {
			return false, true
		}

		_, url := message.ReadUntilBlank(offset)

		headers := parseHeaders(message)
		traceType, traceId := tools.ParseTraceHeader(headers)
		if len(traceType) > 0 && len(traceId) > 0 {
			message.AddStringAttribute(constlabels.HttpApmTraceType, traceType)
			message.AddStringAttribute(constlabels.HttpApmTraceId, traceId)
		}

		message.AddStringAttribute(constlabels.HttpMethod, string(method))
		message.AddStringAttribute(constlabels.HttpUrl, string(url))
		message.AddStringAttribute(constlabels.HttpRequestPayload, string(message.GetData(0, 80)))

		contentKey := getContentKey(string(url))
		if len(contentKey) == 0 {
			contentKey = "*"
		}
		message.AddStringAttribute(constlabels.ContentKey, contentKey)
		return true, true
	}
}

func getContentKey(url string) string {
	if url == "" {
		return ""
	}
	index := strings.Index(url, "?")
	if index != -1 {
		url = url[:index]
	}
	return httpMerger.GetContentKey(url)
}

var httpMethodsList = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"HEAD":    true,
	"TRACE":   true,
	"OPTIONS": true,
	"CONNECT": true,
}
