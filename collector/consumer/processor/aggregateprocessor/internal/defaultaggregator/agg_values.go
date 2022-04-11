package defaultaggregator

import (
	"github.com/Kindling-project/kindling/collector/model"
	"sync"
	"sync/atomic"
)

type aggregatorKind int

const (
	sumKind aggregatorKind = iota
	maxKind
	avgKind
	lastKind
)

func toAggKindMap(input map[string][]string) map[string][]aggregatorKind {
	ret := make(map[string][]aggregatorKind, len(input))
	for k, v := range input {
		kindSlice := make([]aggregatorKind, len(v))
		for i, kind := range v {
			switch kind {
			case "sum":
				kindSlice[i] = sumKind
			case "max":
				kindSlice[i] = maxKind
			case "avg":
				kindSlice[i] = avgKind
			case "last":
				kindSlice[i] = lastKind
			}
		}
		ret[k] = kindSlice
	}
	return ret
}

type aggValuesMap interface {
	// calculate should be thread-safe to use
	calculate(name string, value int64) int64
	get(name string) int64
	getAll() []*model.Gauge
}

type defaultValuesMap struct {
	values map[string]aggregatedValues
}

func newAggValuesMap(gauges []*model.Gauge, kindMap map[string][]aggregatorKind) aggValuesMap {
	ret := &defaultValuesMap{values: make(map[string]aggregatedValues)}
	for _, gauge := range gauges {
		kindSlice, found := kindMap[gauge.Name]
		if !found {
			continue
		}
		for _, kind := range kindSlice {
			ret.values[gauge.Name] = newAggValue(kind)
		}
	}
	return ret
}

// calculate returns the result value
func (m *defaultValuesMap) calculate(name string, value int64) int64 {
	v, ok := m.values[name]
	if !ok {
		return -1
	}
	return v.calculate(value)
}

func (m *defaultValuesMap) get(name string) int64 {
	v, ok := m.values[name]
	if !ok {
		return -1
	}
	return v.get()
}

func (m *defaultValuesMap) getAll() []*model.Gauge {
	ret := make([]*model.Gauge, len(m.values))
	index := 0
	for k, v := range m.values {
		gauge := &model.Gauge{
			Name:  k,
			Value: v.get(),
		}
		ret[index] = gauge
		index++
	}
	return ret
}

func newAggValue(kind aggregatorKind) aggregatedValues {
	switch kind {
	case sumKind:
		return &sumValue{}
	case maxKind:
		return &maxValue{}
	case avgKind:
		return &avgValue{}
	case lastKind:
		return &lastValue{}
	default:
		return &lastValue{}
	}
}

type aggregatedValues interface {
	calculate(value int64) int64
	get() int64
}

type maxValue struct {
	value int64
	mut   sync.RWMutex
}

func (v *maxValue) calculate(value int64) int64 {
	v.mut.Lock()
	defer v.mut.Unlock()
	if v.value < value {
		v.value = value
	}
	return v.value
}
func (v *maxValue) get() int64 {
	v.mut.RLock()
	defer v.mut.RUnlock()
	return v.value
}

type sumValue struct {
	value int64
}

func (v *sumValue) calculate(value int64) int64 {
	return atomic.AddInt64(&v.value, value)
}
func (v *sumValue) get() int64 {
	return atomic.LoadInt64(&v.value)
}

type avgValue struct {
	value int64
	count int64
	mut   sync.RWMutex
}

// calculate returns the total count of data.
// Note it will not return the current average value which could be got by calling get()
func (v *avgValue) calculate(value int64) int64 {
	v.mut.Lock()
	defer v.mut.Unlock()
	v.value += value
	v.count++
	return v.count
}
func (v *avgValue) get() int64 {
	v.mut.RLock()
	defer v.mut.RUnlock()
	return v.value / v.count
}

type lastValue struct {
	value int64
}

func (v *lastValue) calculate(value int64) int64 {
	return atomic.SwapInt64(&v.value, value)
}
func (v *lastValue) get() int64 {
	return atomic.LoadInt64(&v.value)
}
