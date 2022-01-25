package conntracker

import (
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/ti-mo/conntrack"
	"golang.org/x/sys/unix"
	"net"
	"sync/atomic"
)

type connKey struct {
	srcIP   uint32
	srcPort uint16

	dstIP   uint32
	dstPort uint16

	// the transport protocol of the connection, using the same values as specified in the agent payload.
	isUdp uint32
}

type IPTranslation struct {
	ReplSrcIP   net.IP
	ReplDstIP   net.IP
	ReplSrcPort uint16
	ReplDstPort uint16
}

type conntrackCache struct {
	//TODO replace simplelru with groupcache
	cache *simplelru.LRU
	stats struct {
		add     int64
		gets    int64
		remove  int64
		evicts  int64
		zombies int64 //timeout entry
	}
	//TODO Add logic to periodically clear timeout entries to lower the memory usage
	//timerQueue	*list.List
	//timeout		time.Duration
}

func newConntrackCache(maxStateSize int) *conntrackCache {
	c := &conntrackCache{}
	c.cache, _ = simplelru.NewLRU(maxStateSize, func(key, value interface{}) {
		atomic.AddInt64(&c.stats.evicts, 1)
	})

	return c
}

func (cc *conntrackCache) Get(k connKey) (*IPTranslation, bool) {
	atomic.AddInt64(&cc.stats.gets, 1)
	entry, ok := cc.cache.Get(k)
	if !ok {
		return nil, false
	}
	t := entry.(*IPTranslation)
	return t, true
}

func (cc *conntrackCache) Add(f *conntrack.Flow) bool {
	defer func() {
		atomic.AddInt64(&cc.stats.add, 1)
	}()
	AddToCache := func(a, b *conntrack.Tuple) bool {
		key, ok := tupleToKey(a)
		if !ok {
			return false
		}
		value := tupleToIPTranslation(b)
		cc.cache.Add(key, value)
		return true
	}
	return AddToCache(&f.TupleOrig, &f.TupleReply) && AddToCache(&f.TupleReply, &f.TupleOrig)
}

func (cc *conntrackCache) Remove(k connKey) bool {
	if cc.cache.Remove(k) {
		atomic.AddInt64(&cc.stats.remove, 1)
		return true
	}
	return false
}

func (cc *conntrackCache) Len() int {
	return cc.cache.Len()
}

func tupleToKey(tuple *conntrack.Tuple) (connKey, bool) {
	k := connKey{
		srcIP:   IPToUInt32(tuple.IP.SourceAddress),
		dstIP:   IPToUInt32(tuple.IP.DestinationAddress),
		srcPort: tuple.Proto.SourcePort,
		dstPort: tuple.Proto.DestinationPort,
	}

	proto := tuple.Proto.Protocol
	switch proto {
	case unix.IPPROTO_TCP:
		k.isUdp = 0
	case unix.IPPROTO_UDP:
		k.isUdp = 1
	default:
		return k, false
	}

	return k, true
}

func tupleToIPTranslation(tuple *conntrack.Tuple) *IPTranslation {
	return &IPTranslation{
		ReplSrcIP:   tuple.IP.SourceAddress,
		ReplDstIP:   tuple.IP.DestinationAddress,
		ReplSrcPort: tuple.Proto.SourcePort,
		ReplDstPort: tuple.Proto.DestinationPort,
	}
}
