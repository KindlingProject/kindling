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

func (k aggregatorKind) name() string {
	switch k {
	case sumKind:
		return "sum"
	case maxKind:
		return "max"
	case avgKind:
		return "avg"
	case lastKind:
		return "last"
	default:
		return ""
	}
}

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
	calculate(name string, value int64)
	get(name string) []*model.Gauge
	getAll() []*model.Gauge
}

type defaultValuesMap struct {
	values map[string][]aggregatedValues
}

func newAggValuesMap(gauges []*model.Gauge, kindMap map[string][]aggregatorKind) aggValuesMap {
	ret := &defaultValuesMap{values: make(map[string][]aggregatedValues)}
	for _, gauge := range gauges {
		kindSlice, found := kindMap[gauge.Name]
		if !found {
			continue
		}

		aggValuesSlice := make([]aggregatedValues, len(kindSlice))
		for i, kind := range kindSlice {
			aggValuesSlice[i] = newAggValue(gauge.Name+"_"+kind.name(), kind)
		}
		ret.values[gauge.Name] = aggValuesSlice
	}
	return ret
}

// calculate returns the result value
func (m *defaultValuesMap) calculate(name string, value int64) {
	vSlice, ok := m.values[name]
	if !ok {
		return
	}
	for _, v := range vSlice {
		v.calculate(value)
	}
}

func (m *defaultValuesMap) get(name string) []*model.Gauge {
	vSlice, ok := m.values[name]
	if !ok {
		return nil
	}
	ret := make([]*model.Gauge, len(vSlice))
	for i, v := range vSlice {
		ret[i] = &model.Gauge{
			Name:  v.getName(),
			Value: v.get(),
		}
	}
	return ret
}

func (m *defaultValuesMap) getAll() []*model.Gauge {
	ret := make([]*model.Gauge, 0)
	for k := range m.values {
		ret = append(ret, m.get(k)...)
	}
	return ret
}

func newAggValue(name string, kind aggregatorKind) aggregatedValues {
	switch kind {
	case sumKind:
		return &sumValue{name: name}
	case maxKind:
		return &maxValue{name: name}
	case avgKind:
		return &avgValue{name: name}
	case lastKind:
		return &lastValue{name: name}
	default:
		return &lastValue{name: name}
	}
}

type aggregatedValues interface {
	calculate(value int64) int64
	// get returns the value
	get() int64
	// getName returns the value's name
	getName() string
}

type maxValue struct {
	name  string
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
func (v *maxValue) getName() string {
	return v.name
}

type sumValue struct {
	name  string
	value int64
}

func (v *sumValue) calculate(value int64) int64 {
	return atomic.AddInt64(&v.value, value)
}
func (v *sumValue) get() int64 {
	return atomic.LoadInt64(&v.value)
}
func (v *sumValue) getName() string {
	return v.name
}

type avgValue struct {
	name  string
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
func (v *avgValue) getName() string {
	return v.name
}

type lastValue struct {
	name  string
	value int64
}

func (v *lastValue) calculate(value int64) int64 {
	return atomic.SwapInt64(&v.value, value)
}
func (v *lastValue) get() int64 {
	return atomic.LoadInt64(&v.value)
}
func (v *lastValue) getName() string {
	return v.name
}
