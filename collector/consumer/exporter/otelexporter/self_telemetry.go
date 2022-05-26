package otelexporter

import (
	"context"
	"go.opentelemetry.io/otel/metric"
	"sync"
)

var otelexporterMetricgroupsReceivedTotal = "kindling_telemetry_otelexporter_metricgroups_received_total"
var otelexporterCardinalitySize = "kindling_telemetry_otelexporter_cardinality_size"

var once sync.Once

var labelsSet map[labelKey]bool
var labelsSetMutex sync.RWMutex

var dataGroupReceiverCounter metric.Int64Counter
var metricExportedCardinalitySize metric.Int64UpDownCounterObserver

func newSelfMetrics(meterProvider metric.MeterProvider) {
	once.Do(func() {
		dataGroupReceiverCounter = metric.Must(meterProvider.Meter("kindling")).NewInt64Counter(otelexporterMetricgroupsReceivedTotal)
		metricExportedCardinalitySize = metric.Must(meterProvider.Meter("kindling")).NewInt64UpDownCounterObserver(
			otelexporterCardinalitySize, func(ctx context.Context, result metric.Int64ObserverResult) {
				labelsSetMutex.Lock()
				defer labelsSetMutex.Unlock()
				result.Observe(int64(len(labelsSet)))
			})
		labelsSet = make(map[labelKey]bool, 0)
	})
}
