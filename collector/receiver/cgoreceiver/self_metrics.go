package cgoreceiver

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Kindling-project/kindling/collector/model/constnames"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var once sync.Once

const (
	eventReceivedMetric = "kindling_telemetry_cgoreceiver_events_total"
	channelSizeMetric   = "kindling_telemetry_cgoreceiver_channel_size"
)

func newSelfMetrics(meterProvider metric.MeterProvider, receiver *CgoReceiver) {
	once.Do(func() {
		meter := metric.Must(meterProvider.Meter("kindling"))
		meter.NewInt64CounterObserver(eventReceivedMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				for name, value := range receiver.stats.getStats() {
					result.Observe(value, attribute.String("name", name))
				}
			})
		meter.NewInt64GaugeObserver(channelSizeMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(int64(len(receiver.eventChannel)))
			})
	})
}

type eventCounter interface {
	add(name string, value int64)
	getStats() map[string]int64
}

type atomicInt64Counter struct {
	v int64
}

func (c *atomicInt64Counter) add(value int64) {
	atomic.AddInt64(&c.v, value)
}

func (c *atomicInt64Counter) get() int64 {
	return atomic.LoadInt64(&c.v)
}

type dynamicStats struct {
	stats map[string]*atomicInt64Counter
}

func newDynamicStats(subEvents []SubEvent) *dynamicStats {
	ret := &dynamicStats{
		stats: make(map[string]*atomicInt64Counter),
	}
	for _, event := range subEvents {
		var rawName string
		nameSegments := strings.Split(event.Name, "-")
		if len(nameSegments) > 1 {
			rawName = nameSegments[1]
		} else {
			rawName = nameSegments[0]
		}
		ret.stats[rawName] = &atomicInt64Counter{0}
	}
	ret.stats[constnames.OtherEvent] = &atomicInt64Counter{0}
	return ret
}

func (s *dynamicStats) add(name string, value int64) {
	c, ok := s.stats[name]
	if ok {
		c.add(value)
	} else {
		c = s.stats[constnames.OtherEvent]
		c.add(value)
	}
}

func (s *dynamicStats) getStats() map[string]int64 {
	ret := make(map[string]int64, len(s.stats))
	for k, v := range s.stats {
		ret[k] = v.get()
	}
	return ret
}
