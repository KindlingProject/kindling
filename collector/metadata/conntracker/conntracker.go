package conntracker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mdlayher/netlink"
	"github.com/ti-mo/conntrack"
	"github.com/ti-mo/netfilter"
)

const (
	// initializationTimeout is the timeout of new conntracker object
	initializationTimeout = time.Second * 10
	// eventChannelSize is the size of the channel for Consumer.
	eventChannelSize = 1024
	// workerNumber is the number of event worker
	workerNumber = 4
	// netlinkBufferSize is the size of receive buffer of Netlink
	netlinkBufferSize = 1024 * 1024
)

var globalConntracker *Conntracker
var once sync.Once

type Conntracker struct {
	mu           sync.RWMutex
	conn         *conntrack.Conn
	cache        *conntrackCache
	maxCacheSize int
	//TODO add machanism to remove timeout entry in time
	//evictTicker	*time.Ticker
}

func NewConntracker(conntrackMaxStateSize int) (*Conntracker, error) {
	var (
		err         error
		conntracker *Conntracker
	)

	ctx, cancel := context.WithTimeout(context.Background(), initializationTimeout)
	defer cancel()

	done := make(chan struct{})

	go func() {
		conntracker, err = newConntrackerOnce(conntrackMaxStateSize, workerNumber)
		done <- struct{}{}
	}()

	select {
	case <-done:
		return conntracker, err
	case <-ctx.Done():
		return nil, fmt.Errorf("could not initialize conntrack within: %s", initializationTimeout)
	}
}

func newConntrackerOnce(maxStateSize int, workerNumber uint8) (*Conntracker, error) {
	var errMsg error
	once.Do(func() {
		c, err := conntrack.Dial(nil)
		if err != nil {
			log.Fatal(err)
		}
		if err = c.SetOption(netlink.ListenAllNSID, true); err != nil {
			log.Printf("Warn: error setting up Netlink option ListenAllNSID: %s", err)
		}
		// if err = c.SetOption(netlink.NoENOBUFS, true); err != nil {
		// 	log.Printf("Warn: error setting up Netlink option NoENOBUFS: %s", err)
		// }
		// This will modify the net.core.rmem_default config, which is about 200KB by default, to
		// receive conntrack flows as many as possible before complaining about "no buffer" error.
		if err = c.SetReadBuffer(netlinkBufferSize); err != nil {
			log.Printf("Warn: error setting up Netlink read buffer: %s", err)
		}
		globalConntracker = &Conntracker{
			conn:         c,
			cache:        newConntrackCache(maxStateSize),
			maxCacheSize: maxStateSize,
		}
		flows, err := globalConntracker.conn.Dump()
		if err != nil {
			errMsg = fmt.Errorf("dump conntrack table failed:%w", err)
			return
		}

		if err = globalConntracker.initialConntrackTable(flows); err != nil {
			errMsg = fmt.Errorf("initial conntrack table failed:%w", err)
			return
		}

		if err = globalConntracker.poll(workerNumber); err != nil {
			errMsg = fmt.Errorf("poll conntrack update failed:%w", err)
			return
		}
		return
	})
	return globalConntracker, errMsg
}

// poll gets incremental update from conntrack table continuously
func (ctr *Conntracker) poll(workerNumber uint8) (err error) {
	evtCh := make(chan conntrack.Event, eventChannelSize)
	errCh, err := ctr.conn.Listen(evtCh, workerNumber, append(netfilter.GroupsCT))
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				log.Printf("Conntrack statistics: %v", ctr.cache.stats)
				log.Printf("Current cache size: %d", ctr.cache.Len())
			case err := <-errCh:
				log.Printf("error occured during receiving message from conntracker socket: %s", err)
			}
		}

	}()

	go func() {
		for event := range evtCh {
			ctr.updateCache(event.Flow)
		}
	}()

	return err
}

func (ctr *Conntracker) initialConntrackTable(flows []conntrack.Flow) error {
	for _, f := range flows {
		if !IsNAT(&f) {
			continue
		}
		ctr.cache.Add(&f)
	}
	return nil
}

func (ctr *Conntracker) updateCache(f *conntrack.Flow) bool {
	if !IsNAT(f) {
		return false
	}

	ctr.mu.Lock()
	defer ctr.mu.Unlock()

	return ctr.cache.Add(f)
}

func (ctr *Conntracker) GetDNATTuple(srcIP uint32, dstIP uint32, srcPort uint16, dstPort uint16, isUdp uint32) *IPTranslation {
	k := connKey{
		srcIP:   srcIP,
		srcPort: srcPort,
		dstIP:   dstIP,
		dstPort: dstPort,
		isUdp:   isUdp,
	}
	ctr.mu.RLock()
	defer ctr.mu.RUnlock()

	entry, ok := ctr.cache.Get(k)
	if !ok {
		return nil
	}

	return entry
}

func (ctr *Conntracker) GetDNATTupleWithString(srcIP string, dstIP string, srcPort uint16, dstPort uint16, isUdp uint32) *IPTranslation {
	sourceIP := StringToUint32(srcIP)
	destinationIP := StringToUint32(dstIP)
	if sourceIP == 0 || destinationIP == 0 {
		return nil
	}
	return ctr.GetDNATTuple(sourceIP, destinationIP, srcPort, dstPort, isUdp)
}

func (ctr *Conntracker) GetStats() map[string]int64 {
	m := map[string]int64{
		"cache_size":   int64(ctr.cache.Len()),
		"gets_times":   atomic.LoadInt64(&ctr.cache.stats.gets),
		"add_times":    atomic.LoadInt64(&ctr.cache.stats.add),
		"remove_times": atomic.LoadInt64(&ctr.cache.stats.remove),
		"evicts_total": atomic.LoadInt64(&ctr.cache.stats.evicts),
	}
	return m
}

func (ctr *Conntracker) Close() {
	ctr.conn.Close()
}
