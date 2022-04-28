package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"go.opentelemetry.io/otel/attribute"
)

type Adapter interface {
	// Adapt  We use convert to deal with any GaugeGroup received by exporter, this method should contain functions below
	//
	Adapt(group *model.GaugeGroup) ([]*AdaptedResult, error)
}

type AdaptedResult struct {
	ResultType ResultType

	// Attrs Maybe null if Gauges is nil or only has gauges which need to be preAgg (like LastValue)
	Attrs []attribute.KeyValue
	// Labels Maybe null if Gauges is nil or only has gauges which don't need to be preAgg (like counter or histogram)
	Labels *model.AttributeMap

	// Metrics to export
	Gauges []*model.Gauge

	// Timestamp
	Timestamp uint64

	FreeAttrsMap
	FreeAttrsList
}

type FreeAttrsMap func(attrsMap *model.AttributeMap)
type FreeAttrsList func(attrsList []attribute.KeyValue)

type ResultType string

const (
	Metric ResultType = "metric"
	Trace             = "trace"
)

type RenameRule int

const (
	ServerMetrics RenameRule = iota
	TopologyMetrics
	KeepOrigin
)

func (r *AdaptedResult) Free() {
	if r.FreeAttrsMap != nil {
		r.FreeAttrsMap(r.Labels)
	}
	if r.FreeAttrsList != nil {
		r.FreeAttrsList(r.Attrs)
	}
}
