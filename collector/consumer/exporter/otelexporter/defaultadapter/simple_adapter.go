package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"go.opentelemetry.io/otel/attribute"
)

type SimpleAdapter struct {
	acceptMetricGroupNames map[string]struct{}
	constLabels            []attribute.KeyValue
}

func (d *SimpleAdapter) Adapt(dataGroup *model.DataGroup) ([]*AdaptedResult, error) {
	if _, accept := d.acceptMetricGroupNames[dataGroup.Name]; !accept {
		return nil, nil
	}
	return []*AdaptedResult{
		{
			ResultType: Metric,
			AttrsMap:   dataGroup.Labels,
			AttrsList:  GetLabels(dataGroup.Labels, d.constLabels),
			Metrics:    dataGroup.Metrics,
			Timestamp:  dataGroup.Timestamp,
		},
	}, nil
}

func NewSimpleAdapter(
	acceptMetricGroupNames []string,
	customLabels []attribute.KeyValue,
) *SimpleAdapter {
	acceptMap := make(map[string]struct{}, len(acceptMetricGroupNames))
	for i := 0; i < len(acceptMetricGroupNames); i++ {
		acceptMap[acceptMetricGroupNames[i]] = struct{}{}
	}
	return &SimpleAdapter{
		constLabels:            customLabels,
		acceptMetricGroupNames: acceptMap,
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
