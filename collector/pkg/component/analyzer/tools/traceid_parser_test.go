package tools

import "testing"

func TestParseTraceHeader(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		traceType string
		traceId   string
	}{
		{name: "zipkin", key: "x-b3-traceid", value: "223f3b00a283c75c", traceType: "zipkin", traceId: "223f3b00a283c75c"},
		{name: "jaeger", key: "uber-trace-id", value: "3997ed0a6a71f050:cf49be2de63d86e7:e02475aab05fd358:1", traceType: "jaeger", traceId: "3997ed0a6a71f050"},
		{name: "w3c-request", key: "traceparent", value: "00-4bf92f3577b34da6a3ce929d0e0e4736-d75597dee50b0cac-00", traceType: "w3c", traceId: "4bf92f3577b34da6a3ce929d0e0e4736"},
		{name: "w3c-response", key: "traceresponse", value: "00-4bf92f3577b34da6a3ce929d0e0e4736-828c5d0d435ba505-01", traceType: "w3c", traceId: "4bf92f3577b34da6a3ce929d0e0e4736"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				tt.key: tt.value,
			}

			traceType, traceId := ParseTraceHeader(headers)
			if traceType != tt.traceType {
				t.Errorf("Fail to check traceType, got = %s, want %s", traceType, tt.traceType)
			}

			if traceId != tt.traceId {
				t.Errorf("Fail to check traceId, got = %s, want %s", traceId, tt.traceId)
			}
		})
	}
}
