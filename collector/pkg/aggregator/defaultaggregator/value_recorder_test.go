package defaultaggregator

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/pkg/aggregator"
	"reflect"
	"testing"
)

func TestRecord(t *testing.T) {
	aggKindMap := AggregatedConfig{KindMap: map[string][]KindConfig{
		"duration": {
			{Kind: SumKind, OutputName: "duration_sum"},
			{Kind: CountKind, OutputName: "request_count"},
			{Kind: MaxKind, OutputName: "duration_max"},
			{Kind: AvgKind, OutputName: "duration_avg"},
			{Kind: HistogramKind, OutputName: "duration_histogram", ExplicitBoundaries: []int64{0, 100, 200, 500, 1000}},
		},
		"last": {{Kind: LastKind, OutputName: "last"}},
	}}

	recorder := newValueRecorder("testRecorder", aggKindMap.KindMap)
	keys := aggregator.NewLabelKeys([]aggregator.LabelKey{
		{
			Name:  "stringKey",
			Value: "stringValue",
			VType: aggregator.StringType,
		},
		{
			Name:  "booleanKey",
			Value: "true",
			VType: aggregator.BooleanType,
		},
		{
			Name:  "intKey",
			Value: "100",
			VType: aggregator.IntType,
		},
	}...)

	for i := 0; i < 100; i++ {
		gaugeValues := []*model.Gauge{
			model.NewIntGauge("duration", int64(100)),
			model.NewIntGauge("last", int64(i)),
		}
		recorder.Record(keys, gaugeValues, 0)
	}
	retGaugeGroup := recorder.dump()
	sumValue, _ := retGaugeGroup[0].GetGauge("duration_sum")
	if sumValue.GetInt().Value != 10000 {
		t.Errorf("sum expected %v, got %v", 10000, sumValue.GetInt().Value)
	}
	countValue, _ := retGaugeGroup[0].GetGauge("request_count")
	if countValue.GetInt().Value != 100 {
		t.Errorf("expected %v, got %v", 100, countValue.GetInt().Value)
	}
	maxValue, _ := retGaugeGroup[0].GetGauge("duration_max")
	if maxValue.GetInt().Value != 100 {
		t.Errorf("expected %v, got %v", 100, maxValue.GetInt().Value)
	}
	avgValue, _ := retGaugeGroup[0].GetGauge("duration_avg")
	if avgValue.GetInt().Value != 100 {
		t.Errorf("expected %v, got %v", 100, avgValue.GetInt().Value)
	}
	lastValue, _ := retGaugeGroup[0].GetGauge("last")
	if lastValue.GetInt().Value != 99 {
		t.Errorf("expected %v, got %v", 99, lastValue.GetInt().Value)
	}
	histogramValue, _ := retGaugeGroup[0].GetGauge("duration_histogram")
	expectedValue := &model.Histogram{
		Sum:                10000,
		Count:              100,
		ExplicitBoundaries: []int64{0, 100, 200, 500, 1000},
		BucketCounts:       []uint64{100, 100, 0, 0, 0},
	}
	if !reflect.DeepEqual(histogramValue.GetHistogram(), expectedValue) {
		t.Errorf("expected %+v, got %+v", expectedValue, histogramValue.GetHistogram())
	}
}
