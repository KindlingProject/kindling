package kafka

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
)

/*
      Request                                         Response
       /            \                                            /          \
fetch   produce                               fetch   produce
*/
func NewKafkaParser() *protocol.ProtocolParser {
	requestParser := protocol.CreatePkgParser(fastfailRequest(), parseRequest())
	requestParser.Add(fastfailRequestFetch(), parseRequestFetch())
	requestParser.Add(fastfailRequestProduce(), parseRequestProduce())
	requestParser.Add(fastfailRequestOther(), parseRequestOther())

	responseParser := protocol.CreatePkgParser(fastfailResponse(), parseResponse())
	responseParser.Add(fastfailResponseFetch(), parseResponseFetch())
	responseParser.Add(fastfailResponseProduce(), parseResponseProduce())
	responseParser.Add(fastfailResponseOther(), parseResponseOther())

	parser := protocol.NewProtocolParser(protocol.KAFKA, requestParser, responseParser, nil)
	return parser
}
