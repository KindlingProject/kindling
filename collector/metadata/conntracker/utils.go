package conntracker

import (
	"net"
	"strconv"
	"strings"
)

func int32ToIp(i uint32) net.IP {
	ip := make(net.IP, net.IPv4len)
	ip[3] = byte(i >> 24)
	ip[2] = byte(i >> 16)
	ip[1] = byte(i >> 8)
	ip[0] = byte(i)
	return ip
}

func IPToUInt32(ip net.IP) uint32 {
	b := ip.To4()
	if b == nil {
		return 0
	}
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func StringToUint32(ip string) uint32 {
	bytes := strings.Split(ip, ".")
	if len(bytes) < 4 {
		return 0
	}
	b0, _ := strconv.Atoi(bytes[0])
	b1, _ := strconv.Atoi(bytes[1])
	b2, _ := strconv.Atoi(bytes[2])
	b3, _ := strconv.Atoi(bytes[3])

	return uint32(b0) | uint32(b1)<<8 | uint32(b2)<<16 | uint32(b3)<<24
}
