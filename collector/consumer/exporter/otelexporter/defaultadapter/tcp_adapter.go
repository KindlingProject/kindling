package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"go.opentelemetry.io/otel/attribute"
)

type TcpAdapter struct {
	tcpAdapterCache *LabelConverter
}

func (t *TcpAdapter) Adapt(gaugeGroup *model.GaugeGroup) ([]*AdaptedResult, error) {
	if gaugeGroup.Name != constnames.TcpGaugeGroupName {
		return nil, nil
	}
	results := make([]*AdaptedResult, 0, 1)

	attrs, _ := t.tcpAdapterCache.convert(gaugeGroup)
	labels, _ := t.tcpAdapterCache.transform(gaugeGroup)
	results = append(results, &AdaptedResult{
		ResultType:   Metric,
		Attrs:        attrs,
		Gauges:       gaugeGroup.Values,
		Labels:       labels,
		Timestamp:    gaugeGroup.Timestamp,
		AggGroupName: constnames.TcpGaugeGroupName,
	})
	return results, nil
}

func NewTcpAdapter(
	customLabels []attribute.KeyValue,
) *TcpAdapter {
	cache, _ := newAdapterBuilder(tcpBaseDictList, nil).withConstLabels(customLabels).build()
	return &TcpAdapter{
		cache,
	}
}
