package conntracker

import "github.com/Datadog/datadog-agent/pkg/network/netlink"

func NewNewConntracker() {
	netlink.NewConntracker()
}