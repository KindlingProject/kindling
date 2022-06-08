package model

import (
	"fmt"
	"strings"
)

// DataGroup describes the result of analyzers.
// Notice: Currently the definition of DataGroup is not stable.
type DataGroup struct {
	Name      string
	Metrics   []*Metric
	Labels    *AttributeMap
	Timestamp uint64
}

func NewDataGroup(name string, labels *AttributeMap, timestamp uint64, values ...*Metric) *DataGroup {
	return &DataGroup{
		Name:      name,
		Metrics:   values,
		Labels:    labels,
		Timestamp: timestamp,
	}
}

func (g *DataGroup) GetMetric(name string) (*Metric, bool) {
	for _, metric := range g.Metrics {
		if metric.Name == name {
			return metric, true
		}
	}
	return nil, false
}

func (g *DataGroup) AddIntMetricWithName(name string, value int64) {
	g.AddMetric(NewIntMetric(name, value))
}

func (g *DataGroup) AddMetric(metric *Metric) {
	if g.Metrics == nil {
		g.Metrics = make([]*Metric, 0)
	}
	g.Metrics = append(g.Metrics, metric)
}

// UpdateAddIntMetric overwrite the metric with the key of 'name' if existing, or adds the metric if not existing.
func (g *DataGroup) UpdateAddIntMetric(name string, value int64) {
	if metric, ok := g.GetMetric(name); ok {
		metric.Data = &Metric_Int{Int: &Int{Value: value}}
	} else {
		g.AddIntMetricWithName(name, value)
	}
}

func (g *DataGroup) RemoveMetric(name string) {
	newValues := make([]*Metric, 0)
	for _, value := range g.Metrics {
		if value.Name == name {
			continue
		}
		newValues = append(newValues, value)
	}
	g.Metrics = newValues
}

func (g *DataGroup) String() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("GagugeGroup:\n"))
	str.WriteString(fmt.Sprintf("\tName: %s\n", g.Name))
	str.WriteString(fmt.Sprintf("\tValues: \n"))
	for _, v := range g.Metrics {
		switch v.DataType() {
		case IntMetricType:
			str.WriteString(fmt.Sprintf("\t\t{Name: %s, Value: %d}\n", v.Name, v.GetInt().Value))
		case HistogramMetricType:
			histogram := v.GetHistogram()
			str.WriteString(fmt.Sprintf("\t\t{Name: %s, Sum: %d, Count: %d,ExplicitBoundaries: %v,BucketCount: %v}\n", v.Name, histogram.Sum, histogram.Count, histogram.ExplicitBoundaries, histogram.BucketCounts))
		}
	}
	str.WriteString(fmt.Sprintf("\tLabels: %v\n", g.Labels))
	str.WriteString(fmt.Sprintf("\tTimestamp: %d\n", g.Timestamp))
	return str.String()
}

func (g *DataGroup) Reset() {
	g.Name = ""
	for _, v := range g.Metrics {
		v.Clear()
	}
	g.Labels.ResetValues()
	g.Timestamp = 0
}
