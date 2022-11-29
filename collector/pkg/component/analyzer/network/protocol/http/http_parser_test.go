package http

import (
	"reflect"
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func Test_urlMerge(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{

		{
			name: "more than two segments",
			args: args{url: "/test/123/456/test?a=1"},
			want: "/test/123/456/test",
		},
		{
			name: "two segments",
			args: args{url: "/test/123"},
			want: "/test/123",
		},
		{
			name: "two segments, but more than two /",
			args: args{url: "/test/123/"},
			want: "/test/123/",
		},
		{
			name: "one segment",
			args: args{url: "/test"},
			want: "/test",
		},
		{
			name: "one segment, but two /",
			args: args{url: "/test/"},
			want: "/test/",
		},
		{
			name: "only one /",
			args: args{url: "/"},
			want: "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getContentKey(tt.args.url); got != tt.want {
				t.Errorf("UrlMerge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpParser_getContentKey(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "normal", args: args{url: "/test/arg?a=12314"}, want: "/test/arg"},
		{name: "normal", args: args{url: "/test/arg/sar?a=12314"}, want: "/test/arg/sar"},
		{name: "normal", args: args{url: "/test"}, want: "/test"},
		{name: "zero", args: args{url: ""}, want: ""},
		{name: "zero", args: args{url: "/test/arg/adf/fadf/adf"}, want: "/test/arg/adf/fadf/adf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getContentKey(tt.args.url); got != tt.want {
				t.Errorf("getContentKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseHttpRequest_GetPayLoad(t *testing.T) {
	type args struct {
		message *protocol.PayloadMessage
	}
	httpData := "POST /test?sleep=0&respbyte=1000&statusCode=200 HTTP/1.1\r\nKey1: value1\r\n\r\nHello world"

	tests := []struct {
		name string
		size int
		want string
	}{
		{name: "substring", size: 10, want: "POST /test"},
		{name: "equal", size: 85, want: httpData},
		{name: "overflow", size: 100, want: httpData},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protocol.SetPayLoadLength(protocol.HTTP, tt.size)
			message := protocol.NewRequestMessage([]byte(httpData))
			NewHttpParser("").ParseRequest(message)

			if !message.HasAttribute(constlabels.RequestPayload) {
				t.Errorf("Fail to parse HttpRequest()")
			}
			if got := message.GetStringAttribute(constlabels.RequestPayload); got != tt.want {
				t.Errorf("GetHttpPayload() = %v, want %v", got, tt.want)
			}
		})
	}

	tests2 := []struct {
		name string
		args args
		size int
		want map[string]string
	}{
		{
			name: "normal case",
			args: args{
				message: protocol.NewRequestMessage([]byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\nAPM-AgentID: TTXvC3EQS6KLwxx3eIqINFjAW2olRm+cr8M+yuvwhkY=\r\nTransfer-Encoding: chunked\r\nContent-Type: application/json\r\nAPM-TransactionID: 5e480579c718a4a6498a9")),
			},
			want: map[string]string{
				"connection":        "keep-alive",
				"apm-agentid":       "TTXvC3EQS6KLwxx3eIqINFjAW2olRm+cr8M+yuvwhkY=",
				"transfer-encoding": "chunked",
				"content-type":      "application/json",
				"apm-transactionid": "5e480579c718a4a6498a9",
			},
		},
		{
			name: "no values",
			args: args{
				protocol.NewRequestMessage([]byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\nTransfer-Encoding: ")),
			},
			want: map[string]string{
				"connection": "keep-alive",
			},
		},
		{

			name: "no spaces",
			args: args{
				protocol.NewRequestMessage([]byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\nTransfer-Encoding:")),
			},
			want: map[string]string{
				"connection": "keep-alive",
			},
		},
		{
			name: "no colon",
			args: args{
				protocol.NewRequestMessage([]byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\nTransfer-Encoding")),
			},
			want: map[string]string{
				"connection": "keep-alive",
			},
		},
	}
	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			protocol.SetPayLoadLength(protocol.HTTP, tt.size)

			message := protocol.NewResponseMessage([]byte(httpData), model.NewAttributeMap())
			NewHttpParser("").ParseResponse(message)

			if !message.HasAttribute(constlabels.ResponsePayload) {
				t.Errorf("Fail to parse HttpResponse()")
			}
			if got := message.GetStringAttribute(constlabels.ResponsePayload); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetHttpPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseHeaders(t *testing.T) {
	type args struct {
		message *protocol.PayloadMessage
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "normal case",
			args: args{
				message: protocol.NewRequestMessage([]byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\nAPM-AgentID: TTXvC3EQS6KLwxx3eIqINFjAW2olRm+cr8M+yuvwhkY=\r\nTransfer-Encoding: chunked\r\nContent-Type: application/json\r\nAPM-TransactionID: 5e480579c718a4a6498a9")),
			},
			want: map[string]string{
				"connection":        "keep-alive",
				"apm-agentid":       "TTXvC3EQS6KLwxx3eIqINFjAW2olRm+cr8M+yuvwhkY=",
				"transfer-encoding": "chunked",
				"content-type":      "application/json",
				"apm-transactionid": "5e480579c718a4a6498a9",
			},
		},
		{
			name: "no values",
			args: args{
				protocol.NewRequestMessage([]byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\nTransfer-Encoding: ")),
			},
			want: map[string]string{
				"connection": "keep-alive",
			},
		},
		{

			name: "no spaces",
			args: args{
				protocol.NewRequestMessage([]byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\nTransfer-Encoding:")),
			},
			want: map[string]string{
				"connection": "keep-alive",
			},
		},
		{
			name: "no colon",
			args: args{
				protocol.NewRequestMessage([]byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\nTransfer-Encoding")),
			},
			want: map[string]string{
				"connection": "keep-alive",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseHeaders(tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}
