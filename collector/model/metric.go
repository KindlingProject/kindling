package model

const (
	IntMetricType MetricType = iota
	HistogramMetricType
	NoneMetricType
)

type MetricType int

type Metric struct {
	Name string
	//	Data can be assigned by:
	//	Int
	//	Histogram
	Data isMetricData
}

func (i *Metric) GetData() isMetricData {
	if i != nil {
		return i.Data
	}
	return nil
}

func (i *Metric) GetInt() *Int {
	if x, ok := i.GetData().(*Metric_Int); ok {
		return x.Int
	}
	return nil
}

func (i *Metric) GetHistogram() *Histogram {
	if x, ok := i.GetData().(*Metric_Histogram); ok {
		return x.Histogram
	}
	return nil
}

func (i *Metric) DataType() MetricType {
	switch i.GetData().(type) {
	case *Metric_Int:
		return IntMetricType
	case *Metric_Histogram:
		return HistogramMetricType
	default:
		return NoneMetricType
	}
}

func (i *Metric) Clear() {
	i.Data = nil
}

type Int struct {
	Value int64
}

func NewIntMetric(name string, value int64) *Metric {
	return &Metric{Name: name, Data: &Metric_Int{Int: &Int{Value: value}}}
}

type Histogram struct {
	Sum                int64
	Count              uint64
	ExplicitBoundaries []int64
	BucketCounts       []uint64
}

func NewHistogramMetric(name string, histogram *Histogram) *Metric {
	return &Metric{Name: name, Data: &Metric_Histogram{Histogram: histogram}}
}

type isMetricData interface {
	isMetricData()
}

func (*Metric_Int) isMetricData()       {}
func (*Metric_Histogram) isMetricData() {}

type Metric_Int struct {
	Int *Int
}

type Metric_Histogram struct {
	Histogram *Histogram
}
