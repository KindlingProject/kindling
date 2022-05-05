package http

import (
	"testing"

	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
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
			NewHttpParser().ParseRequest(message)

			if !message.HasAttribute(constlabels.HttpRequestPayload) {
				t.Errorf("Fail to parse HttpRequest()")
			}
			if got := message.GetStringAttribute(constlabels.HttpRequestPayload); got != tt.want {
				t.Errorf("GetHttpPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseHttpResponse_GetPayLoad(t *testing.T) {
	httpData := "HTTP/1.1 200 OK\r\nContent-Length: 1000\r\nContent-Type: text/plain; charset=utf-8\r\nKey1: value1\r\n\r\nHello world"

	tests := []struct {
		name string
		size int
		want string
	}{
		{name: "substring", size: 10, want: "HTTP/1.1 2"},
		{name: "equal", size: 107, want: httpData},
		{name: "overflow", size: 200, want: httpData},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protocol.SetPayLoadLength(protocol.HTTP, tt.size)

			message := protocol.NewResponseMessage([]byte(httpData), model.NewAttributeMap())
			NewHttpParser().ParseResponse(message)

			if !message.HasAttribute(constlabels.HttpResponsePayload) {
				t.Errorf("Fail to parse HttpResponse()")
			}
			if got := message.GetStringAttribute(constlabels.HttpResponsePayload); got != tt.want {
				t.Errorf("GetHttpPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}
