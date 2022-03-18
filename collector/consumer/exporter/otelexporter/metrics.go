package otelexporter

import (
	"context"
	"go.opentelemetry.io/otel/metric"
)

var gaugeGroupReceivedMetric = "kindling_telemetry_metrics_gaugegroups_exporter_received_total"
var metricExportedCountMetrics = "kindling_telemetry_metrics_metrics_exporter_exported_total"

type selfMetrics struct {
	gaugeGroupReceiverCounter  metric.Int64Counter
	metricExportedCountMetrics metric.Int64UpDownCounterObserver
}

func NewSelfMetrics(meterProvider metric.MeterProvider) *selfMetrics {
	return &selfMetrics{
		gaugeGroupReceiverCounter: metric.Must(meterProvider.Meter("kindling")).NewInt64Counter(gaugeGroupReceivedMetric),
		metricExportedCountMetrics: metric.Must(meterProvider.Meter("kindling")).NewInt64UpDownCounterObserver(
			metricExportedCountMetrics, func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(int64(len(labelsSet)))
			}),
	}
}
