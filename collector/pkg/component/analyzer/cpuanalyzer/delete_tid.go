package cpuanalyzer

import (
	"sync"
	"time"
)

type tidDeleteQueue struct {
	queueMutex sync.Mutex
	queue      []deleteTid
}

type deleteTid struct {
	pid      uint32
	tid      uint32
	exitTime time.Time
}

func newTidDeleteQueue() *tidDeleteQueue {
	return &tidDeleteQueue{queue: make([]deleteTid, 0)}
}

func (dq *tidDeleteQueue) GetFront() *deleteTid {
	if len(dq.queue) > 0 {
		return &dq.queue[0]
	}
	return nil
}

func (dq *tidDeleteQueue) Push(elem deleteTid) {
	dq.queue = append(dq.queue, elem)
}

func (dq *tidDeleteQueue) Pop() {
	if len(dq.queue) > 0 {
		dq.queue = dq.queue[1:len(dq.queue)]
	}
}

//Add procexit tid
func (ca *CpuAnalyzer) AddTidToDeleteCache(curTime time.Time, pid uint32, tid uint32) {
	cacheElem := deleteTid{pid: pid, tid: tid, exitTime: curTime}
	ca.tidExpiredQueue.Push(cacheElem)
}

func (ca *CpuAnalyzer) TidDelete(interval time.Duration, expiredDuration time.Duration) {
	for {
		select {
		case <-time.After(interval):
			now := time.Now()
			func() {
				ca.lock.Lock()
				defer ca.lock.Unlock()
				ca.tidExpiredQueue.queueMutex.Lock()
				defer ca.tidExpiredQueue.queueMutex.Unlock()
				for {
					elem := ca.tidExpiredQueue.GetFront()
					if elem == nil {
						break
					}
					if elem.exitTime.Add(expiredDuration).Before(now) {
						tidEventsMap := ca.cpuPidEvents[elem.pid]
						if tidEventsMap == nil {
							ca.tidExpiredQueue.Pop()
							continue
						}
						ca.telemetry.Logger.Debugf("Delete expired thread... pid=%d, tid=%d", elem.pid, elem.tid)
						//fmt.Printf("Go Test: Delete expired thread... pid=%d, tid=%d\n", elem.pid, elem.tid)
						delete(tidEventsMap, elem.tid)
						ca.tidExpiredQueue.Pop()
					} else {
						break
					}
				}
			}()
		}
	}
}
