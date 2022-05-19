package defaultaggregator

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/pkg/aggregator"
	"sync/atomic"
	"time"
)

// GaugeGroup
// Name:
//   labels:
//      GaugeGroup

type CumulativeAggregator struct {
	*DefaultAggregator
	MetricExpiration time.Duration
}

type cumulativeValueRecorder struct {
	*valueRecorder
}

func (r *cumulativeValueRecorder) RecordWithDefaultSumKindAgg(key *aggregator.LabelKeys, gaugeValues []*model.Gauge, timestamp uint64, now time.Time) {
	if key == nil {
		return
	}
	aggValues, ok := r.labelValues.Load(*key)
	if !ok {
		// double check to avoid double writing
		aggValues, _ = r.labelValues.LoadOrStore(*key, newExpiredAggValuesMapWithDefaultSumKind(gaugeValues, r.aggKindMap))
	}
	for _, gauge := range gaugeValues {
		aggValues.(*expiredValuesMap).calculateWithExpired(gauge, timestamp, now)
	}
}

func remove(s []aggregatedValues, i int) []aggregatedValues {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (r *cumulativeValueRecorder) removeExpired(expirationTime time.Time) {
	expiredLabelValues := make([]interface{}, 0)
	r.labelValues.Range(func(key, value interface{}) bool {
		valueMap := value.(*expiredValuesMap)
		if expirationTime.After(valueMap.update) {
			expiredLabelValues = append(expiredLabelValues, key)
		}
		return true
	})
	for i := 0; i < len(expiredLabelValues); i++ {
		// TODO log debug expired
		r.labelValues.Delete(expiredLabelValues[i])
	}
}

func (c *CumulativeAggregator) AggregatorWithAllLabelsAndGauge(g *model.GaugeGroup, now time.Time) {
	name := g.Name
	c.mut.RLock()
	defer c.mut.RUnlock()
	recorder, ok := c.recordersMap.Load(name)
	// Won't enter this branch too many times, as the recordersMap
	// will become stable after running a period of time.
	if !ok {
		// double check to avoid double writing
		recorder, _ = c.recordersMap.LoadOrStore(name, &cumulativeValueRecorder{newValueRecorder(name, c.config.KindMap)})
	}
	key := aggregator.GetLabelsKeys(g.Labels)
	recorder.(*cumulativeValueRecorder).RecordWithDefaultSumKindAgg(key, g.Values, g.Timestamp, now)
}

func (c *CumulativeAggregator) DumpAndRemoveExpired(now time.Time) []*model.GaugeGroup {
	expirationTime := now.Add(-c.MetricExpiration)
	ret := make([]*model.GaugeGroup, 0)
	c.mut.Lock()
	defer c.mut.Unlock()
	c.recordersMap.Range(func(_, value interface{}) bool {
		recorder := value.(*cumulativeValueRecorder)
		ret = append(ret, recorder.dump()...)
		// Assume that the recordersMap will become stable after running a period of time.
		recorder.removeExpired(expirationTime)
		return true
	})
	return ret
}

func NewCumulativeAggregator(config *AggregatedConfig, metricExpiration time.Duration) *CumulativeAggregator {
	ret := &CumulativeAggregator{
		DefaultAggregator: NewDefaultAggregator(config),
		MetricExpiration:  metricExpiration,
	}
	return ret
}

type expiredValuesMap struct {
	*defaultValuesMap
	update time.Time
}

func (e *expiredValuesMap) calculateWithExpired(gauge *model.Gauge, timestamp uint64, now time.Time) {
	vSlice, ok := e.values[gauge.Name]
	if !ok {
		return
	}
	for _, v := range vSlice {
		if gauge.DataType() == model.IntGaugeType {
			v.calculate(gauge.GetInt().Value)
		} else if gauge.DataType() == model.HistogramGaugeType {
			v.merge(gauge.GetHistogram())
		}
	}
	e.update = now
	atomic.StoreUint64(&e.timestamp, timestamp)
}

func newExpiredAggValuesMapWithDefaultSumKind(gauges []*model.Gauge, kindMap map[string][]KindConfig) aggValuesMap {
	ret := &expiredValuesMap{defaultValuesMap: &defaultValuesMap{values: make(map[string][]aggregatedValues)}}
	for _, gauge := range gauges {
		kindSlice, found := kindMap[gauge.Name]
		var aggValuesSlice []aggregatedValues
		if found {
			aggValuesSlice = make([]aggregatedValues, len(kindSlice))
			for i, kind := range kindSlice {
				aggValuesSlice[i] = newAggValue(kind)
			}
		} else {
			aggValuesSlice = make([]aggregatedValues, 1)
			aggValuesSlice[0] = newAggValue(KindConfig{
				OutputName: gauge.Name,
				Kind:       SumKind,
			})
		}
		ret.values[gauge.Name] = aggValuesSlice
	}
	return ret
}
