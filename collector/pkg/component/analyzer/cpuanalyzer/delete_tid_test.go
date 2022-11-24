package cpuanalyzer

import (
	"strconv"
	"testing"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/component"
)

var (
	visit    []deleteTid
	ca       *CpuAnalyzer
	exitTid  map[uint32]int
	enterCnt int
	exitCnt  int
)

func TestDeleteQueue(t *testing.T) {

	cpupidEvents := make(map[uint32]map[uint32]*TimeSegments, 100000)
	testTelemetry := component.NewTelemetryManager().GetGlobalTelemetryTools()
	ca = &CpuAnalyzer{cpuPidEvents: cpupidEvents, telemetry: testTelemetry}

	ca.tidExpiredQueue = newTidDeleteQueue()

	visit = make([]deleteTid, 0)
	exitTid = make(map[uint32]int, 0)

	go ca.TidDelete(5*time.Second, 4*time.Second)
	go CheckQueueLoop(t)
	for i := 0; i < 20; i++ {

		ev := new(CpuEvent)
		curTime := time.Now()
		ev.EndTime = uint64(curTime.Add(time.Second).Nanosecond())
		ev.StartTime = uint64(curTime.Nanosecond())

		//check tid which exist in queue but not in the map
		if i%4 != 0 {
			PutElemToMap(uint32(i), uint32(i)+5, "threadname"+strconv.Itoa(i+100), ev)
		}

		var queueLen int

		func() {
			ca.tidExpiredQueue.queueMutex.Lock()
			defer ca.tidExpiredQueue.queueMutex.Unlock()
			queueLen = len(ca.tidExpiredQueue.queue)

			cacheElem := deleteTid{uint32(i), uint32(i) + 5, curTime.Add(time.Second)}
			ca.tidExpiredQueue.Push(cacheElem)
			visit = append(visit, cacheElem)
			if len(ca.tidExpiredQueue.queue) != queueLen+1 {
				t.Errorf("the length of queue is not added, expection: %d but: %d\n", queueLen+1, len(ca.tidExpiredQueue.queue))
			}
		}()

		t.Logf("pid=%d, tid=%d enter time=%s\n", uint32(i), uint32(i)+5, curTime.Format("2006-01-02 15:04:05"))
		enterCnt++
		time.Sleep(3 * time.Second)
	}
	time.Sleep(10 * time.Second)

	if enterCnt != exitCnt {
		t.Fatalf("The number of threads that entering and exiting the queue is not equal! enterCount=%d, exitCount=%d\n", enterCnt, exitCnt)
	} else {
		t.Logf("All threads have exited normally. enterCount=%d, exitCount=%d\n", enterCnt, exitCnt)
	}

}

func CheckQueueLoop(t *testing.T) {
	for {
		select {
		case <-time.After(time.Second * 3):
			func() {
				ca.tidExpiredQueue.queueMutex.Lock()
				defer ca.tidExpiredQueue.queueMutex.Unlock()
				queueLen := len(ca.tidExpiredQueue.queue)
				curTime := time.Now()
				for i := 0; i < len(visit); i++ {
					tmpv := visit[i]
					var j int
					for j = 0; j < queueLen; j++ {
						tmpq := ca.tidExpiredQueue.queue[j]
						if tmpv.tid == tmpq.tid {
							if curTime.After(tmpq.exitTime.Add(12 * time.Second)) {
								t.Errorf("there is a expired threads that is not deleted. pid=%d, tid=%d, exitTime=%s\n", tmpv.pid, tmpv.tid, tmpv.exitTime.Format("2006-01-02 15:04:05"))
							}
							break
						}
					}

					if _, exist := exitTid[tmpv.tid]; j >= queueLen && !exist {
						exitTid[tmpv.tid] = 1
						exitCnt++
						t.Logf("pid=%d, tid=%d exit time=%s\n", tmpv.pid, tmpv.tid, curTime.Format("2006-01-02 15:04:05"))
					}
				}
			}()
		}
	}
}

func PutElemToMap(pid uint32, tid uint32, threadName string, event TimedEvent) {

	tidCpuEvents, exist := ca.cpuPidEvents[pid]
	if !exist {
		tidCpuEvents = make(map[uint32]*TimeSegments)
		ca.cpuPidEvents[pid] = tidCpuEvents
	}

	newTimeSegments := &TimeSegments{
		Pid:      pid,
		Tid:      tid,
		BaseTime: event.StartTimestamp() / nanoToSeconds,
		Segments: NewCircleQueue(40),
	}
	for i := 0; i < 40; i++ {
		segment := newSegment(pid, tid, threadName,
			(newTimeSegments.BaseTime+uint64(i))*nanoToSeconds,
			(newTimeSegments.BaseTime+uint64(i+1))*nanoToSeconds)
		newTimeSegments.Segments.UpdateByIndex(i, segment)
	}

	endOffset := int(event.EndTimestamp()/nanoToSeconds - newTimeSegments.BaseTime)

	for i := 0; i <= endOffset && i < 40; i++ {
		val := newTimeSegments.Segments.GetByIndex(i)
		segment := val.(*Segment)
		segment.putTimedEvent(event)
		segment.IsSend = 0
	}

	tidCpuEvents[tid] = newTimeSegments
}
