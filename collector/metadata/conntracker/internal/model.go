package internal

import (
	"github.com/DataDog/datadog-agent/pkg/process/util"
	"net"
)

type IPTranslation struct {
	ReplSrcIP   net.IP
	ReplDstIP   net.IP
	ReplSrcPort uint16
	ReplDstPort uint16
}

// ConnectionStats stores statistics for a single connection.  Field order in the struct should be 8-byte aligned
type ConnectionStats struct {
	Source util.Address
	Dest   util.Address

	SPort uint16
	DPort uint16
	Type  uint8
}
