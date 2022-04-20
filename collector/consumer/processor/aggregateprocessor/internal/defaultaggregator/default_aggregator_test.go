package defaultaggregator

import (
	"github.com/Kindling-project/kindling/collector/consumer/processor/aggregateprocessor/internal"
	"github.com/Kindling-project/kindling/collector/model"
	"sync"
	"testing"
	"time"
)

func TestConcurrentAggregator(t *testing.T) {
	aggKindMap := &AggregatedConfig{KindMap: map[string][]KindConfig{
		"duration": {
			{Kind: SumKind, OutputName: "duration_sum"},
			{Kind: CountKind, OutputName: "request_count"},
		},
	}}

	aggregator := NewDefaultAggregator(aggKindMap)
	labels := model.NewAttributeMap()
	labels.AddStringValue("key", "value")

	labelSelectors := internal.NewLabelSelectors(
		internal.LabelSelector{Name: "key", VType: internal.StringType},
	)

	wg := sync.WaitGroup{}
	wg.Add(1)
	stopCh := make(chan bool)

	var runLoop = 2000
	var duration int64 = 100
	go func() {
		for i := 0; i < runLoop; i++ {
			gaugeValues := []*model.Gauge{
				{"duration", duration},
			}
			gaugeGroup := model.NewGaugeGroup("testGauge", labels, 0, gaugeValues...)
			aggregator.Aggregate(gaugeGroup, labelSelectors)
			time.Sleep(time.Millisecond)
		}
		stopCh <- true
		wg.Done()
	}()

	wg.Add(1)
	var durationSum int64
	var requestCount int64
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				ret := aggregator.Dump()
				for _, g := range ret {
					t.Log(g)
					for _, v := range g.Values {
						if v.Name == "duration_sum" {
							durationSum += v.Value
						} else if v.Name == "request_count" {
							requestCount += v.Value
						}
					}

				}
			case <-stopCh:
				wg.Done()
				return
			}
		}
	}()
	wg.Wait()
	expectedDurationSum := int64(runLoop) * duration
	expectedRequestCount := int64(runLoop)
	if durationSum != expectedDurationSum || requestCount != expectedRequestCount {
		t.Errorf("request_count = %v, duration_sum := %v", requestCount, durationSum)
	}
}
