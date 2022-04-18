package defaultaggregator

import (
	"github.com/Kindling-project/kindling/collector/consumer/processor/aggregateprocessor/internal"
	"github.com/Kindling-project/kindling/collector/model"
	"testing"
)

func TestRecord(t *testing.T) {
	aggKindMap := AggregatedConfig{KindMap: map[string][]KindConfig{
		"duration": {
			{Kind: SumKind, OutputName: "duration_sum"},
			{Kind: CountKind, OutputName: "request_count"},
			{Kind: MaxKind, OutputName: "duration_max"},
			{Kind: AvgKind, OutputName: "duration_avg"},
		},
		"last": {{Kind: LastKind, OutputName: "last"}},
	}}

	recorder := newValueRecorder("testRecorder", aggKindMap.KindMap)
	keys := internal.NewLabelKeys([]internal.LabelKey{
		{
			Name:  "stringKey",
			Value: "stringValue",
			VType: internal.StringType,
		},
		{
			Name:  "booleanKey",
			Value: "true",
			VType: internal.BooleanType,
		},
		{
			Name:  "intKey",
			Value: "100",
			VType: internal.IntType,
		},
	}...)

	for i := 0; i < 100; i++ {
		gaugeValues := []*model.Gauge{
			{"duration", 100},
			{"last", int64(i)},
		}
		recorder.Record(keys, gaugeValues, 0)
	}
	retGaugeGroup := recorder.dump()
	sumValue, _ := retGaugeGroup[0].GetGauge("duration_sum")
	if sumValue.Value != 10000 {
		t.Errorf("expected %v, got %v", 10000, sumValue)
	}
	countValue, _ := retGaugeGroup[0].GetGauge("request_count")
	if countValue.Value != 100 {
		t.Errorf("expected %v, got %v", 100, countValue)
	}
	maxValue, _ := retGaugeGroup[0].GetGauge("duration_max")
	if maxValue.Value != 100 {
		t.Errorf("expected %v, got %v", 100, maxValue)
	}
	avgValue, _ := retGaugeGroup[0].GetGauge("duration_avg")
	if avgValue.Value != 100 {
		t.Errorf("expected %v, got %v", 100, avgValue)
	}
	lastValue, _ := retGaugeGroup[0].GetGauge("last")
	if lastValue.Value != 99 {
		t.Errorf("expected %v, got %v", 99, lastValue)
	}
}
