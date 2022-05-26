package defaultaggregator

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/pkg/aggregator"
	"sync"
)

type valueRecorder struct {
	name string
	// aggValuesMap is responsible for its own thread-safe access.
	labelValues sync.Map
	aggKindMap  map[string][]KindConfig
}

func newValueRecorder(recorderName string, aggKindMap map[string][]KindConfig) *valueRecorder {
	return &valueRecorder{
		name:        recorderName,
		labelValues: sync.Map{},
		aggKindMap:  aggKindMap,
	}
}

// Record is thread-safe, and return the result value
func (r *valueRecorder) Record(key *aggregator.LabelKeys, metricValues []*model.Metric, timestamp uint64) {
	if key == nil {
		return
	}
	aggValues, ok := r.labelValues.Load(*key)
	if !ok {
		// double check to avoid double writing
		aggValues, _ = r.labelValues.LoadOrStore(*key, newAggValuesMap(metricValues, r.aggKindMap))
	}
	for _, metric := range metricValues {
		aggValues.(aggValuesMap).calculate(metric, timestamp)
	}
}

// dump a set of metric from counter cache.
// The return value holds the reference to the metric, not the copied one.
func (r *valueRecorder) dump() []*model.DataGroup {
	ret := make([]*model.DataGroup, 0)
	r.labelValues.Range(func(key, value interface{}) bool {
		k := key.(aggregator.LabelKeys)
		v := value.(aggValuesMap)
		metricGroup := model.NewDataGroup(r.name, k.GetLabels(), v.getTimestamp(), v.getAll()...)
		ret = append(ret, metricGroup)
		return true
	})
	return ret
}

// reset the labelValues for further aggregation.
// This method is not thread safe.
func (r *valueRecorder) reset() {
	r.labelValues = sync.Map{}
}
