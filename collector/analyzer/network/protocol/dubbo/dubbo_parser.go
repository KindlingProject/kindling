package dubbo

import "github.com/Kindling-project/kindling/collector/analyzer/network/protocol"

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

func NewDubboParser() *protocol.ProtocolParser {
	requestParser := protocol.CreatePkgParser(fastfailDubboRequest(), parseDubboRequest())
	responseParser := protocol.CreatePkgParser(fastfailDubboResponse(), parseDubboResponse())
	return protocol.NewProtocolParser(protocol.DUBBO, requestParser, responseParser, nil)
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
