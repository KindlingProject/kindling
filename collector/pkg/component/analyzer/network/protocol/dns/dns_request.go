package dns

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailDnsRequest() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) <= DNSHeaderSize
	}
}

/*
Header format

	  0  1  2  3  4  5  6  7  8  9  A  B  C  D  E  F
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	|                      ID                       |
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	|QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	|                    QDCOUNT                    |
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	|                    ANCOUNT                    |
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	|                    NSCOUNT                    |
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	|                    ARCOUNT                    |
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
*/
func parseDnsRequest() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		if message.Protocol == model.L4Proto_TCP {
			message.Offset += 2
		}
		offset := message.Offset
		id, _ := message.ReadUInt16(offset)

		numOfQuestions, _ := message.ReadUInt16(offset + 4)
		if numOfQuestions == 0 {
			return false, true
		}
		domain, err := readQuery(message, numOfQuestions)
		if err != nil {
			return false, true
		}
		message.AddIntAttribute(constlabels.DnsId, int64(id))
		message.AddStringAttribute(constlabels.DnsDomain, domain)
		return true, true
	}
}
