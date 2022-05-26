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

func (e *OtelExporter) Consume(metricGroup *model.DataGroup) error {
	if metricGroup == nil {
		// no need consume
		return nil
	}

	metricGroupReceiverCounter.Add(context.Background(), 1, attribute.String("name", metricGroup.Name))
	if ce := e.telemetry.Logger.Check(zap.DebugLevel, "exporter receives a metricGroup: "); ce != nil {
		ce.Write(
			zap.String("metricGroup", metricGroup.String()),
		)
	}

	for i := 0; i < len(e.adapters); i++ {
		results, err := e.adapters[i].Adapt(metricGroup)
		if err != nil {
			e.telemetry.Logger.Error("Failed to adapt metricGroup", zap.Error(err))
		}
		if results != nil && len(results) > 0 {
			e.Export(results)
		}
	}
	return nil
}

func (e *OtelExporter) Export(results []*defaultadapter.AdaptedResult) {
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
		result.Free()
	}
}

func (e *OtelExporter) exportTrace(result *defaultadapter.AdaptedResult) {
	if e.defaultTracer == nil {
		e.telemetry.Logger.Error("Send span failed: this exporter doesn't support Span Data", zap.String("exporter", e.cfg.ExportKind))
		return
	}
	_, span := e.defaultTracer.Start(
		context.Background(),
		constvalues.SpanInfo,
		apitrace.WithAttributes(result.AttrsList...),
	)
	span.End()
}

func (e *OtelExporter) exportMetric(result *defaultadapter.AdaptedResult) {
	measurements := make([]metric.Measurement, 0, len(result.Metrics))
	for s := 0; s < len(result.Metrics); s++ {
		metric := result.Metrics[s]
		if metricKind, ok := e.findInstrumentKind(metric.Name); ok && metricKind == MAGaugeKind {
			if result.AttrsMap == nil {
				e.telemetry.Logger.Error("Unexpected Error: no labels find for MAGaugeKind", zap.String("MetricName", metric.Name))
			}
			err := e.instrumentFactory.recordLastValue(metric.Name, &model.DataGroup{
				Name:      metric.Name,
				Metrics:   []*model.Metric{model.NewIntMetric(metric.Name, metric.GetInt().Value)},
				Labels:    result.AttrsMap,
				Timestamp: result.Timestamp,
			})
			if err != nil {
				e.telemetry.Logger.Error("Failed to record Metric", zap.Error(err))
			}
		} else if ok {
			measurements = append(measurements, e.instrumentFactory.getInstrument(metric.Name, metricKind).Measurement(metric.GetInt().Value))
		} else {
			e.telemetry.Logger.Warn("Undefined metricKind for this Metric", zap.String("MetricName", metric.Name))
		}
	}
	if len(measurements) > 0 {
		e.instrumentFactory.meter.RecordBatch(context.Background(), result.AttrsList, measurements...)
	}
}
