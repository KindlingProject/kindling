package otelexporter

import (
	"context"
	"github.com/Kindling-project/kindling/collector/consumer/exporter/otelexporter/defaultadapter"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	apitrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (e *OtelExporter) Consume(gaugeGroup *model.GaugeGroup) error {
	if gaugeGroup == nil {
		// no need consume
		return nil
	}
	gaugeGroupReceiverCounter.Add(context.Background(), 1, attribute.String("name", gaugeGroup.Name))
	if ce := e.telemetry.Logger.Check(zap.DebugLevel, "exporter receives a gaugeGroup: "); ce != nil {
		ce.Write(
			zap.String("gaugeGroup", gaugeGroup.String()),
		)
	}

	var hasResult = false
	for i := 0; i < len(e.adapters); i++ {
		results, _ := e.adapters[i].Adapt(gaugeGroup)
		if results != nil && len(results) > 0 {
			e.Export(results, e.adapters[i])
			hasResult = true
		}
	}
	if hasResult == false {
		if ce := e.telemetry.Logger.Check(zap.DebugLevel, "No adapter can deal with this gaugeGroup"); ce != nil {
			ce.Write(
				zap.String("gaugeGroup", gaugeGroup.String()),
			)
		}
	}
	return nil
}

func (e *OtelExporter) Export(results []*defaultadapter.AdaptedResult, adapter defaultadapter.Adapter) {
	for i := 0; i < len(results); i++ {
		result := results[i]
		switch result.ResultType {
		case defaultadapter.Metric:
			e.exportMetric(result)
		case defaultadapter.Trace:
			e.exportTrace(result)
		default:
			e.telemetry.Logger.Error("Unexpected ResultType", zap.String("type", string(result.ResultType)))
		}
	}
}

func (e *OtelExporter) exportTrace(result *defaultadapter.AdaptedResult) {
	if e.defaultTracer != nil && e.cfg.AdapterConfig.NeedTraceAsResourceSpan {
		_, span := e.defaultTracer.Start(
			context.Background(),
			constvalues.SpanInfo,
			apitrace.WithAttributes(result.Attrs...),
		)
		span.End()
	} else if e.defaultTracer == nil && e.cfg.AdapterConfig.NeedTraceAsResourceSpan {
		e.telemetry.Logger.Error("send span failed: this exporter can not support Span Data", zap.String("exporter", e.cfg.ExportKind))
	}
}

func (e *OtelExporter) exportMetric(result *defaultadapter.AdaptedResult) {
	// Get Measurement
	measurements := make([]metric.Measurement, 0, len(result.Gauges))
	for s := 0; s < len(result.Gauges); s++ {
		gauge := result.Gauges[s]
		if metricKind, ok := e.findInstrumentKind(gauge.Name); ok && metricKind == MAGaugeKind {
			err := e.instrumentFactory.recordLastValue(gauge.Name, &model.GaugeGroup{
				Name:      result.AggGroupName,
				Values:    []*model.Gauge{{gauge.Name, gauge.Value}},
				Labels:    result.Labels,
				Timestamp: result.Timestamp,
			})
			if err != nil {
				e.telemetry.Logger.Error("Failed to record Gauge", zap.Error(err))
			}
		} else if ok {
			measurements = append(measurements, e.instrumentFactory.getInstrument(gauge.Name, metricKind).Measurement(gauge.Value))
		} else {
			//measurements = append(measurements, e.instrumentFactory.getInstrument(metricName, MACounterKind).Measurement(gauge.Value))
			e.telemetry.Logger.Warn("This metric don't have any metricKind,please update this in kindlingCfg!", zap.String("GaugeGroup", gauge.Name))
		}
	}
	if len(measurements) > 0 {
		e.instrumentFactory.meter.RecordBatch(context.Background(), result.Attrs, measurements...)
	}
}
