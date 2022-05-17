package model

const (
	IntGaugeType GaugeType = iota
	HistogramGaugeType
	NoneGaugeType
)

type GaugeType int

type Gauge struct {
	Name string
	//	Data can be assigned by:
	//	Int
	//	Histogram
	Data isGaugeData
}

func (i *Gauge) GetData() isGaugeData {
	if i != nil {
		return i.Data
	}
	return nil
}

func (i *Gauge) GetInt() *Int {
	if x, ok := i.GetData().(*Gauge_Int); ok {
		return x.Int
	}
	return nil
}

func (i *Gauge) GetHistogram() *Histogram {
	if x, ok := i.GetData().(*Gauge_Histogram); ok {
		return x.Histogram
	}
	return nil
}

func (i *Gauge) DataType() GaugeType {
	switch i.GetData().(type) {
	case *Gauge_Int:
		return IntGaugeType
	case *Gauge_Histogram:
		return HistogramGaugeType
	default:
		return NoneGaugeType
	}
}

func (i *Gauge) Clear() {
	i.Data = nil
}

type Int struct {
	Value int64
}

func NewIntGauge(name string, value int64) *Gauge {
	return &Gauge{Name: name, Data: &Gauge_Int{Int: &Int{Value: value}}}
}

type Histogram struct {
	Sum                int64
	Count              uint64
	ExplicitBoundaries []int64
	BucketCounts       []uint64
}

func NewHistogramGauge(name string, histogram *Histogram) *Gauge {
	return &Gauge{Name: name, Data: &Gauge_Histogram{Histogram: histogram}}
}

type isGaugeData interface {
	isGaugeData()
}

func (*Gauge_Int) isGaugeData()       {}
func (*Gauge_Histogram) isGaugeData() {}

type Gauge_Int struct {
	Int *Int
}

type Gauge_Histogram struct {
	Histogram *Histogram
}
