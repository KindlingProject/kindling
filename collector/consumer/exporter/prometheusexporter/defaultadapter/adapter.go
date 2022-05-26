package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"go.opentelemetry.io/otel/attribute"
)

// Adapter is used to transform *model.DataGroup into Trace or Metric
type Adapter interface {
	Adapt(group *model.DataGroup) ([]*AdaptedResult, error)
}

type AdaptedResult struct {
	ResultType ResultType

	// AttrsMap contains labels for Async Metric
	AttrsMap  *model.AttributeMap
	Metrics   []*model.Metric
	Timestamp uint64

	// FreeAttrsMap provides an interface for those adapters which need to reuse AttrsList and AttrsMap
	FreeAttrsMap
	FreeAttrsList
}

type FreeAttrsMap func(attrsMap *model.AttributeMap)
type FreeAttrsList func(attrsList []attribute.KeyValue)

type ResultType string

const (
	Metric ResultType = "metric"
)

func (r *AdaptedResult) Free() {
	if r.FreeAttrsMap != nil {
		r.FreeAttrsMap(r.AttrsMap)
	}
}
