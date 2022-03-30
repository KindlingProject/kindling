package conntracker

import "net"

type IPTranslation struct {
	ReplSrcIP   net.IP
	ReplDstIP   net.IP
	ReplSrcPort uint16
	ReplDstPort uint16
}
