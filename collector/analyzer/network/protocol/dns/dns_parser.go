package dns

import (
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
)

const (
	DNSHeaderSize  = 12
	MaxNumRR       = 25
	MaxMessageSize = 512

	maxDomainNameWireOctets         = 255 // See RFC 1035 section 2.3.4
	maxCompressionPointers          = (maxDomainNameWireOctets+1)/2 - 2
	maxDomainNamePresentationLength = 61*4 + 1 + 63*4 + 1 + 63*4 + 1 + 63*4 + 1
)

/**
https://www.rfc-editor.org/rfc/rfc1035

  +---------------------+
  |        Header       |
  +---------------------+
  |       Question      | the question for the name server
  +---------------------+
  |        Answer       | RRs answering the question
  +---------------------+
  |      Authority      | RRs pointing toward an authority
  +---------------------+
  |      Additional     | RRs holding additional information
  +---------------------+
*/
func NewDnsParser() *protocol.ProtocolParser {
	requestParser := protocol.CreatePkgParser(fastfailDnsRequest(), parseDnsRequest())
	responseParser := protocol.CreatePkgParser(fastfailDnsResponse(), parseDnsResponse())

	return protocol.NewProtocolParser(protocol.DNS, requestParser, responseParser, dnsPair())
}

func dnsPair() protocol.PairMatch {
	return func(requests []*protocol.PayloadMessage, response *protocol.PayloadMessage) int {
		for i, request := range requests {
			if request.GetIntAttribute(constlabels.DnsId) == response.GetIntAttribute(constlabels.DnsId) &&
				request.GetStringAttribute(constlabels.DnsDomain) == response.GetStringAttribute(constlabels.DnsDomain) {
				return i
			}
		}
		return -1
	}
}

func readQuery(message *protocol.PayloadMessage, queryCount uint16) (domain string, err error) {
	var name string
	offset := message.Offset + 12

	for i := 0; i < int(queryCount); i++ {
		if message.IsComplete() {
			return "", protocol.ErrEof
		}

		/*
			uint16 qname
			uint16 qtype
			uint16 qclass
		*/
		name, offset, err = unpackDomainName(message.Data, offset)
		if err != nil || offset >= len(message.Data) {
			return "", protocol.ErrMessageInvalid
		}
		if len(domain) == 0 {
			domain = name
		}
		offset += 4
	}
	message.Offset = offset
	return domain, nil
}
