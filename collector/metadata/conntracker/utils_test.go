package conntracker

import (
	"net"
	"testing"
)

func TestIPToUInt32(t *testing.T) {
	type args struct {
		ip net.IP
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{name: "", args: args{ip: net.IPv4(1, 2, 3, 4)}, want: 67305985},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IPToUInt32(tt.args.ip); got != tt.want {
				t.Errorf("IPToUInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringToUint32(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{name: "normal", args: args{ip: "1.2.3.4"}, want: 67305985},
		{name: "illegal", args: args{ip: "1.2"}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringToUint32(tt.args.ip); got != tt.want {
				t.Errorf("StringToUint32() = %v, want %v", got, tt.want)
			}
		})
	}
}
