package conntracker

import (
	"github.com/ti-mo/conntrack"
	"net"
	"strconv"
	"strings"
)

func IPToUInt32(ip net.IP) uint32 {
	b := ip.To4()
	if b == nil {
		return 0
	}
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func StringToUint32(ip string) uint32 {
	bytes := strings.Split(ip, ".")
	if len(bytes) <= 4 {
		return 0
	}
	b0, _ := strconv.Atoi(bytes[0])
	b1, _ := strconv.Atoi(bytes[1])
	b2, _ := strconv.Atoi(bytes[2])
	b3, _ := strconv.Atoi(bytes[3])

	return uint32(b0) | uint32(b1)<<8 | uint32(b2)<<16 | uint32(b3)<<24
}

func IsNAT(f *conntrack.Flow) bool {
	if len(f.TupleReply.IP.SourceAddress) == 0 ||
		len(f.TupleOrig.IP.SourceAddress) == 0 ||
		len(f.TupleOrig.IP.DestinationAddress) == 0 ||
		len(f.TupleReply.IP.DestinationAddress) == 0 ||
		f.TupleOrig.Proto.Protocol == 0 ||
		f.TupleReply.Proto.Protocol == 0 {
		return false
	}
	return !(f.TupleOrig.IP.SourceAddress.Equal(f.TupleReply.IP.DestinationAddress)) ||
		!(f.TupleOrig.IP.DestinationAddress).Equal(f.TupleReply.IP.SourceAddress) ||
		f.TupleOrig.Proto.SourcePort != f.TupleReply.Proto.DestinationPort ||
		f.TupleOrig.Proto.DestinationPort != f.TupleReply.Proto.SourcePort
}
