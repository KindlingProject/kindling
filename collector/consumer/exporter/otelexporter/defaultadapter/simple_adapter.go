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
			Attrs:      otelexporter.GetLabels(gaugeGroup.Labels, d.constLabels),
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
