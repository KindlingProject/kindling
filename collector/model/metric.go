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

type Gauge struct {
	Name  string
	Value int64
}

func NewGaugeGroup(name string, labels *AttributeMap, timestamp uint64, values ...*Gauge) *GaugeGroup {
	return &GaugeGroup{
		Name:      name,
		Values:    values,
		Labels:    labels,
		Timestamp: timestamp,
	}
}

func (g *GaugeGroup) GetValue(name string) (*Gauge, bool) {
	for _, gauge := range g.Values {
		if gauge.Name == name {
			return gauge, true
		}
	}
	return &Gauge{}, false
}

func (g *GaugeGroup) String() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("GagugeGroup:\n"))
	str.WriteString(fmt.Sprintf("\tName: %s\n", g.Name))
	str.WriteString(fmt.Sprintf("\tValues: %v\n", g.Values))
	str.WriteString(fmt.Sprintf("\tLabels: %v\n", g.Labels))
	str.WriteString(fmt.Sprintf("\tTimestamp: %d\n", g.Timestamp))
	return str.String()
}

func (g *GaugeGroup) Reset() {
	g.Name = ""
	for _, v := range g.Values {
		v.Value = 0
	}
	g.Labels.ResetValues()
	g.Timestamp = 0
}
