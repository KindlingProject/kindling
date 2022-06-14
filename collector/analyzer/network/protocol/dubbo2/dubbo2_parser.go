package dubbo2

import (
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
)

const (
	// Zero : byte zero
	Zero = byte(0x00)

	// magic header
	MagicHigh = byte(0xda)
	MagicLow  = byte(0xbb)

	// message flag.
	FlagRequest = byte(0x80)
	FlagTwoWay  = byte(0x40)
	FlagEvent   = byte(0x20) // for heartbeat
	SerialMask  = 0x1f

	AsciiLow     = byte(0x20)
	AsciiHigh    = byte(0x7e)
	AsciiReplace = byte(0x2e) // .
)

func NewDubbo2Parser() *protocol.ProtocolParser {
	requestParser := protocol.CreatePkgParser(fastfailDubbo2Request(), parseDubbo2Request())
	responseParser := protocol.CreatePkgParser(fastfailDubbo2Response(), parseDubbo2Response())
	return protocol.NewProtocolParser(protocol.DUBBO2, requestParser, responseParser, dubbo2Pair())
}

func dubbo2Pair() protocol.PairMatch {
	return func(requests []*protocol.PayloadMessage, response *protocol.PayloadMessage) int {
		for i, request := range requests {
			if request.GetIntAttribute(constlabels.Dubbo2RpcRequestId) == response.GetIntAttribute(constlabels.Dubbo2RpcRequestId) {
				return i
			}
		}
		return -1
	}
}

func getRcpRequestId(data []byte) int64 {
	return int64(uint64(data[4])<<56 | uint64(data[5])<<48 | uint64(data[6])<<40 | uint64(data[7])<<32 |
		uint64(data[8])<<24 | uint64(data[9])<<16 | uint64(data[10])<<8 | uint64(data[11]))
}

/**
  Get the ascii readable string, replace other value to '.', like wireshark.
*/
func getAsciiString(data []byte) string {
	length := len(data)
	if length == 0 {
		return ""
	}

	newData := make([]byte, length)
	for i := 0; i < length; i++ {
		if data[i] > AsciiHigh || data[i] < AsciiLow {
			newData[i] = AsciiReplace
		} else {
			newData[i] = data[i]
		}
	}
	return string(newData)
}
