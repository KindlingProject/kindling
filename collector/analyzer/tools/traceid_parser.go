package tools

import "strings"

func ParseTraceHeader(headers map[string]string) (traceType string, traceId string) {
	if harmonycloud, ok := headers["apm-transactionid"]; ok {
		return "harmonycloud", harmonycloud
	}

	if zipkin, ok := headers["x-b3-traceid"]; ok {
		return "zipkin", zipkin
	}

	if jaeger, ok := headers["uber-trace-id"]; ok {
		if pos := strings.Index(jaeger, ":"); pos > 0 {
			return "jaeger", jaeger[0:pos]
		}
		return "jaeger", jaeger
	}

	if w3c, ok := headers["traceparent"]; ok && len(w3c) >= 35 {
		return "w3c", w3c[3:35]
	}
	if w3c, ok := headers["traceresponse"]; ok && len(w3c) >= 35 {
		return "w3c", w3c[3:35]
	}

	return "", ""
}
