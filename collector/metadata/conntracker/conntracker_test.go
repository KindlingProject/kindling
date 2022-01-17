package conntracker

import (
	"fmt"
	"github.com/ti-mo/conntrack"
	"log"
	"testing"
	"time"
)

const maxStateSize = 131072

func TestConntracker_GetDNATTuple(t *testing.T) {
	c, err := conntrack.Dial(nil)
	if err != nil {
		log.Fatal(err)
	}
	conntracker := &Conntracker{
		conn:         c,
		cache:        newConntrackCache(maxStateSize),
		maxCacheSize: maxStateSize,
	}
	flows, err := conntracker.conn.Dump()
	fmt.Println("flow's number:", len(flows))
	if err != nil {
		fmt.Printf("initialConnAndDump failed: %v", err)
	}
	if err := conntracker.initialConntrackTable(flows); err != nil {
		fmt.Printf("initial conntrack table failed: %v", err)
	}
	fmt.Println("cachesize:", conntracker.cache.Len())

	if err := conntracker.poll(workerNumber); err != nil {
		fmt.Printf("poll conntrack table failed: %v", err)
	}

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			fmt.Println("cachesize:", conntracker.cache.Len())
		}
	}()
	for _, flow := range flows {
		k1, _ := tupleToKey(&flow.TupleOrig)
		k2, _ := tupleToKey(&flow.TupleReply)
		fmt.Println(conntracker.cache.cache.Get(k1))
		fmt.Println(conntracker.cache.cache.Get(k2))
	}
	select {}
}
