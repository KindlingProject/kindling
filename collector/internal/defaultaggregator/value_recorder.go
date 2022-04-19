package defaultaggregator

import (
	"github.com/Kindling-project/kindling/collector/internal"
	"github.com/Kindling-project/kindling/collector/model"
	"sync"
)

type valueRecorder struct {
	name        string
	labelValues map[internal.LabelKeys]aggValuesMap
	// mutex is only used to make sure the access to the labelValues is thread-safe.
	// aggValuesMap is responsible for its own thread-safe access.
	mutex      sync.RWMutex
	aggKindMap map[string][]KindConfig
}

func newValueRecorder(recorderName string, aggKindMap map[string][]KindConfig) *valueRecorder {
	return &valueRecorder{
		name:        recorderName,
		labelValues: make(map[internal.LabelKeys]aggValuesMap),
		mutex:       sync.RWMutex{},
		aggKindMap:  aggKindMap,
	}
}

// Record is thread-safe, and return the result value
func (r *valueRecorder) Record(key *internal.LabelKeys, gaugeValues []*model.Gauge, timestamp uint64) {
	if key == nil {
		return
	}
	r.mutex.RLock()
	aggValues, ok := r.labelValues[*key]
	r.mutex.RUnlock()
	if !ok {
		r.mutex.Lock()
		// double check to avoid double writing
		aggValues, ok = r.labelValues[*key]
		if !ok {
			aggValues = newAggValuesMap(gaugeValues, r.aggKindMap)
			r.labelValues[*key] = aggValues
		}
		r.mutex.Unlock()
	}
	for _, gauge := range gaugeValues {
		aggValues.calculate(gauge.Name, gauge.Value, timestamp)
	}
}

// dump a set of metric from counter cache.
// The return value holds the reference to the metric, not the copied one.
func (r *valueRecorder) dump() []*model.GaugeGroup {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	ret := make([]*model.GaugeGroup, len(r.labelValues))
	index := 0
	for k, v := range r.labelValues {
		gaugeGroup := model.NewGaugeGroup(r.name, k.GetLabels(), v.getTimestamp(), v.getAll()...)
		ret[index] = gaugeGroup
		index++
	}
	return ret
}

func (r *valueRecorder) reset() {
	r.mutex.Lock()
	r.labelValues = make(map[internal.LabelKeys]aggValuesMap)
	r.mutex.Unlock()
}
