package model

import (
	"fmt"
	"strings"
)

// GaugeGroup describes the result of analyzers.
// Notice: Currently the definition of GaugeGroup is not stable.
type GaugeGroup struct {
	Name      string
	Values    []*Gauge
	Labels    *AttributeMap
	Timestamp uint64
}

func NewGaugeGroup(name string, labels *AttributeMap, timestamp uint64, values ...*Gauge) *GaugeGroup {
	return &GaugeGroup{
		Name:      name,
		Values:    values,
		Labels:    labels,
		Timestamp: timestamp,
	}
}

func (g *GaugeGroup) GetGauge(name string) (*Gauge, bool) {
	for _, gauge := range g.Values {
		if gauge.Name == name {
			return gauge, true
		}
	}
	return nil, false
}

func (g *GaugeGroup) AddIntGaugeWithName(name string, value int64) {
	g.AddGauge(NewIntGauge(name, value))
}

func (g *GaugeGroup) AddGauge(gauge *Gauge) {
	if g.Values == nil {
		g.Values = make([]*Gauge, 0)
	}
	g.Values = append(g.Values, gauge)
}

// UpdateAddIntGauge overwrite the gauge with the key of 'name' if existing, or adds the gauge if not existing.
func (g *GaugeGroup) UpdateAddIntGauge(name string, value int64) {
	if gauge, ok := g.GetGauge(name); ok {
		gauge.Data = &Gauge_Int{Int: &Int{Value: value}}
	} else {
		g.AddIntGaugeWithName(name, value)
	}
}

func (g *GaugeGroup) RemoveGauge(name string) {
	newValues := make([]*Gauge, 0)
	for _, value := range g.Values {
		if value.Name == name {
			continue
		}
		newValues = append(newValues, value)
	}
	g.Values = newValues
}

func (g *GaugeGroup) String() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("GagugeGroup:\n"))
	str.WriteString(fmt.Sprintf("\tName: %s\n", g.Name))
	str.WriteString(fmt.Sprintf("\tValues: \n"))
	for _, v := range g.Values {
		switch v.DataType() {
		case IntGaugeType:
			str.WriteString(fmt.Sprintf("\t\t{Name: %s, Value: %d}\n", v.Name, v.GetInt().Value))
		case HistogramGaugeType:
			histogram := v.GetHistogram()
			str.WriteString(fmt.Sprintf("\t\t{Name: %s, Sum: %d, Count: %d,ExplicitBoundaries: %v,BucketCount: %v}\n", v.Name, histogram.Sum, histogram.Count, histogram.ExplicitBoundaries, histogram.BucketCounts))
		}
	}
	str.WriteString(fmt.Sprintf("\tLabels: %v\n", g.Labels))
	str.WriteString(fmt.Sprintf("\tTimestamp: %d\n", g.Timestamp))
	return str.String()
}

func (g *GaugeGroup) Reset() {
	g.Name = ""
	for _, v := range g.Values {
		v.Clear()
	}
	g.Labels.ResetValues()
	g.Timestamp = 0
}
