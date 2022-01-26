package kindlingformatprocessor

import "testing"

func Test_fillCommonProtocolLabels(t *testing.T) {
	type args struct {
		g        *gauges
		protocol ProtocolType
		isServer bool
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "HttpTopology",
			args: args{
				g:        newGauges(newInnerGauges(false)),
				protocol: http,
				isServer: false,
			},
			want: map[string]string{
				"status_code": "200",
			},
		},
		{
			name: "HttpEntity",
			args: args{
				g:        newGauges(newInnerGauges(true)),
				protocol: http,
				isServer: true,
			},
			want: map[string]string{
				"request_content":  "httpUrl",
				"response_content": "200",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fillCommonProtocolLabels(tt.args.g, tt.args.protocol, tt.args.isServer)
			result := tt.args.g.getResult()
			for k, v := range tt.want {
				if result.Labels.GetStringValue(k) != v {
					t.Errorf("Result()= %s:%s,Want %s:%s", k, result.Labels.GetStringValue(k), k, v)
				}
			}
		})
	}
}

func TestUrlMerge(t *testing.T) {
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
			want: "/test/123/*",
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
			name: "one segment with param",
			args: args{url: "/test?a=1"},
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
			if got := UrlMerge(tt.args.url); got != tt.want {
				t.Errorf("UrlMerge() = %v, want %v", got, tt.want)
			}
		})
	}
}
