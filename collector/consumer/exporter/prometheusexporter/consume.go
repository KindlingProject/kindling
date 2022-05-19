package prometheusexporter

import (
	"github.com/Kindling-project/kindling/collector/consumer/exporter/otelexporter/defaultadapter"
	"github.com/Kindling-project/kindling/collector/model"
	"go.uber.org/zap"
)

func (p *prometheusExporter) Consume(gaugeGroup *model.GaugeGroup) error {
	if gaugeGroup == nil {
		// no need consume
		return nil
	}
	if ce := p.telemetry.Logger.Check(zap.DebugLevel, "exporter receives a gaugeGroup: "); ce != nil {
		ce.Write(
			zap.String("gaugeGroup", gaugeGroup.String()),
		)
	}

	if adapters, ok := p.adapters[gaugeGroup.Name]; ok {
		for i := 0; i < len(adapters); i++ {
			results, err := adapters[i].Adapt(gaugeGroup)
			if err != nil {
				p.telemetry.Logger.Error("Failed to adapt gaugeGroup", zap.Error(err))
			}
			if results != nil && len(results) > 0 {
				p.Export(results)
			}
		}
	} else {
		results, err := p.defaultAdapter.Adapt(gaugeGroup)
		if err != nil {
			p.telemetry.Logger.Error("Failed to adapt gaugeGroup", zap.Error(err))
		}
		if results != nil && len(results) > 0 {
			p.Export(results)
		}
	}
	return nil
}

func (p *prometheusExporter) Export(results []*defaultadapter.AdaptedResult) {
	for i := 0; i < len(results); i++ {
		result := results[i]
		switch result.ResultType {
		case defaultadapter.Metric:
			p.exportMetric(result)
		default:
			p.telemetry.Logger.Error("Unexpected ResultType", zap.String("type", string(result.ResultType)))
		}
		result.Free()
	}
}

func (p *prometheusExporter) exportMetric(result *defaultadapter.AdaptedResult) {
	p.collector.recordGaugeGroups(model.NewGaugeGroup("", result.AttrsMap, result.Timestamp, result.Gauges...))
}
