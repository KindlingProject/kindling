package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/consumer/exporter/otelexporter"
	"github.com/Kindling-project/kindling/collector/model"
	"go.opentelemetry.io/otel/attribute"
)

type SimpleAdapter struct {
	constLabels []attribute.KeyValue
}

func (d *SimpleAdapter) Adapt(gaugeGroup *model.GaugeGroup) ([]*AdaptedResult, error) {
	return []*AdaptedResult{
		{
			ResultType: Metric,
			Attrs:      GetLabels(gaugeGroup.Labels, d.constLabels),
			Gauges:     gaugeGroup.Values,
			Timestamp:  gaugeGroup.Timestamp,
		},
	}, nil
}

func NewDefaultAdapter(
	customLabels []attribute.KeyValue,
) *SimpleAdapter {
	return &SimpleAdapter{constLabels: customLabels}
}

func GetLabels(attributeMap *model.AttributeMap, customLabels []attribute.KeyValue) []attribute.KeyValue {
	kv := otelexporter.ToStringKeyValues(attributeMap.GetValues())
	kv = append(kv, customLabels...)
	return kv
}
