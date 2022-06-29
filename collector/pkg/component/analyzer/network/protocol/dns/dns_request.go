package dns

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailDnsRequest() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) <= DNSHeaderSize || len(message.Data) > MaxMessageSize
	}
}

/**
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
		offset := message.Offset
		_, id := message.ReadUInt16(offset)
		_, flags := message.ReadUInt16(offset + 2)

		qr := (flags >> 15) & 0x1
		opcode := (flags >> 11) & 0xf
		rcode := flags & 0xf

		_, numOfQuestions := message.ReadUInt16(offset + 4)
		_, numOfAnswers := message.ReadUInt16(offset + 6)
		_, numOfAuth := message.ReadUInt16(offset + 8)
		_, numOfAddl := message.ReadUInt16(offset + 10)
		numOfRR := numOfQuestions + numOfAnswers + numOfAuth + numOfAddl

		/*
			QR: Request(0) Response(1)
			Kind of query in this message
				0	a standard query (QUERY)
				1	an inverse query (IQUERY)
				2	a server status request (STATUS)
				3-15 	reserved for future use

			Response code
				0	No error condition
				1 	Format error
				2 	Server failure
				3	Name Error
				4 	Not Implemented
				5 	Refused
				6-15 	Reserved for future use.
		*/
		if qr != 0 || opcode > 2 || rcode > 5 || numOfQuestions == 0 || numOfAnswers > 0 || numOfRR > MaxNumRR {
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
