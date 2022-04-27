package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"go.opentelemetry.io/otel/attribute"
)

type SimpleAdapter struct {
	acceptGaugeGroupNames map[string]struct{}
	constLabels           []attribute.KeyValue
}

func (d *SimpleAdapter) Adapt(gaugeGroup *model.GaugeGroup) ([]*AdaptedResult, error) {
	if _, ok := d.acceptGaugeGroupNames[gaugeGroup.Name]; ok {
		return []*AdaptedResult{
			{
				ResultType: Metric,
				Labels:     gaugeGroup.Labels,
				Attrs:      GetLabels(gaugeGroup.Labels, d.constLabels),
				Gauges:     gaugeGroup.Values,
				Timestamp:  gaugeGroup.Timestamp,
			},
		}, nil
	}
	return nil, nil
}

func NewDefaultAdapter(
	acceptGaugeGroupNames []string,
	customLabels []attribute.KeyValue,
) *SimpleAdapter {
	acceptMap := make(map[string]struct{}, len(acceptGaugeGroupNames))
	for i := 0; i < len(acceptGaugeGroupNames); i++ {
		acceptMap[acceptGaugeGroupNames[i]] = struct{}{}
	}
	return &SimpleAdapter{
		constLabels:           customLabels,
		acceptGaugeGroupNames: acceptMap,
	}
}

func GetLabels(attributeMap *model.AttributeMap, customLabels []attribute.KeyValue) []attribute.KeyValue {
	return ToStringKeyValues(attributeMap.GetValues(), customLabels)
}

func ToStringKeyValues(values map[string]model.AttributeValue, customLabels []attribute.KeyValue) []attribute.KeyValue {
	stringKeyValues := make([]attribute.KeyValue, 0, len(values)+len(customLabels))
	for k, v := range values {
		stringKeyValues = append(stringKeyValues, attribute.String(k, v.ToString()))
	}
	return append(stringKeyValues, customLabels...)
}
