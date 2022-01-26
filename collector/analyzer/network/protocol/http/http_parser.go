package http

import (
	"strings"

	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
)

func NewHttpParser() *protocol.ProtocolParser {
	requestParser := protocol.CreatePkgParser(fastfailHttpRequest(), parseHttpRequest())
	responseParser := protocol.CreatePkgParser(fastfailHttpResponse(), parseHttpResponse())

	return protocol.NewProtocolParser(protocol.HTTP, requestParser, responseParser, nil)
}

/*
Requet-Line\r\n
Key:Value\r\n
...
Key:Value\r\n
\r\n
Data
*/
func parseHeaders(message *protocol.PayloadMessage) map[string]string {
	header := make(map[string]string)

	from, data := message.ReadUntilCRLF(0)
	if data == nil {
		return header
	}
	for {
		from, data = message.ReadUntilCRLF(from)
		if data == nil {
			return header
		}
		if position := strings.Index(string(data), ":"); position > 0 && position < len(data)-1 {
			header[strings.ToLower(string(data[0:position]))] = string(data[position+1])
			continue
		}
		return header
	}
}
