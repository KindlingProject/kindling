package udsreceiver

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"os"
	"sync"
	"testing"
	"time"
)

func runTest(counter eventCounter, workerNum int, loopNum int) {
	wg := sync.WaitGroup{}
	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		go func() {
			runRecordCounter(loopNum, counter)
			wg.Done()
		}()
	}
	wg.Wait()
}

var eventLists = []string{"read", "write", "readv", "writev", "sendto", "recvfrom", "sendmsg", "recvmsg",
	"grpc_uprobe", "tcp_close", "tcp_rcv_established", "tcp_drop", "tcp_retransmit_skb", "another_event"}

func runRecordCounter(loopNum int, counter eventCounter) {
	for i := 0; i < loopNum; i++ {
		for _, name := range eventLists {
			counter.add(name, 1)
		}
	}
}

func assertTest(t *testing.T, counter eventCounter, workerNum int, loopNum int) {
	runTest(counter, workerNum, loopNum)
	expectedNum := workerNum * loopNum
	for _, value := range counter.getStats() {
		if value != int64(expectedNum) {
			t.Errorf("The result is expected to be %d, but got %d", expectedNum, value)
		}
	}
}

func TestCounterMutexMap(t *testing.T) {
	counter := &mutexMap{m: make(map[string]int64)}
	assertTest(t, counter, 5, 100000)
}

func TestCounterRwMutexMap(t *testing.T) {
	counter := &rwMutexMap{m: make(map[string]int64)}
	assertTest(t, counter, 5, 100000)
}

func TestCounterIntCombination(t *testing.T) {
	counter := &stats{}
	assertTest(t, counter, 5, 100000)
}

func BenchmarkCounterMutexMap(b *testing.B) {
	counter := &mutexMap{m: make(map[string]int64)}
	initOtelCounterObserver(counter)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runTest(counter, 5, 1000)
	}
}

func BenchmarkCounterRwMutexMap(b *testing.B) {
	counter := &rwMutexMap{m: make(map[string]int64)}
	initOtelCounterObserver(counter)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runTest(counter, 5, 1000)
	}
}

func BenchmarkCounterIntCombination(b *testing.B) {
	counter := &stats{}
	initOtelCounterObserver(counter)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runTest(counter, 5, 1000)
	}
}

func BenchmarkCounterOtelCounter(b *testing.B) {
	counter := newOtelRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runTest(counter, 5, 1000)
	}
}

// It's not practical to implement with sync.Map,
// because you still need to lock the value when increasing it.
// type syncMap struct {
// 	m sync.Map
// 	mutex sync.Mutex
// }

type mutexMap struct {
	m     map[string]int64
	mutex sync.Mutex
}

func (m *mutexMap) add(name string, value int64) {
	m.mutex.Lock()
	v := m.m[name]
	m.m[name] = v + value
	m.mutex.Unlock()
}
func (m *mutexMap) getStats() map[string]int64 {
	m.mutex.Lock()
	ret := make(map[string]int64, len(m.m))
	for k, v := range m.m {
		ret[k] = v
	}
	m.mutex.Unlock()
	return ret
}

type rwMutexMap struct {
	m     map[string]int64
	mutex sync.RWMutex
}

func (m *rwMutexMap) add(name string, value int64) {
	m.mutex.Lock()
	v := m.m[name]
	m.m[name] = v + value
	m.mutex.Unlock()
}
func (m *rwMutexMap) getStats() map[string]int64 {
	m.mutex.RLock()
	ret := make(map[string]int64, len(m.m))
	for k, v := range m.m {
		ret[k] = v
	}
	m.mutex.RUnlock()
	return ret
}

type otelRecorder struct {
	otelCounter metric.Int64Counter
}

func newOtelRecorder() *otelRecorder {
	meter := initOpentelemetry()
	return &otelRecorder{otelCounter: meter.NewInt64Counter("event_counter_total")}
}
func (m *otelRecorder) add(name string, value int64) {
	m.otelCounter.Add(context.Background(), value, attribute.String("name", name))
}
func (m *otelRecorder) getStats() map[string]int64 {
	return nil
}

func initOpentelemetry() metric.MeterMust {
	devNullWriter, _ := os.Open(os.DevNull)
	exp, _ := stdoutmetric.New(stdoutmetric.WithWriter(devNullWriter))

	cont := controller.New(
		otelprocessor.NewFactory(selector.NewWithInexpensiveDistribution(), exp),
		controller.WithExporter(exp),
		controller.WithCollectPeriod(100*time.Millisecond),
	)
	_ = cont.Start(context.Background())
	return metric.Must(cont.Meter("kindling"))
}

func initOtelCounterObserver(counter eventCounter) {
	meter := initOpentelemetry()
	meter.NewInt64CounterObserver("event_counter_total",
		func(ctx context.Context, result metric.Int64ObserverResult) {
			metrics := counter.getStats()
			for name, value := range metrics {
				result.Observe(value, attribute.String("name", name))
			}
		})
}
