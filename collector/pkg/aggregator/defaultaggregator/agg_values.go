package defaultaggregator

import (
	"github.com/Kindling-project/kindling/collector/model"
	"sync"
	"sync/atomic"
)

type AggregatorKind int

const (
	SumKind AggregatorKind = iota
	MaxKind
	AvgKind
	LastKind
	CountKind
	HistogramKind
)

func (k AggregatorKind) name() string {
	switch k {
	case SumKind:
		return "sum"
	case MaxKind:
		return "max"
	case AvgKind:
		return "avg"
	case LastKind:
		return "last"
	case CountKind:
		return "count"
	case HistogramKind:
		return "histogram"
	default:
		return ""
	}
}

func GetAggregatorKind(kind string) AggregatorKind {
	switch kind {
	case "sum":
		return SumKind
	case "max":
		return MaxKind
	case "avg":
		return AvgKind
	case "last":
		return LastKind
	case "count":
		return CountKind
	case "histogram":
		return HistogramKind
	default:
		return SumKind
	}
}

type (
	AggregatedConfig struct {
		KindMap map[string][]KindConfig
	}

	KindConfig struct {
		OutputName string
		Kind       AggregatorKind
		// Only HistogramKind has this value
		ExplicitBoundaries []int64
	}
)

type aggValuesMap interface {
	// calculate should be thread-safe to use
	calculate(name string, value int64, timestamp uint64)
	get(name string) []*model.Gauge
	getAll() []*model.Gauge
	getTimestamp() uint64
}

type defaultValuesMap struct {
	values    map[string][]aggregatedValues
	timestamp uint64
}

func newAggValuesMap(gauges []*model.Gauge, kindMap map[string][]KindConfig) aggValuesMap {
	ret := &defaultValuesMap{values: make(map[string][]aggregatedValues)}
	for _, gauge := range gauges {
		kindSlice, found := kindMap[gauge.Name]
		if !found {
			continue
		}

		aggValuesSlice := make([]aggregatedValues, len(kindSlice))
		for i, kind := range kindSlice {
			aggValuesSlice[i] = newAggValue(kind)
		}
		ret.values[gauge.Name] = aggValuesSlice
	}
	return ret
}

// calculate returns the result value
func (m *defaultValuesMap) calculate(name string, value int64, timestamp uint64) {
	vSlice, ok := m.values[name]
	if !ok {
		return
	}
	for _, v := range vSlice {
		v.calculate(value)
	}
	atomic.StoreUint64(&m.timestamp, timestamp)
}

func (m *defaultValuesMap) get(name string) []*model.Gauge {
	vSlice, ok := m.values[name]
	if !ok {
		return nil
	}
	ret := make([]*model.Gauge, len(vSlice))
	for i, v := range vSlice {
		ret[i] = v.get()
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

func (m *defaultValuesMap) getTimestamp() uint64 {
	return atomic.LoadUint64(&m.timestamp)
}

func newAggValue(cfg KindConfig) aggregatedValues {
	name := cfg.OutputName
	switch cfg.Kind {
	case SumKind:
		return &sumValue{name: name}
	case MaxKind:
		return &maxValue{name: name}
	case AvgKind:
		return &avgValue{name: name}
	case LastKind:
		return &lastValue{name: name}
	case CountKind:
		return &countValue{name: name}
	case HistogramKind:
		return &histogramValue{name: name, explicitBoundaries: cfg.ExplicitBoundaries, bucketCounts: make([]uint64, len(cfg.ExplicitBoundaries))}
	default:
		return &lastValue{name: name}
	}
}

type aggregatedValues interface {
	calculate(value int64) int64
	// get returns the value
	get() *model.Gauge
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
func (v *maxValue) get() *model.Gauge {
	v.mut.RLock()
	defer v.mut.RUnlock()
	return model.NewIntGauge(v.name, v.value)
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
func (v *sumValue) get() *model.Gauge {
	return model.NewIntGauge(v.name, atomic.LoadInt64(&v.value))
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
func (v *avgValue) get() *model.Gauge {
	v.mut.RLock()
	defer v.mut.RUnlock()
	return model.NewIntGauge(v.name, v.value/v.count)
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
func (v *lastValue) get() *model.Gauge {
	return model.NewIntGauge(v.name, atomic.LoadInt64(&v.value))
}
func (v *lastValue) getName() string {
	return v.name
}

type countValue struct {
	name  string
	value int64
}

// calculate add 1 to its own value
func (v *countValue) calculate(_ int64) int64 {
	return atomic.AddInt64(&v.value, 1)
}
func (v *countValue) get() *model.Gauge {
	return model.NewIntGauge(v.name, atomic.LoadInt64(&v.value))
}
func (v *countValue) getName() string {
	return v.name
}

type histogramValue struct {
	name               string
	sum                int64
	count              uint64
	explicitBoundaries []int64
	bucketCounts       []uint64
	mut                sync.RWMutex
}

func (h *histogramValue) calculate(value int64) int64 {
	h.mut.Lock()
	defer h.mut.Unlock()
	h.sum += value
	h.count++
	for i := 0; i < len(h.explicitBoundaries); i++ {
		if value < h.explicitBoundaries[i] {
			atomic.AddUint64(&h.bucketCounts[i], 1)
		} else {
			break
		}
	}
	return int64(h.count)
}

func (h *histogramValue) get() *model.Gauge {
	h.mut.RLock()
	defer h.mut.RUnlock()
	return model.NewHistogramGauge(h.name, &model.Histogram{
		Sum:                h.sum,
		Count:              h.count,
		ExplicitBoundaries: h.explicitBoundaries,
		BucketCounts:       h.bucketCounts,
	})
}

func (h *histogramValue) getName() string {
	return h.name
}
