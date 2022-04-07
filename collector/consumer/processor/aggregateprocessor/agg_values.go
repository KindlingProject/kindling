package aggregateprocessor

import "github.com/Kindling-project/kindling/collector/model"

type aggregatorKind int

const (
	sumKind aggregatorKind = iota
	maxKind
	avgKind
	lastKind
)

type aggValuesMap interface {
	calculate(name string, value int64) int64
	get(name string) int64
	getAll() []*model.Gauge
}

type defaultValuesMap struct {
	values map[string]aggregatedValues
}

func newAggValuesMap(kindMap map[string]aggregatorKind) aggValuesMap {
	ret := &defaultValuesMap{values: make(map[string]aggregatedValues, len(kindMap))}
	for k, v := range kindMap {
		ret.values[k] = newAggValue(v)
	}
	return ret
}

// calculate returns the result value
func (m *defaultValuesMap) calculate(name string, value int64) int64 {
	v := m.values[name]
	return v.calculate(value)
}

func (m *defaultValuesMap) get(name string) int64 {
	return m.values[name].get()
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
}

func (v *maxValue) calculate(value int64) int64 {
	if v.value < value {
		v.value = value
	}
	return v.value
}
func (v *maxValue) get() int64 {
	return v.value
}

type sumValue struct {
	value int64
}

func (v *sumValue) calculate(value int64) int64 {
	v.value += value
	return v.value
}
func (v *sumValue) get() int64 {
	return v.value
}

type avgValue struct {
	value int64
	count int64
}

func (v *avgValue) calculate(value int64) int64 {
	sum := v.count * v.value
	v.count++
	avg := (sum + value) / v.count
	v.value = avg
	return v.value
}
func (v *avgValue) get() int64 {
	return v.value
}

type lastValue struct {
	value int64
}

func (v *lastValue) calculate(value int64) int64 {
	v.value = value
	return v.value
}
func (v *lastValue) get() int64 {
	return v.value
}
