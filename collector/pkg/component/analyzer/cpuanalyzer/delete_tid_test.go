package cpuanalyzer

import (
	"testing"
	"time"
)

func TestDeleteQueue(t *testing.T) {

	cpupidEvents := make(map[uint32]map[uint32]*TimeSegments, 100000)
	ca := &CpuAnalyzer{cpuPidEvents: cpupidEvents}

	ca.tidExpiredQueue = newTidDeleteQueue()

	go ca.TidDelete(3*time.Second, 4*time.Second)
	for i := 0; i < 10; i++ {
		ca.AddTidToDeleteCache(time.Now(), uint32(i), uint32(i)+5)
		t.Logf("pid=%d, tid=%d enter\n", uint32(i), uint32(i)+5)
		time.Sleep(1 * time.Second)
	}

}
