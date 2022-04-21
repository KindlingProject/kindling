package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"go.opentelemetry.io/otel/attribute"
)

type Adapter interface {
	Adapt(group *model.GaugeGroup) ([]*AdaptedResult, error)

	// Transform Only Aggregator Value should use the Transform
	Transform(group *model.GaugeGroup) (*model.AttributeMap, error)
}

type AdaptedResult struct {
	ResultType ResultType
	// Maybe null if Gauges is nil or only has gauges which need to be preAgg (like LastValue)
	Attrs      []attribute.KeyValue
	Gauges     []*model.Gauge
	RenameRule RenameRule
	OriginData *model.GaugeGroup
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
