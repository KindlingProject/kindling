package conntracker

import (
	"github.com/ti-mo/conntrack"
	"net"
	"sync"
	"testing"
)

func TestConcurrentAdd(t *testing.T) {
	cache := newConntrackCache(10)
	numFlows := 10000

	var wg sync.WaitGroup
	var mutex sync.RWMutex
	addFlowsFunc := func() {
		// Create IPv4 flows
		for i := 1; i <= numFlows; i++ {
			f := conntrack.NewFlow(6, 0, net.IPv4(1, 2, 3, 4), net.IPv4(5, 6, 7, 8), 1234, uint16(i), 120, 0)
			mutex.Lock()
			cache.Add(&f)
			mutex.Unlock()
		}
		wg.Done()
	}

	getFlowsFunc := func() {
		for i := 1; i <= numFlows*2; i++ {
			k := &connKey{
				srcIP:   IPToUInt32(net.IPv4(1, 2, 3, 4)),
				srcPort: 1234,
				dstIP:   IPToUInt32(net.IPv4(5, 6, 7, 8)),
				dstPort: uint16(i),
				isUdp:   0,
			}
			mutex.RLock()
			cache.Get(k)
			mutex.RUnlock()
		}
		wg.Done()
	}

	workers := 4

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go addFlowsFunc()
		wg.Add(1)
		go getFlowsFunc()
	}

	wg.Wait()
}

func BenchmarkConntrackCache_Add(b *testing.B) {
	cache := newConntrackCache(10)
	for i := 0; i < b.N; i++ {
		numFlows := 1000000
		for i := 1; i <= numFlows; i++ {
			f := conntrack.NewFlow(6, 0, net.IPv4(1, 2, 3, 4), net.IPv4(5, 6, 7, 8), 1234, uint16(i), 120, 0)
			cache.Add(&f)
		}
	}
}
