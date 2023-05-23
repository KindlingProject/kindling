package cpuanalyzer

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/component"
)

var (
	cnt         int
	quitCnt     int
)


func TestJavaTraceDeleteQueue(t *testing.T) {

	jt := make(map[string]*TransactionIdEvent, 100000)
	testTelemetry := component.NewTelemetryManager().GetGlobalTelemetryTools()
	mycfg := &Config{SegmentSize: 40}
	ca = &CpuAnalyzer{javaTraces: jt, telemetry: testTelemetry, cfg: mycfg}
	ca.javaTraceExpiredQueue = newJavaTraceDeleteQueue()
	expiredDuration := time.Second * 1
	interval := time.Second *1
	go func () {
		for {
			select {
			case <-ca.stopProfileChan:
				return
			case <-time.After(interval):
				now := time.Now()
				func() {
					ca.javaTraceExpiredQueue.queueMutex.Lock()
					defer ca.javaTraceExpiredQueue.queueMutex.Unlock()
					for {
						val := ca.javaTraceExpiredQueue.GetFront()
						if val == nil {
							break
						}
						if val.enterTime.Add(expiredDuration).After(now) {
							break
						}
						//Delete expired threads (current_time >= thread_exit_time + interval_time).
						func() {
							ca.lock.Lock()
							defer ca.lock.Unlock()
							event := ca.javaTraces[val.key]
							if event == nil {
								ca.javaTraceExpiredQueue.Pop()
							} else {
								t.Logf("Delete expired thread... pid=%s, tid=%s", event.PidString, event.TraceId)
								delete(ca.javaTraces, val.key)
								quitCnt++;
								ca.javaTraceExpiredQueue.Pop()
							}
						}()
					}
				}()
			}
		}
	}()
	for i := 0; i < 20; i++ {

		ev := new(TransactionIdEvent)
		ev.TraceId = strconv.Itoa(rand.Intn(10000))
		ev.PidString = strconv.Itoa(rand.Intn(10000))
		ev.IsEntry = 1
		key:= ev.TraceId + ev.PidString
		ca.javaTraces[key] = ev
		val := new(deleteVal)
		val.key = ev.TraceId+ev.PidString
		val.enterTime = time.Now()
		ca.javaTraceExpiredQueue.Push(*val)
		t.Logf("pid=%s, tid=%s enter time=%s\n",ev.PidString, ev.TraceId, val.enterTime.Format("2006-01-02 15:04:05.000"))
		cnt++
		time.Sleep(3 * timeDuration)
	}
	time.Sleep(10 * timeDuration)

	if cnt != quitCnt {
		t.Fatalf("The number of javatraces that entering and exiting the queue is not equal! enterCount=%d, exitCount=%d\n", cnt, quitCnt)
	} else {
		t.Logf("All javatraces have exited normally. enterCount=%d, exitCount=%d\n",  cnt, quitCnt)
	}

	time.Sleep(10*time.Minute)

}
