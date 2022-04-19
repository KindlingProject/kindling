package dubbo

import "github.com/Kindling-project/kindling/collector/analyzer/network/protocol"

const (
	// Zero : byte zero
	Zero = byte(0x00)

	// magic header
	MAGIC_HIGH = byte(0xda)
	MAGIC_LOW  = byte(0xbb)

	// message flag.
	FLAG_REQUEST = byte(0x80)
	FLAG_TWOWAY  = byte(0x40)
	FLAG_EVENT   = byte(0x20) // for heartbeat
	SERIAL_MASK  = 0x1f

	ASCII_LOW     = byte(0x20)
	ASCII_HIGH    = byte(0x7e)
	ASCII_REPLACE = byte(0x2e) // .
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
		if data[i] > ASCII_HIGH || data[i] < ASCII_LOW {
			newData[i] = ASCII_REPLACE
		} else {
			newData[i] = data[i]
		}
	}
	return string(newData)
}
