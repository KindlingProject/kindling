package prometheusexporter

import (
	"github.com/Kindling-project/kindling/collector/consumer/exporter/otelexporter/defaultadapter"
	"github.com/Kindling-project/kindling/collector/model"
	"go.uber.org/zap"
)

func (p *prometheusExporter) Consume(metricGroup *model.DataGroup) error {
	if metricGroup == nil {
		// no need consume
		return nil
	}
	if ce := p.telemetry.Logger.Check(zap.DebugLevel, "exporter receives a metricGroup: "); ce != nil {
		ce.Write(
			zap.String("metricGroup", metricGroup.String()),
		)
	}

	if adapters, ok := p.adapters[metricGroup.Name]; ok {
		for i := 0; i < len(adapters); i++ {
			results, err := adapters[i].Adapt(metricGroup)
			if err != nil {
				p.telemetry.Logger.Error("Failed to adapt metricGroup", zap.Error(err))
			}
			if results != nil && len(results) > 0 {
				p.Export(results)
			}
		}
	} else {
		results, err := p.defaultAdapter.Adapt(metricGroup)
		if err != nil {
			p.telemetry.Logger.Error("Failed to adapt metricGroup", zap.Error(err))
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
	p.collector.recordMetricGroups(model.NewDataGroup("", result.AttrsMap, result.Timestamp, result.Metrics...))
}
