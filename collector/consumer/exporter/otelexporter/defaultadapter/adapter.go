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

	// AggGroupName is a key for those gauges which need to preAgg
	AggGroupName string

	// Metrics to export
	Gauges []*model.Gauge

	// Timestamp
	Timestamp uint64
}

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
