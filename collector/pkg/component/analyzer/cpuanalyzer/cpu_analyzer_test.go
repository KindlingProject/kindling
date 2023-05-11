package cpuanalyzer

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	keys := make([]JavaTracesKey, 10)
	for i := 0; i < 10; i++ {
		offset := time.Duration(rand.Int63n(int64(10 * time.Second)))
		keys[i] = JavaTracesKey{
			TraceId:   strconv.Itoa(rand.Intn(1000000)),
			PidString: strconv.Itoa(rand.Intn(1000000)),
			StartTime: time.Now().Add(offset - 5*time.Second),
		}
	}
	tevent := &TransactionIdEvent{
		TraceId: "0",
		PidString: "1",
	}
	javaTraces := make(map[JavaTracesKey]*TransactionIdEvent)
	for _, key := range keys {
		javaTraces[key] = tevent
	}
	config:= NewDefaultConfig()
	ca := &CpuAnalyzer{
		javaTraces: javaTraces,
		cfg:           config,
	}
	ca.cleanerTicker = time.NewTicker(time.Duration(ca.cfg.JavaTraceDeleteInterval) * time.Second)
	go func() {
		for range ca.cleanerTicker.C {
			ca.lock.Lock()
			now := time.Now()
			for key:= range ca.javaTraces {
				if now.Sub(key.StartTime) > time.Duration(ca.cfg.JavaTraceExpirationTime)*time.Second {
					delete(ca.javaTraces, key)
					t.Log("已删除pid="+
					key.PidString+ "，当前时间："+
					time.Now().Truncate(time.Second).Format("15:04:05")+
					"，map剩余数量:"+strconv.Itoa(len(javaTraces)))
				}
			}
			ca.lock.Unlock()
		}
	}()
	time.Sleep(10*time.Minute)
}

func Test(t *testing.T){

}
