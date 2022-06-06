package conntracker

import (
	"fmt"
	"github.com/Kindling-project/kindling/collector/metadata/conntracker/internal"
	"go.opentelemetry.io/otel/metric/global"
	"log"
	"net"
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
			cfg := &internal.Config{
				Enabled:                      config.Enabled,
				ProcRoot:                     config.ProcRoot,
				ConntrackInitTimeout:         config.ConntrackInitTimeout,
				ConntrackRateLimit:           config.ConntrackRateLimit,
				ConntrackMaxStateSize:        config.ConntrackMaxStateSize,
				EnableConntrackAllNamespaces: config.EnableConntrackAllNamespaces,
			}
			conntracker, err := internal.NewConntracker(cfg)
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
	conntracker internal.Conntracker
	cfg         *Config
}

func (ctr *NetlinkConntracker) GetDNATTupleWithString(srcIP string, dstIP string, srcPort uint16, dstPort uint16, isUdp uint32) *IPTranslation {
	conn := internal.ConnectionStats{
		Source: net.ParseIP(srcIP),
		SPort:  srcPort,
		Dest:   net.ParseIP(dstIP),
		DPort:  dstPort,
		Type:   internal.ConnectionType(isUdp),
	}
	return ctr.getDNATTuple(conn)
}

func (ctr *NetlinkConntracker) GetDNATTuple(srcIP uint32, dstIP uint32, srcPort uint16, dstPort uint16, isUdp uint32) *IPTranslation {
	conn := internal.ConnectionStats{
		Source: int32ToIp(srcIP),
		SPort:  srcPort,
		Dest:   int32ToIp(dstIP),
		DPort:  dstPort,
		Type:   internal.ConnectionType(isUdp),
	}
	return ctr.getDNATTuple(conn)
}

// getDNATTuple is a helper function for public methods with private parameter.
func (ctr *NetlinkConntracker) getDNATTuple(conn internal.ConnectionStats) *IPTranslation {
	ret := ctr.conntracker.GetTranslationForConn(conn)
	if ret == nil {
		return nil
	}
	// Check whether the result is DNAT or SNAT.
	if conn.Dest.Equal(ret.ReplSrcIP) && ret.ReplSrcPort == conn.DPort {
		// Most likely is SNAT which is not needed
		return nil
	}
	return &IPTranslation{
		ReplSrcIP:   ret.ReplSrcIP,
		ReplDstIP:   ret.ReplDstIP,
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
