package defaultadapter

import (
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
	kv := ToStringKeyValues(attributeMap.GetValues())
	kv = append(kv, customLabels...)
	return kv
}

func ToStringKeyValues(values map[string]model.AttributeValue) []attribute.KeyValue {
	stringKeyValues := make([]attribute.KeyValue, 0, len(values))
	for k, v := range values {
		stringKeyValues = append(stringKeyValues, attribute.String(k, v.ToString()))
	}
	return stringKeyValues
}
