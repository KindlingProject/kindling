package http

import (
	"testing"
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
			want: "/test/*",
		},
		{
			name: "two segments",
			args: args{url: "/test/123"},
			want: "/test/*",
		},
		{
			name: "two segments, but more than two /",
			args: args{url: "/test/123/"},
			want: "/test/*",
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
		{name: "normal", args: args{url: "/test/arg/sar?a=12314"}, want: "/test/arg"},
		{name: "normal", args: args{url: "/test"}, want: "/test"},
		{name: "zero", args: args{url: ""}, want: ""},
		{name: "zero", args: args{url: "/test/arg/adf/fadf/adf"}, want: "/test/arg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getContentKey(tt.args.url); got != tt.want {
				t.Errorf("getContentKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
