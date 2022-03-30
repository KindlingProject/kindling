package conntracker

import (
	"net"
	"reflect"
	"testing"
)

func Test_int32ToIp(t *testing.T) {
	type args struct {
		i uint32
	}
	tests := []struct {
		name string
		args args
		want net.IP
	}{
		{"normal", args{67305985}, net.IP{1, 2, 3, 4}},
		{"zero", args{0}, net.IP{0, 0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := int32ToIp(tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("int32ToIp() = %v, want %v", got, tt.want)
			}
		})
	}
}
