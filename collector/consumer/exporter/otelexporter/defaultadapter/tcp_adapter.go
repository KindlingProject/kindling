package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/consumer/exporter/otelexporter"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"go.opentelemetry.io/otel/attribute"
)

type TcpAdapter struct {
	tcpAdapterCache *adapterCache
}

func (t *TcpAdapter) Adapt(gaugeGroup *model.GaugeGroup) ([]*otelexporter.AdaptedResult, error) {
	if gaugeGroup.Name != constnames.TcpGaugeGroupName {
		return nil, nil
	}
	results := make([]*otelexporter.AdaptedResult, 0, 1)

	attrs, _ := t.tcpAdapterCache.adapt(gaugeGroup)
	results = append(results, &otelexporter.AdaptedResult{
		ResultType: otelexporter.Metric,
		Attrs:      attrs,
		Gauges:     gaugeGroup.Values,
		RenameRule: otelexporter.KeepOrigin,
		OriginData: gaugeGroup,
	})
	return results, nil
}

func (t *TcpAdapter) Transform(gaugeGroup *model.GaugeGroup) (*model.AttributeMap, error) {
	// Only the SRTTMicroseconds
	return t.tcpAdapterCache.transform(gaugeGroup)
}

func NewTcpAdapter(
	customLabels []attribute.KeyValue,
) *TcpAdapter {
	cache, _ := newAdapterBuilder(tcpBaseDictList, nil).withConstLabels(customLabels).build()
	return &TcpAdapter{
		cache,
	}
}
