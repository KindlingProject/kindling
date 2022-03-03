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
		{
			name: "GrpcTopology",
			args: args{
				g:        newGauges(newInnerGauges(false)),
				protocol: grpc,
				isServer: false,
			},
			want: map[string]string{
				"status_code": "200",
			},
		},
		{
			name: "GrpcEntity",
			args: args{
				g:        newGauges(newInnerGauges(true)),
				protocol: grpc,
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
