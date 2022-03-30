package conntracker

import (
	"net"
)

func int32ToIp(i uint32) net.IP {
	ip := make(net.IP, net.IPv4len)
	ip[3] = byte(i >> 24)
	ip[2] = byte(i >> 16)
	ip[1] = byte(i >> 8)
	ip[0] = byte(i)
	return ip
}
