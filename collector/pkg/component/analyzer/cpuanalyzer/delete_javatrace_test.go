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
	mycfg := &Config{SegmentSize: 40,JavaTraceDeleteInterval:15,JavaTraceExpirationTime: 10}
	ca = &CpuAnalyzer{javaTraces: jt, telemetry: testTelemetry, cfg: mycfg}
	ca.javaTraceExpiredQueue = newJavaTraceDeleteQueue()
	ca.Start()
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
		time.Sleep(3 * time.Second)
	}
	time.Sleep(10 * timeDuration)

	if cnt != quitCnt {
		t.Fatalf("The number of javatraces that entering and exiting the queue is not equal! enterCount=%d, exitCount=%d\n", cnt, quitCnt)
	} else {
		t.Logf("All javatraces have cleaned normally. enterCount=%d, exitCount=%d\n",  cnt, quitCnt)
	}
	time.Sleep(10*time.Minute)

}
