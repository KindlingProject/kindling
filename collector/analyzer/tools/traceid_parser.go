package tools

import "strings"

func ParseTraceHeader(headers map[string]string) (traceType string, traceId string) {
	if eagleEye, ok := headers["eagleeye-traceid"]; ok {
		return "arms", eagleEye
	}

	if zipkin, ok := headers["x-b3-traceid"]; ok {
		return "zipkin", zipkin
	}

	if jaeger, ok := headers["uber-trace-id"]; ok {
		if pos := strings.Index(jaeger, ":"); pos > 0 {
			return "zipkin", jaeger[0:pos]
		}
		return "zipkin", jaeger
	}

	if w3c, ok := headers["traceparent"]; ok && len(w3c) >= 32 {
		return "w3c", w3c[3:32]
	}

	return "", ""
}
