package otelexporter

import (
	"context"
	"go.opentelemetry.io/otel/metric"
	"sync"
)

var otelexporterGaugegroupsReceivedTotal = "kindling_telemetry_otelexporter_gaugegroups_received_total"
var otelexporterCardinalitySize = "kindling_telemetry_otelexporter_cardinality_size"

var once sync.Once

var labelsSet map[labelKey]bool
var labelsSetMutex sync.RWMutex

var gaugeGroupReceiverCounter metric.Int64Counter
var metricExportedCardinalitySize metric.Int64UpDownCounterObserver

func newSelfMetrics(meterProvider metric.MeterProvider) {
	once.Do(func() {
		gaugeGroupReceiverCounter = metric.Must(meterProvider.Meter("kindling")).NewInt64Counter(otelexporterGaugegroupsReceivedTotal)
		metricExportedCardinalitySize = metric.Must(meterProvider.Meter("kindling")).NewInt64UpDownCounterObserver(
			otelexporterCardinalitySize, func(ctx context.Context, result metric.Int64ObserverResult) {
				labelsSetMutex.Lock()
				defer labelsSetMutex.Unlock()
				result.Observe(int64(len(labelsSet)))
			})
		labelsSet = make(map[labelKey]bool, 0)
	})
}
