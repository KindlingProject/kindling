package conntracker

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/ebpf"
	"github.com/DataDog/datadog-agent/pkg/network"
	datadogcfg "github.com/DataDog/datadog-agent/pkg/network/config"
	"github.com/DataDog/datadog-agent/pkg/network/netlink"
	"github.com/DataDog/datadog-agent/pkg/process/util"
	"go.opentelemetry.io/otel/metric/global"
	"log"
	"sync"
)

type Conntracker interface {
	GetDNATTupleWithString(srcIP string, dstIP string, srcPort uint16, dstPort uint16, isUdp uint32) *IPTranslation
	GetDNATTuple(srcIP uint32, dstIP uint32, srcPort uint16, dstPort uint16, isUdp uint32) *IPTranslation
	GetStats() map[string]int64
}

var singletonConntracker Conntracker
var errMessage error
var once sync.Once

func NewConntracker(config *Config) (Conntracker, error) {
	once.Do(func() {
		if config == nil || !config.Enabled {
			log.Printf("Conntracker is not enabled.")
			singletonConntracker = NewNoopConntracker(config)
		} else {
			cfg := &datadogcfg.Config{
				Config: ebpf.Config{
					ProcRoot: config.ProcRoot,
				},
				ConntrackInitTimeout:         config.ConntrackInitTimeout,
				ConntrackRateLimit:           config.ConntrackRateLimit,
				ConntrackMaxStateSize:        config.ConntrackMaxStateSize,
				EnableConntrackAllNamespaces: config.EnableConntrackAllNamespaces,
			}

			conntracker, err := netlink.NewConntracker(cfg)
			if err != nil {
				errMessage = fmt.Errorf("failed to create conntracker: %w", err)
				singletonConntracker = NewNoopConntracker(config)
			} else {
				singletonConntracker = &NetlinkConntracker{
					conntracker: conntracker,
					cfg:         config,
				}
			}
		}
		newSelfMetrics(global.GetMeterProvider(), singletonConntracker)
	})

	return singletonConntracker, errMessage
}

type NetlinkConntracker struct {
	conntracker netlink.Conntracker
	cfg         *Config
}

func (ctr *NetlinkConntracker) GetDNATTupleWithString(srcIP string, dstIP string, srcPort uint16, dstPort uint16, isUdp uint32) *IPTranslation {
	conn := network.ConnectionStats{
		Source: util.AddressFromString(srcIP),
		SPort:  srcPort,
		Dest:   util.AddressFromString(dstIP),
		DPort:  dstPort,
		Type:   network.ConnectionType(isUdp),
	}

	ret := ctr.conntracker.GetTranslationForConn(conn)
	return &IPTranslation{
		ReplSrcIP:   ret.ReplSrcIP.Bytes(),
		ReplDstIP:   ret.ReplDstIP.Bytes(),
		ReplSrcPort: ret.ReplSrcPort,
		ReplDstPort: ret.ReplDstPort,
	}
}

func (ctr *NetlinkConntracker) GetDNATTuple(srcIP uint32, dstIP uint32, srcPort uint16, dstPort uint16, isUdp uint32) *IPTranslation {
	conn := network.ConnectionStats{
		Source: util.AddressFromNetIP(int32ToIp(srcIP)),
		SPort:  srcPort,
		Dest:   util.AddressFromNetIP(int32ToIp(dstIP)),
		DPort:  dstPort,
		Type:   network.ConnectionType(isUdp),
	}
	ret := ctr.conntracker.GetTranslationForConn(conn)
	return &IPTranslation{
		ReplSrcIP:   ret.ReplSrcIP.Bytes(),
		ReplDstIP:   ret.ReplDstIP.Bytes(),
		ReplSrcPort: ret.ReplSrcPort,
		ReplDstPort: ret.ReplDstPort,
	}
}

func (ctr *NetlinkConntracker) GetStats() map[string]int64 {
	ret := ctr.conntracker.GetStats()
	ret["cache_max_size"] = int64(ctr.cfg.ConntrackMaxStateSize)
	return ret
}

type NoopConntracker struct {
	cfg   *Config
	stats map[string]int64
}

func NewNoopConntracker(cfg *Config) *NoopConntracker {
	return &NoopConntracker{
		cfg:   cfg,
		stats: make(map[string]int64),
	}
}

func (ctr *NoopConntracker) GetDNATTupleWithString(srcIP string, dstIP string, srcPort uint16, dstPort uint16, isUdp uint32) *IPTranslation {
	return nil
}

func (ctr *NoopConntracker) GetDNATTuple(srcIP uint32, dstIP uint32, srcPort uint16, dstPort uint16, isUdp uint32) *IPTranslation {
	return nil
}

func (ctr *NoopConntracker) GetStats() map[string]int64 {
	return ctr.stats
}
